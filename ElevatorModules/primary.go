package ElevatorModules

import (
	el "ElevatorProject/Elevator"
	pm "ElevatorProject/PrimaryModules"
	io "ElevatorProject/elevio"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var hallRequests = make([][]int, io.NumFloors)
var cabRequestMap = make(map[int][io.NumFloors]int) //Format: {ID: {floor 0, ..., floor N-1}}
var elevatorStatesMap = make(map[int][3]int)        //Format: {ID: {state, direction, floor}}

var hallRequestMtx = sync.Mutex{}
var cabRequestMtx = sync.Mutex{}
var elevatorStatesMtx = sync.Mutex{}

var ConnectedToBackup = false
var backupTimeoutTime = 3

var elevatorLives = make(chan int, 30)
var checkLiving = make(chan int)
var requestLiving = make(chan int, 5)
var idOfLivingElev = make(chan int, 5)
var printList = make(chan int)
var newOrderCh = make(chan [3]int, 30)
var clearOrderCh = make(chan [3]int, 100)
var newStatesCh = make(chan [4]int, 30)
var retrieveElevatorStates = make(chan int)
var elevatorStates = make(chan map[int][3]int)
var terminateBackupConnection = make(chan int, 10)
var newlyAliveID = make(chan int)
var transmittedCabOrderCh = make(chan [2]int, 4)
var listOfLivingCh = make(chan map[int]time.Time, 10)
var reassignCh = make(chan int, 5)
var ackedMessageCh = make(chan string, 5)
var otherPrimaryID = 0

type HRAElevState struct {
	Behavior    string             `json:"behaviour"`
	Floor       int                `json:"floor"`
	Direction   string             `json:"direction"`
	CabRequests [io.NumFloors]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func InitPrimaryMatrix() {
	hallRequestMtx.Lock()
	for i := 0; i < io.NumFloors; i++ {
		hallRequests[i] = make([]int, io.NumButtons-1)
		for j := 0; j < io.NumButtons-1; j++ {
			hallRequests[i][j] = 0
		}
	}
	hallRequestMtx.Unlock()
}

func InitPrimary() {
	go PrimaryAlive()
	go CheckGoneOffline()
	go pm.ListenUDP("29503", elevatorLives, newOrderCh, clearOrderCh, newStatesCh)
	go pm.LivingElevatorHandler(elevatorLives, checkLiving, requestLiving, idOfLivingElev, printList, newlyAliveID, listOfLivingCh)
	go FixNewElevatorLights()
	go UpdateElevatorStates()
	go DialBackup()
	go ReassignRequests()
	go TCPCabOrderListener()
	go TCPCabOrderSender()
	sendHallLightsTicker := time.NewTicker(5 * time.Second)
	go SendHallLightUpdate(sendHallLightsTicker)
	reassignOrdersPeriodicallyTicker := time.NewTicker(5 * time.Second)
	go ReassignOrdersPeriodically(reassignOrdersPeriodicallyTicker)

	for {
		time.Sleep(500 * time.Millisecond)
		UpdateListOfLivingElevators()
	}
}

func UpdateListOfLivingElevators() {
	checkLiving <- 1
}

func ReassignOrdersPeriodically(ticker *time.Ticker) {
	for {
		<-ticker.C
		reassignCh <- 1
	}
}

func DialBackup() {
	time.Sleep(1500 * time.Millisecond)
	for {
		requestLiving <- 1
		livingElevatorsMap := <-listOfLivingCh
		for id, _ := range livingElevatorsMap {
			time.Sleep(500 * time.Millisecond) //Need sleep here, or else dial will spam too much
			addr, err := net.ResolveTCPAddr("tcp", ConvertIDtoIP(id)+":29506")
			if err != nil {
				fmt.Println(err)
				continue
			}
			conn, err := net.DialTCP("tcp", nil, addr)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Connected to backup")
			ConnectedToBackup = true
			go PrimaryAliveTCP(addr, conn)
			go BackupAliveListener(conn)
			go SendOrderToBackup(conn)
			sendDataToNewBackup(conn)
			time.Sleep(1 * time.Second)
			return
		}
	}
}

func sendDataToNewBackup(conn *net.TCPConn) {
	fmt.Println("Sending data to new backup")
	hallRequestMtx.Lock()
	for i := range hallRequests {
		for j := range hallRequests[i] {
			if hallRequests[i][j] == 1 {
				order := [3]int{0, i, j}
				fmt.Println("Writing orders to backup" + fmt.Sprint(order))
				_, err := conn.Write([]byte("n," + fmt.Sprint(order[0]) + "," + fmt.Sprint(order[1]) + "," + fmt.Sprint(order[2]) + ",:"))
				fmt.Println("This is what i wrote: n," + fmt.Sprint(order[0]) + "," + fmt.Sprint(order[1]) + "," + fmt.Sprint(order[2]) + ",:")
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
	hallRequestMtx.Unlock()
	cabRequestMtx.Lock()
	for k, v := range cabRequestMap {
		for i := range v {
			if v[i] == 1 {
				order := [3]int{k, i, 2}
				fmt.Println("Writing orders to backup" + fmt.Sprint(order))
				_, err := conn.Write([]byte("n," + fmt.Sprint(order[0]) + "," + fmt.Sprint(order[1]) + "," + fmt.Sprint(order[2]) + ",:"))
				fmt.Println("This is what i wrote: n," + fmt.Sprint(order[0]) + "," + fmt.Sprint(order[1]) + "," + fmt.Sprint(order[2]) + ",:")
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
	cabRequestMtx.Unlock()
}

func PrimaryAliveTCP(addr *net.TCPAddr, conn *net.TCPConn) {
	for {
		_, err := conn.Write(append([]byte("Primary alive,:"), 0))
		if err != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func BackupAliveListener(conn *net.TCPConn) {
	for {
		err := conn.SetReadDeadline(time.Now().Add(time.Duration(backupTimeoutTime) * time.Second))
		if err != nil {
			fmt.Println("Deadline for backup alive reached")
			fmt.Println(err)
			fmt.Println("Backup Died")
			ConnectedToBackup = false
			conn.Close()
			terminateBackupConnection <- 1
			go DialBackup()
			return
		}
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if (strings.Split(string(buf[:n]), ","))[0] == "n" {
			ackedMessageCh <- string(buf[:n])
		}
		if err != nil {
			fmt.Println(err)
			fmt.Println("Backup Died")
			ConnectedToBackup = false
			conn.Close()
			terminateBackupConnection <- 1
			go DialBackup()
			return
		}
	}
}

func CheckForPrimary() {
	InitPrimaryMatrix()
	addr, err := net.ResolveUDPAddr("udp4", ":29501")
	if err != nil {
		fmt.Println("Failed to resolve")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen")
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(time.Duration(time.Duration(1000) * time.Millisecond)))
	buf := make([]byte, 1024)
	_, recievedAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		//BECOME PRIMARY
		fmt.Println("No Primary alive message recieved, Becoming primary")
		elevator.ElevatorType = el.Primary
		go InitPrimary()
		conn.Close()
		return
	}
	senderID := int(recievedAddr.IP[3])
	go RecieveCabOrders(senderID)
	fmt.Printf("Recieved message: %s\n", buf[:])
	go AcceptPrimaryDial()
}

func ListenForOtherPrimary() {
	addr, err := net.ResolveUDPAddr("udp4", ":29501")
	if err != nil {
		fmt.Println("Failed to resolve")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen")
	}
	buf := make([]byte, 1024)
	_, recievedAdress, _ := conn.ReadFromUDP(buf)
	otherPrimaryID = int(recievedAdress.IP[3])
	fmt.Println("Heard other primary")
	RestartProgramme()

}

func PrimaryAlive() {
	if PingInternet() == 0 {
		return
	}
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
		conn.Write([]byte("Primary alive"))
		time.Sleep(100 * time.Millisecond)
	}
}

func SendHallLightUpdate(ticker *time.Ticker) {
	for {
		<-ticker.C
		if PingInternet() == 0 {
			return
		}
		hallRequestMtx.Lock()
		for flr := range hallRequests {
			for btn := range hallRequests[flr] {
				SendTurnOnOffLight([3]int{255, flr, btn}, hallRequests[flr][btn])
			}
		}
		hallRequestMtx.Unlock()
	}
}

func WaitForAck(message string) bool {
	startTime := time.Now().Unix()
	for {
		select {
		case recievedAck := <-ackedMessageCh:
			return recievedAck == message
		default:
			if time.Now().Unix() > startTime+1 {
				return false
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

}

func SendOrderToBackup(conn *net.TCPConn) {
	for {
		select {
		case order := <-newOrderCh:
			fmt.Println("Sending to backup")
			reassignCh <- 1
			stringToSend := "n," + fmt.Sprint(order[0]) + "," + fmt.Sprint(order[1]) + "," + fmt.Sprint(order[2]) + ",:"
			for {
				_, err := conn.Write([]byte(stringToSend))
				if err != nil {
					fmt.Print(err)
					return
				}
				if WaitForAck(stringToSend) {
					break
				}
			}
			if order[2] == 2 {
				UpdateCabRequests(order[0], order[1], 1)
			} else {
				UpdateHallRequests(order[2], order[1], 1)
			}
			go SendTurnOnOffLight(order, 1)
		case order := <-clearOrderCh:
			stringToSend := "c," + fmt.Sprint(order[0]) + "," + fmt.Sprint(order[1]) + "," + fmt.Sprint(order[2]) + ",:"
			_, err := conn.Write([]byte(stringToSend))
			if err != nil {
				fmt.Print(err)
				return
			}
			if order[2] == 2 {
				UpdateCabRequests(order[0], order[1], 0)
			} else {
				UpdateHallRequests(order[2], order[1], 0)
			}
			go SendTurnOnOffLight(order, 0)
		case <-terminateBackupConnection:
			return
		case transmittedCabOrder := <-transmittedCabOrderCh:
			if ConnectedToBackup {
				_, err := conn.Write([]byte("n," + fmt.Sprint(transmittedCabOrder[0]) + "," + fmt.Sprint(transmittedCabOrder[1]) + "," + fmt.Sprint(2) + ",:"))
				if err != nil {
					fmt.Print(err)
				}
			}
			UpdateCabRequests(transmittedCabOrder[0], transmittedCabOrder[1], 1)
		}
	}
}

func UpdateHallRequests(btnType int, flr int, setType int) {
	hallRequestMtx.Lock()
	hallRequests[flr][btnType] = setType
	hallRequestMtx.Unlock()
}

func UpdateCabRequests(elevatorID int, flr int, setType int) {
	cabRequestMtx.Lock()
	_, hasKey := cabRequestMap[elevatorID]
	if hasKey {
		cabRequests := cabRequestMap[elevatorID]
		cabRequests[flr] = setType
		cabRequestMap[elevatorID] = cabRequests
	} else {
		cabRequests := [io.NumFloors]int{}
		for i := 0; i < io.NumFloors; i++ {
			if i == flr {
				cabRequests[i] = setType
			} else {
				cabRequests[i] = 0
			}
		}
		cabRequestMap[elevatorID] = cabRequests
	}
	cabRequestMtx.Unlock()
}

func UpdateElevatorStates() {
	for {
		select {
		case newMessage := <-newStatesCh:
			elevatorID := newMessage[0]
			states := [3]int{newMessage[1], newMessage[2], newMessage[3]}
			elevatorStatesMtx.Lock()
			elevatorStatesMap[elevatorID] = states
			elevatorStatesMtx.Unlock()
		case <-retrieveElevatorStates:
			elevatorStatesMtx.Lock()
			elevatorStates <- elevatorStatesMap
			elevatorStatesMtx.Unlock()
		}
	}
}

func FixNewElevatorLights() {
	for {
		id := <-newlyAliveID
		hallRequestMtx.Lock()
		for i := range hallRequests {
			for j := range hallRequests[i] {
				if hallRequests[i][j] == 1 {
					order := [3]int{id, i, j}
					SendTurnOnOffLight(order, 1)
				}
			}
		}
		hallRequestMtx.Unlock()
		cabRequestMtx.Lock()
		for i := range cabRequestMap[id] {
			if cabRequestMap[id][i] == 1 {
				SendTurnOnOffLight([3]int{id, i, 2}, 1)
			}
		}
		cabRequestMtx.Unlock()
	}
}

func SendTurnOnOffLight(order [3]int, turnOn int) {
	sendToIp := ConvertIDtoIP(order[0])
	if order[2] != 2 {
		sendToIp = ConvertIDtoIP(255)
	}
	addr, err := net.ResolveUDPAddr("udp4", sendToIp+":29505")
	if err != nil {
		fmt.Println("Failed to resolve, light light")
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial, light light")
	}
	defer conn.Close()
	conn.Write([]byte(strconv.Itoa(order[0]) + "," + strconv.Itoa(order[1]) + "," + strconv.Itoa(order[2]) + "," + strconv.Itoa(turnOn)))
}

func DistributeOrderMatrix(outputMatrix map[string][][2]bool) {
	for id, req := range outputMatrix {
		idInt, _ := strconv.Atoi(id)
		addr, err := net.ResolveUDPAddr("udp4", ConvertIDtoIP(idInt)+":29504")
		if err != nil {
			fmt.Println("Failed to resolve, distribute orders udp")
		}
		conn, err := net.DialUDP("udp4", nil, addr)
		if err != nil {
			fmt.Println("Failed to dial, distribute orders udp")
		}
		messageToSend := ""
		for flr := range req {
			messageToSend += fmt.Sprint(BoolToInt(req[flr][io.BT_HallUp])) + "," + fmt.Sprint(BoolToInt(req[flr][io.BT_HallDown])) + ","
		}
		conn.Write([]byte(messageToSend))
		conn.Close()
	}
}

func ReassignRequests() {

	for {
		<-reassignCh
		requestLiving <- 1
		LivingElevators := <-listOfLivingCh
		hraExecutable := ""
		switch runtime.GOOS {
		case "linux":
			hraExecutable = "hall_request_assigner"
		case "windows":
			hraExecutable = "hall_request_assigner.exe"
		default:
			panic("OS not supported")
		}

		input := HRAInput{
			HallRequests: [][2]bool{{false}, {false, false}, {false, false}, {false, false}},
			States:       map[string]HRAElevState{},
		}
		for id, _ := range LivingElevators {
			s := ""
			elevatorStatesMtx.Lock()
			if elevatorStatesMap[id][0] == 0 {
				s = "idle"
			}
			if elevatorStatesMap[id][0] == 1 {
				s = "moving"
			}
			if elevatorStatesMap[id][0] == 2 {
				s = "doorOpen"
			}
			d := ""
			if elevatorStatesMap[id][1] == -1 {
				d = "down"
			}
			if elevatorStatesMap[id][1] == 0 {
				d = "stop"
			}
			if elevatorStatesMap[id][1] == 1 {
				d = "up"
			}
			f := elevatorStatesMap[id][2]
			elevatorStatesMtx.Unlock()
			boolCabRequests := [io.NumFloors]bool{}
			cabRequestMtx.Lock()
			for i := range cabRequestMap[id] {
				if cabRequestMap[id][i] == 1 {
					boolCabRequests[i] = true
				} else {
					boolCabRequests[i] = false
				}
			}
			cabRequestMtx.Unlock()
			input.States[fmt.Sprint(id)] = HRAElevState{s, f, d, boolCabRequests}
		}
		hallRequestMtx.Lock()
		for i := range hallRequests {
			for j := range hallRequests[i] {
				if hallRequests[i][j] == 1 {
					input.HallRequests[i][j] = true
				} else {
					input.HallRequests[i][j] = false
				}
			}
		}
		hallRequestMtx.Unlock()

		jsonBytes, err := json.Marshal(input)
		if err != nil {
			fmt.Println("json.Marshal error: ", err)
			return
		}

		ret, err := exec.Command("hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
		if err != nil {
			fmt.Println("exec.Command error: ", err)
			fmt.Println(string(ret))
			return
		}

		output := new(map[string][][2]bool)
		err = json.Unmarshal(ret, &output)
		if err != nil {
			fmt.Println("json.Unmarshal error: ", err)
			return
		}
		DistributeOrderMatrix(*output)

		fmt.Printf("output: \n")
		for k, v := range *output {
			fmt.Printf("%6v :  %+v\n", k, v)
		}
	}
}

func TCPCabOrderListener() {
	addr, err := net.ResolveTCPAddr("tcp", ":29507")
	if err != nil {
		fmt.Println("Failed to resolve")
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("Failed to listen")
	}
	buf := make([]byte, 1024)
	for {
		if PingInternet() == 0 {
			return
		}
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept")

		}
		if PingInternet() == 0 {
			conn.Close()
			return
		}
		n, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			fmt.Println(err)
			fmt.Println("Failed to read, TCP cab transmit")
			break
		}
		raw_recieved_message := strings.Split(string(buf[:n]), ":")
		for i := range raw_recieved_message {
			if raw_recieved_message[i] == "" {
				break
			}
			floor, _ := strconv.Atoi(raw_recieved_message[i])
			fmt.Println("Recieved cab at floor: " + fmt.Sprint(floor))
			remoteIP := conn.RemoteAddr().(*net.TCPAddr).IP
			fmt.Println(remoteIP)
			IPString := fmt.Sprint(remoteIP)
			IpPieces := strings.Split(IPString, ".")
			id, _ := strconv.Atoi(IpPieces[3])
			fmt.Println("Recieved transmitted orders: " + fmt.Sprint([3]int{id, floor, 2}))
			transmittedCabOrderCh <- [2]int{id, floor}
			UpdateCabRequests(id, floor, 1)
		}
	}
}

func TCPCabOrderSender() {
	addr, err := net.ResolveTCPAddr("tcp", ":29508")
	if err != nil {
		fmt.Println("Failed to resolve")
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("Failed to listen")
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept")
		}
		remoteIP := conn.RemoteAddr().(*net.TCPAddr).IP
		IPString := fmt.Sprint(remoteIP)
		IpPieces := strings.Split(IPString, ".")
		id, _ := strconv.Atoi(IpPieces[3])

		var stringToSend = ""

		fmt.Println("ID: " + fmt.Sprint(id))
		cabRequestMtx.Lock()
		fmt.Println(cabRequestMap[id])
		for i := 0; i < io.NumFloors; i++ {
			if cabRequestMap[id][i] == 1 {
				stringToSend += fmt.Sprint(i) + ":"
			}
		}
		cabRequestMtx.Unlock()
		if stringToSend == "" {
			stringToSend = ":"
		}
		_, err = conn.Write([]byte(stringToSend))
		if err != nil {
			fmt.Println("Failed to write, TCP cab transmit")
		}
		conn.Close()
	}
}

func RestartProgramme() {

	TransmitCabOrders(otherPrimaryID)
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run restartSelf.go")
	cmd.Run()
	panic("Dying")
}
