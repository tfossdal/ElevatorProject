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
		fmt.Println(senderIP)
	}
}

func main() {
	ListenUDP("29505")
}
