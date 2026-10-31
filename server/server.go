// custom tcp/ip stack
package server

import (
	"fmt"
	"net"
	"syscall"
	"golang.org/x/sys/unix"
)

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

		fmt.Printf("Message: %q\n", string(buffer[ihl:n]))
	}
}