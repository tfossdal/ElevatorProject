package ElevatorModules

import (
	el "ElevatorProject/Elevator"
	io "ElevatorProject/elevio"
	"fmt"
	"net"
	"strconv"
	"strings"
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

func InitBackupHallRequests() {
	for i := 0; i < io.NumFloors; i++ {
		hallRequests[i] = make([]int, io.NumButtons-1)
		for j := 0; j < io.NumButtons-1; j++ {
			hallRequests[i][j] = 0
		}
	}
}

func AcceptPrimaryDial() (*net.TCPConn, *net.TCPAddr, *net.TCPListener) {
	addr, err := net.ResolveTCPAddr("tcp", ":29506")
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	conn, err := listener.AcceptTCP()
	if err != nil {
		panic(err)
	}
	InitBackupHallRequests()
	fmt.Println("Became backup")
	go BackupAliveTCP(addr, conn)
	go BackupRecieveFromPrimary(conn, listener)
	return conn, addr, listener
}

func SendAck(message string, conn *net.TCPConn) {
	fmt.Println(message)
	_, err := conn.Write([]byte(message))
	fmt.Println("Sent ack to primary")
	if err != nil {
		fmt.Println(err)
	}
}

func BackupRecieveFromPrimary(conn *net.TCPConn, listener *net.TCPListener) {
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Primary died, taking over")
			conn.Close()
			listener.Close()
			BackupTakeover(conn)
			return
		}
		raw_recieved_message := strings.Split(string(buf[:n]), ":")
		for i := range raw_recieved_message {
			if raw_recieved_message[i] == "" {
				continue
			}
			recieved_message := strings.Split(raw_recieved_message[i], ",")

			if recieved_message[0] == "n" {
				go SendAck(string(buf[:n]), conn)
				btn, err := strconv.Atoi(recieved_message[3])
				if err != nil {
					panic(err)
				}
				flr, _ := strconv.Atoi(recieved_message[2])
				elevatorID, _ := strconv.Atoi(recieved_message[1])
				if recieved_message[3] == "2" {
					UpdateCabRequests(elevatorID, flr, 1)
				} else {
					UpdateHallRequests(btn, flr, 1)
				}
				continue
			}
			if recieved_message[0] == "c" {
				go SendAck(string(buf[:n]), conn)
				btn, err := strconv.Atoi(recieved_message[3])
				if err != nil {
					panic(err)
				}
				flr, _ := strconv.Atoi(recieved_message[2])
				elevatorID, _ := strconv.Atoi(recieved_message[1])
				if recieved_message[3] == "2" {
					UpdateCabRequests(elevatorID, flr, 0)
				} else {
					UpdateHallRequests(btn, flr, 0)
				}
				continue
			}
		}
	}
}

func BackupTakeover(conn *net.TCPConn) {
	elevator.ElevatorType = el.Primary
	InitPrimary()
}

func BackupAliveTCP(addr *net.TCPAddr, conn *net.TCPConn) {
	for {
		_, err := conn.Write(append([]byte("Backup alive"), 0))
		if err != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
