// custom tcp/ip stack
package server

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)
 
type IPV4_H struct {
	Version     byte
	IHL         uint8
	TOS         uint8
	TotalLength [2]byte
	Indent      [2]byte
	Flags       uint8
	FragOffset  uint16
	TTL         uint8
	Protocol    uint8
	HeaderCS    [2]byte
	SourceAddr  [4]byte
	DestAddr    [4]byte
	Options     []byte
}

type IP4 struct {
	Header  *IPV4_H
	Payload []byte
}


func ParseIpv4(buffer *[]byte) (IP4, error) {
	ipv4 := IP4{}
	ipv4.Header = &IPV4_H{}

	ipv4.Header.Version = (*buffer)[0] >> 4
	ipv4.Header.IHL = (*buffer)[0] & 0x0F

	ipv4.Header.TOS = (*buffer)[1]
	copy(ipv4.Header.TotalLength[:], (*buffer)[2:4])

	copy(ipv4.Header.Indent[:], (*buffer)[4:6])

	ipv4.Header.Flags = (*buffer)[6] >> 5
	flagsAndFragOffset := binary.BigEndian.Uint16((*buffer)[6:8])
	ipv4.Header.FragOffset = flagsAndFragOffset & 0x1FFF

	ipv4.Header.TTL = (*buffer)[8]
	ipv4.Header.Protocol = (*buffer)[9]

	copy(ipv4.Header.HeaderCS[:], (*buffer)[10:12])
	copy(ipv4.Header.SourceAddr[:], (*buffer)[12:16])
	copy(ipv4.Header.DestAddr[:], (*buffer)[16:20])

	if ipv4.Header.IHL == 5 {
		ipv4.Payload = (*buffer)[20:]
	} else if ipv4.Header.IHL > 5 {
		ipv4.Header.Options = (*buffer)[20 : ipv4.Header.IHL*4]
		ipv4.Payload = (*buffer)[ipv4.Header.IHL*4:]
	} else {
		return IP4{}, fmt.Errorf("IHL header size not valid: %d\n", ipv4.Header.IHL)
	}
	return ipv4, nil
}

func IpToString(ip4 *IP4) (string) {
	source_addr := ip4.Header.SourceAddr
	return fmt.Sprintf("%d.%d.%d.%d", source_addr[0], source_addr[1], source_addr[2], source_addr[3])
}


func Server() {
	fd, err := unix.Open("/dev/net/tun", syscall.O_RDWR, 0)
	if err != nil {
		fmt.Printf("[ERR] open /dev/net/tun: %s\n", err)
		return
	}
	defer unix.Close(fd)

	ifr, err := unix.NewIfreq("tun0")
	if err != nil {
		fmt.Printf("[ERR] NewIfreq: %s\n", err)
		return
	}

	ifr.SetUint16(unix.IFF_TUN | unix.IFF_NO_PI)

	if err := unix.IoctlIfreq(fd, unix.TUNSETIFF, ifr); err != nil {
		fmt.Printf("[ERR] TUNSETIFF: %s\n", err)
		return
	}


	sock, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		fmt.Printf("[ERR] socket: %s\n", err)
		return
	}
	defer unix.Close(sock)

	ifr.SetInet4Addr([]byte{0x0A, 0x00, 0x00, 0x10})
	if err := unix.IoctlIfreq(sock, unix.SIOCSIFADDR, ifr); err != nil {
		fmt.Printf("[ERR] SIOCSIFADDR: %s\n", err)
		return
	}

	ip, err := ifr.Inet4Addr()
	if err != nil {
		fmt.Printf("[ERR] Inet4Addr: %s\n", err)
		return
	}

	fmt.Printf("[INFO] Server running on: %s\n", net.IP(ip[:]).String())

	ifr.SetInet4Addr([]byte{255,255,255,0})

	if err := unix.IoctlIfreq(sock, unix.SIOCSIFNETMASK, ifr); err !=nil {
		fmt.Printf("[ERR] SIOCSIFNETMASK: %s\n", err)
    	return
	}

	if err := unix.IoctlIfreq(sock, unix.SIOCGIFFLAGS, ifr); err != nil {
		fmt.Printf("[ERR] SIOCGIFFLAGS: %s\n", err)
		return
	}

	if ifr.Uint16()&unix.IFF_UP == 0 {
		fmt.Printf("[INFO] Bringing %s up\n", ifr.Name())
		ifr.SetUint16(ifr.Uint16() | unix.IFF_UP)
		if err := unix.IoctlIfreq(sock, unix.SIOCSIFFLAGS, ifr); err != nil {
			fmt.Printf("[ERR] SIOCSIFFLAGS: %s\n", err)
			return
		}
	}


	buffer := make([]byte, 65535)
	for {
		n, err := unix.Read(fd, buffer)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			fmt.Printf("[ERR] tun read: %s\n", err)
			return
		}

		if n < 20 || buffer[0]>>4 != 4 || buffer[9] != 253 {
			continue
		}

		ihl := int(buffer[0]&0x0f) * 4
		if ihl < 20 || n <= ihl {
			continue
		}
		client, err := ParseIpv4(&buffer)
		if err!=nil{
			fmt.Printf("[ERR] client parsing err: %s\n", err)
		}
		
		fmt.Printf("->[%s]: %q\n", IpToString(&client),string(buffer[ihl:n]))
	}
}