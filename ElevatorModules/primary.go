package ElevatorModules

import (
	"ElevatorProject/ElevatorModules/PrimaryModules"
	io "ElevatorProject/elevio"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

var requests = make([][]int, io.NumFloors)

// var elevatorAddresses = []string{"10.100.23.28", "10.100.23.34"}
var backupTimeoutTime = 5

var elevatorLives = make(chan int, 5)
var checkLiving = make(chan int)
var requestId = make(chan int, 5)
var idOfLivingElev = make(chan int, 5)
var printList = make(chan int)

func InitPrimary() {
	//Initialize order matrix
	for i := 0; i < io.NumFloors; i++ {
		requests[i] = make([]int, io.NumButtons)
		for j := 0; j < io.NumButtons; j++ {
			requests[i][j] = 0
		}
	}

	//Start GoRoutines
	go PrimaryRoutine()

	time.Sleep(10 * time.Second)
}

func PrimaryRoutine() {
	go PrimaryAlive()
	go PrimaryModules.ListenUDP("29503", elevatorLives)
	go PrimaryModules.LivingElevatorHandler(elevatorLives, checkLiving, requestId, idOfLivingElev, printList)

	go DialBackup() //PROBLEM This will happen before we know any addresses
	for {
		printList <- 1
		//UpdateLivingElevators()
	}
}

func UpdateLivingElevators() {
	checkLiving <- 1
}

func ConvertIDtoIP(id int) string {
	return "10.100.23." + fmt.Sprint(id)
}

func DialBackup() {
	for i := 1; i <= 3; i++ {
		requestId <- i
		fmt.Println(ConvertIDtoIP(<-idOfLivingElev) + ":29506")
		requestId <- i
		addr, err := net.ResolveTCPAddr("tcp", ConvertIDtoIP(<-idOfLivingElev)+":29506")
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("God to dial")
		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("Connected to backup")
		go PrimaryAliveTCP(addr, conn)
		go BackupAliveListener(conn)
		break
	}
	//defer conn.Close()
}

func PrimaryAliveTCP(addr *net.TCPAddr, conn *net.TCPConn) {
	for {
		_, err := conn.Write(append([]byte("Primary alive"), 0))
		if err != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func BackupAliveListener(conn *net.TCPConn) {
	for {
		conn.SetReadDeadline(time.Now().Add(time.Duration(backupTimeoutTime) * time.Second))
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		fmt.Println(string(buf[:n]))
		if err != nil {
			fmt.Println(err)
			fmt.Println("Backup Died")
			go DialBackup()
			//go DialBackup()
			return
		}
	}
}

func BecomePrimary() {
	//29501
	addr, err := net.ResolveUDPAddr("udp4", ":29501")
	if err != nil {
		fmt.Println("Failed to resolve")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen")
	}
	defer conn.Close()
	rand.Seed(time.Now().UnixNano())
	deadlineTime := rand.Intn(10000)
	conn.SetReadDeadline(time.Now().Add(time.Duration(time.Duration(deadlineTime) * time.Millisecond)))
	buf := make([]byte, 1024)
	_, _, err = conn.ReadFromUDP(buf)
	if err != nil {
		//BECOME PRIMARY
		fmt.Println("No Primary alive message recieved, Becoming primary")
		elevator.elevatorType = Primary
		InitPrimary()
		conn.Close()
		return
	}
	fmt.Printf("Recieved message: %s\n", buf[:])
	AcceptPrimaryDial()
}

func PrimaryAlive() {
	addr, err := net.ResolveUDPAddr("udp4", "10.100.23.255:29501")
	if err != nil {
		fmt.Println("Failed to resolve, primary alive")
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial, primary alive")
	}
	defer conn.Close()
	for {
		//fmt.Println("Sending alive message")
		conn.Write([]byte("Primary alive"))
		//fmt.Println("Message sent: Primary alive")
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
