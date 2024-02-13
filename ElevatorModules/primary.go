package ElevatorModules

import (
	io "ElevatorProject/elevio"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

var requests = make([][]int, io.NumFloors)

func InitPrimary() {
	//Initialize order matrix
	for i := 0; i < io.NumFloors; i++ {
		requests[i] = make([]int, io.NumButtons)
		for j := 0; j < io.NumButtons; j++ {
			requests[i][j] = 0
		}
	}

	//Start GoRoutines
	go PrimaryAlive()
}

func BecomePrimary() {
	//29501
	addr, err := net.ResolveUDPAddr("udp4", "localhost:29501")
	if err != nil {
		fmt.Println("Failed to resolve")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen")
	}
	defer conn.Close()
	rand.Seed(time.Now().UnixNano())
	deadlineTime := rand.Intn(100)
	conn.SetReadDeadline(time.Now().Add(time.Duration(time.Duration(deadlineTime) * time.Millisecond)))
	buf := make([]byte, 1024)
	_, _, err = conn.ReadFromUDP(buf)
	if err != nil {
		//BECOME PRIMARY
		elevator.elevatorType = Primary
		InitPrimary()
		conn.Close()
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
		time.Sleep(10 * time.Millisecond)
	}
}

func OrderListener() {
	//29503
	addr, err := net.ResolveUDPAddr("udp4", ":29503")
	if err != nil {
		fmt.Println("Could not connect")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Could not listen")
	}
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Could not read")
		}
		orderStr := string(buf[:n])
		orderLst := strings.Split(orderStr, ",")
		floorIndex, _ := strconv.Atoi(orderLst[0])
		buttonIndex, _ := strconv.Atoi(orderLst[1])
		requests[floorIndex][buttonIndex] = 1
	}

}

