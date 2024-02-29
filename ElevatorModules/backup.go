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

var backupHallRequests = make([][]int, io.NumFloors)
var backupCabRequestMap = make(map[int][io.NumFloors]int)
var quitJobAsBackup = make(chan bool)

func DebugBackupMaps() {
	for key, value := range backupCabRequestMap {
		fmt.Println(fmt.Sprint(key) + ":" + fmt.Sprint(value[0]) + "," + fmt.Sprint(value[1]) + "," + fmt.Sprint(value[2]) + "," + fmt.Sprint(value[3]))
	}
}

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

func InitBackup() {
	for i := 0; i < io.NumFloors; i++ {
		backupHallRequests[i] = make([]int, io.NumButtons-1)
		for j := 0; j < io.NumButtons-1; j++ {
			backupHallRequests[i][j] = 0
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
	//defer listener.Close()
	conn, err := listener.AcceptTCP()
	if err != nil {
		panic(err)
	}
	InitBackup()
	fmt.Println("Became backup")
	//fmt.Printf("Connected %d", conn.RemoteAddr())
	go BackupAliveTCP(addr, conn)
	go PrimaryAliveListener(conn, listener)
	return conn, addr, listener
}

func PrimaryAliveListener(conn *net.TCPConn, listener *net.TCPListener) { //nytt navn på denne nå?
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
				//fmt.Println("Message recieved: " + string(buf[:n]))
				btn, err := strconv.Atoi(recieved_message[3])
				if err != nil {
					panic(err)
				}
				flr, _ := strconv.Atoi(recieved_message[2])
				elevatorID, _ := strconv.Atoi(recieved_message[1])
				if recieved_message[3] == "2" {
					fmt.Println("Message recieved cab request: " + raw_recieved_message[i])
					UpdateBackupCabRequests(elevatorID, flr)
				} else {
					fmt.Println("Message recieved hall request: " + raw_recieved_message[i])
					UpdateBackupHallRequests(btn, flr)
				}
				//fmt.Println("Message recieved: " + string(buf[:n]))
				continue
			}
		}
	}
}

func UpdateBackupHallRequests(btnType int, flr int) {
	backupHallRequests[flr][btnType] = 1
}

func UpdateBackupCabRequests(elevatorID int, flr int) {
	_, hasKey := backupCabRequestMap[elevatorID]
	if hasKey {
		backupCabRequests := backupCabRequestMap[elevatorID]
		backupCabRequests[flr] = 1
		backupCabRequestMap[elevatorID] = backupCabRequests
	} else {
		backupCabRequests := [io.NumFloors]int{}
		for i := 0; i < io.NumFloors; i++ {
			if i == flr {
				backupCabRequests[i] = 1
			} else {
				backupCabRequests[i] = 0
			}
		}
		backupCabRequestMap[elevatorID] = backupCabRequests
	}
}

func BackupTakeover(conn *net.TCPConn) {
	fmt.Println("Before init:")
	DebugBackupMaps()
	fmt.Println(backupHallRequests)
	InitPrimary()
	fmt.Println("After init:")
	DebugBackupMaps()
	fmt.Println(backupHallRequests)
	quitJobAsBackup <- true
	elevator.ElevatorType = el.Primary
	for k, v := range backupCabRequestMap {
		cabRequestMap[k] = v
	}
	for i := range backupHallRequests {
		_ = copy(hallRequests[i], backupHallRequests[i])
	}
	fmt.Println("After copying init:")
	DebugMaps()
	fmt.Println(hallRequests)
}

func BackupAliveTCP(addr *net.TCPAddr, conn *net.TCPConn) {
	for {
		//fmt.Println("Sending Backup Alive")
		select {
		case <-quitJobAsBackup:
			return
		default:
			_, err := conn.Write(append([]byte("Backup alive"), 0))
			if err != nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// func PrimaryAliveListener() {
// 	fmt.Println("Primary alive listener maybe started")
// 	rand.Seed(time.Now().UnixNano())
// 	delay := rand.Intn(2500)
// 	delay += 500
// 	addr, err := net.ResolveUDPAddr("udp4", "localhost:29502")
// 	if err != nil {
// 		fmt.Println("Failed to resolve, backup alive listener")
// 	}
// 	conn, err := net.ListenUDP("udp4", addr)
// 	if err != nil {
// 		fmt.Println("Failed to listen, backup alive listener")
// 	}
// 	defer conn.Close()

// 	err = conn.SetReadDeadline(time.Now().Add(time.Duration(delay) * time.Millisecond))
// 	if err != nil {
// 		fmt.Println("Error setting read deadline:", err)
// 	}
// 	fmt.Println("Primary alive listener started")
// 	for {
// 		buffer := make([]byte, 1024)
// 		_, _, err := conn.ReadFromUDP(buffer)
// 		if err != nil {
// 			fmt.Println("Failed to read, backup alive listener")
// 		} else {
// 			fmt.Println("Primary alive")
// 		}
// 	}
// }
