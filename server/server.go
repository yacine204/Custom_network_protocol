// custom tcp/ip stack

package server

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/unix"
)

// server on 127.0.0.1:70

const SERVER_PORT = 70 

type TCPPacket struct{ 
	SrcIp [4]byte
	DestIP [4]byte
	SrcPort uint16
	DestPort uint16
	SequenceNumber [8]byte
	Data []byte
}

type IFR struct{

}

func InitPacket() (*TCPPacket){
	tcp := &TCPPacket{}
	copy(tcp.SrcIp[:], []byte{0x7F,0x00,0x00,0x01})
	tcp.SrcPort = SERVER_PORT

	return tcp
}

func parseInComingClientsPackets(buffer *[]byte){
	
}

func Server(){
	fd, err := unix.Open("/dev/net/tun", syscall.O_RDWR, 0644)
	if err!=nil{
		fmt.Printf("[ERR] %s\n", err)
		return
	}
	defer unix.Close(fd)

	ifr, err := unix.NewIfreq("tun0")
	
	if err!=nil{
		fmt.Printf("[ERR] %s\n", err)
		return
	}

	ifr.SetUint16(unix.IFF_TUN | unix.IFF_NO_PI)

	err = unix.IoctlIfreq(fd, unix.TUNSETIFF, ifr)

	if err!=nil{
		fmt.Printf("[ERR] %s\n", err)
		return
	}

	ifr.SetInet4Addr([]byte{0x7F,0x00,0x00,0x01})

	sock, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	
	if err!=nil{
		fmt.Printf("[ERR] %s\n", err)
		return
	}

	unix.IoctlIfreq(sock, unix.SIOCSIFADDR, ifr) 

	// ip, _ := ifr.Inet4Addr()
	// fmt.Println(ifr.Name(), ip)
	
	unix.IoctlIfreq(sock, unix.SIOCGIFFLAGS, ifr)

	if ifr.Uint16() & unix.IFF_UP == 0 {
		fmt.Printf("getting %s up\n", ifr.Name())
		// add up flag for activation
		ifr.SetUint16(ifr.Uint16() | unix.IFF_UP)
		unix.IoctlIfreq(sock, unix.SIOCSIFFLAGS, ifr)
	}

	for {
		
	}
	
}