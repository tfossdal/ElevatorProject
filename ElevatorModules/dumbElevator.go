package ElevatorModules

import (
	"fmt"
	"net"
	"time"
)

func IAmAlive() {
	addr, err := net.ResolveUDPAddr("udp4", "10.100.23.255:29503")
	if err != nil {
		fmt.Println("Failed to resolve, send order")
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial, send order")
	}
	defer conn.Close()
	for {
		//fmt.Println("Sending message")
		conn.Write([]byte("4"))
		//fmt.Println("Message sent: Elevator alive")
		time.Sleep(10 * time.Millisecond)
	}
}
