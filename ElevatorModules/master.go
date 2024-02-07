package ElevatorModules

import (
	io "ElevatorProject/elevio"
	"fmt"
	"net"
	"time"
)

func Master() {
	requests := make([][]int, io.NumFloors)

	for i := 0; i < io.NumFloors; i++ {
		requests[i] = make([]int, io.NumButtons)
		for j := 0; j < io.NumButtons; j++ {
			requests[i][j] = 0
		}
	}
}

func PrimaryAlive() {
	addr, err := net.ResolveUDPAddr("udp4", "localhost:29501")
	if err != nil {
		fmt.Println("Failed to resolve, primary alive")
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial, primary alive")
	}
	defer conn.Close()
	for {
		conn.Write([]byte("Primary alive"))
		time.Sleep(100 * time.Millisecond)
	}
}


