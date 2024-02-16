package main

import (
	"fmt"
	"net"
)

func ListenUDP(port string) {
	addr, err := net.ResolveUDPAddr("udp4", ":"+port)
	if err != nil {
		fmt.Println("Failed to resolve")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen")
	}
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, recievedAddr, err := conn.ReadFromUDP(buf)
		senderIP := recievedAddr.IP
		fmt.Println("Read something")
		if err != nil {
			fmt.Println("Failed to listen")
		}
		fmt.Println(string(buf[:n]))
		SenderNumber := fmt.Sprintf("%d", senderIP[3])
		fmt.Println(senderIP.String())
		fmt.Println(SenderNumber)
	}
}

func main() {
	ListenUDP("29505")
}
