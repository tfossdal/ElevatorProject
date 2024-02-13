package ElevatorModules

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func BackupAlive() {
	addr, err := net.ResolveUDPAddr("udp4", "localhost:29502")
	if err != nil {
		fmt.Println("Failed to resolve, backup alive")
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial, backup alive")
	}
	defer conn.Close()
	for {
		conn.Write([]byte("Backup alive"))
		time.Sleep(100 * time.Millisecond)
	}
}

func AcceptPrimaryDial() (*net.TCPConn, *net.TCPAddr) {
	addr, err := net.ResolveTCPAddr("tcp", ":33546")
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("Backup server established")
	conn, err := listener.AcceptTCP()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Connected %d", conn.RemoteAddr())
	return conn, addr //DONT KNOW IF ADDR IS CORRECT
}

func BackupAliveTCP(addr *net.TCPAddr, conn *net.TCPConn) {
	for {
		conn.Write(append([]byte("Backup alive"), 0))
		time.Sleep(10 * time.Millisecond)
	}
}

func PrimaryAliveListener() {
	fmt.Println("Primary alive listener maybe started")
	rand.Seed(time.Now().UnixNano())
	delay := rand.Intn(2500)
	delay += 500
	addr, err := net.ResolveUDPAddr("udp4", "localhost:29502")
	if err != nil {
		fmt.Println("Failed to resolve, backup alive listener")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen, backup alive listener")
	}
	defer conn.Close()

	err = conn.SetReadDeadline(time.Now().Add(time.Duration(delay) * time.Millisecond))
	if err != nil {
		fmt.Println("Error setting read deadline:", err)
	}
	fmt.Println("Primary alive listener started")
	for {
		buffer := make([]byte, 1024)
		_, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Failed to read, backup alive listener")
		} else {
			fmt.Println("Primary alive")
		}
	}
}
