package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

func ipv4Checksum(buf []byte) uint16 {
    var sum uint32

    for i := 0; i+1 < len(buf); i += 2 {
        sum += uint32(buf[i])<<8 | uint32(buf[i+1])
    }

    if len(buf)%2 != 0 {
        sum += uint32(buf[len(buf)-1]) << 8
    }

    for sum>>16 != 0 {
        sum = (sum & 0xffff) + (sum >> 16)
    }

    return ^uint16(sum)
}

func buildIPv4Packet(dst [4]byte, payload []byte) []byte {
    packet := make([]byte, 20+len(payload))

    packet[0] = 0x45 
    packet[2] = byte((20 + len(payload)) >> 8)
    packet[3] = byte(20 + len(payload))
    packet[8] = 64
    packet[9] = 255 

    copy(packet[12:16], []byte{10, 0, 0, 16}) 
    copy(packet[16:20], dst[:])                

    checksum := ipv4Checksum(packet[:20])
    packet[10] = byte(checksum >> 8)
    packet[11] = byte(checksum)

    copy(packet[20:], payload)
    return packet
}

func Conn() {
    fd, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_RAW)
    if err != nil {
        fmt.Printf("[ERR] socket: %s\n", err)
        return
    }
    defer unix.Close(fd)

    if err := unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_HDRINCL, 1); err != nil {
        fmt.Printf("[ERR] setsockopt: %s\n", err)
        return
    }

    reader := bufio.NewReader(os.Stdin)

    for {
        fmt.Print("Message> ")
        msg, err := reader.ReadString('\n')
        if err != nil {
            fmt.Printf("[ERR] read stdin: %s\n", err)
            return
        }

        msg = strings.TrimRight(msg, "\r\n")
        if msg == "" {
            continue
        }

        dst := [4]byte{10, 0, 0, 10} // server IP
        packet := buildIPv4Packet(dst, []byte(msg))

        if err := unix.Sendto(fd, packet, 0, &unix.SockaddrInet4{
            Addr: dst,
        }); err != nil {
            fmt.Printf("[ERR] sendto: %s\n", err)
            return
        }

        fmt.Printf("Sent: %s\n", msg)
    }
}