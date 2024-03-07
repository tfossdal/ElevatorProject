package ElevatorModules

import (
	el "ElevatorProject/Elevator"
	pm "ElevatorProject/PrimaryModules"
	io "ElevatorProject/elevio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var hallRequests = make([][]int, io.NumFloors)
var cabRequestMap = make(map[int][io.NumFloors]int) //Format: {ID: {floor 0, ..., floor N-1}}
var elevatorStatesMap = make(map[int][3]int)        //Format: {ID: {state, direction, floor}}

var ConnectedToBackup = false

// var elevatorAddresses = []string{"10.100.23.28", "10.100.23.34"}
var backupTimeoutTime = 3

var elevatorLives = make(chan int, 30)
var checkLiving = make(chan int)
var requestId = make(chan int, 5)
var idOfLivingElev = make(chan int, 5)
var printList = make(chan int)
var newOrderCh = make(chan [3]int)
var clearOrderCh = make(chan [3]int)
var newStatesCh = make(chan [4]int, 30)
var retrieveElevatorStates = make(chan int)
var elevatorStates = make(chan map[int][3]int)
var orderTransferCh = make(chan [3]int)
var terminateBackupConnection = make(chan int)
var newlyAliveID = make(chan int)
var transmittedCabOrderCh = make(chan [2]int)

// var listOfLivingCh = make(chan int)
var listOfLivingCh = make(chan map[int]time.Time)
var reassignCh = make(chan int, 5)
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

func DebugMaps() {
	for key, value := range cabRequestMap {
		fmt.Println(fmt.Sprint(key) + ":" + fmt.Sprint(value[0]) + "," + fmt.Sprint(value[1]) + "," + fmt.Sprint(value[2]) + "," + fmt.Sprint(value[3]))
	}
}

func InitPrimaryMatrix() {
	for i := 0; i < io.NumFloors; i++ {
		hallRequests[i] = make([]int, io.NumButtons-1)
		for j := 0; j < io.NumButtons-1; j++ {
			hallRequests[i][j] = 0
		}
	}
}

func InitPrimary() {
	//Initialize order matrix
	//Start GoRoutines
	go PrimaryRoutine()

	//time.Sleep(10 * time.Second)
}

func PrimaryRoutine() {
	go PrimaryAlive()
	go CheckGoneOffline()
	go pm.ListenUDP("29503", elevatorLives, newOrderCh, clearOrderCh, newStatesCh)
	//go pm.LivingElevatorHandler(elevatorLives, checkLiving, requestId, idOfLivingElev, printList, numberOfElevators, newlyAliveID, listOfLivingCh)
	go pm.LivingElevatorHandler(elevatorLives, checkLiving, requestId, idOfLivingElev, printList, newlyAliveID, listOfLivingCh)
	go FixNewElevatorLights()
	go UpdateElevatorStates()
	go DialBackup()
	go ReassignRequests()
	go TCPCabOrderListener()
	go TCPCabOrderSender()

	for {
		time.Sleep(500 * time.Millisecond)
		UpdateLivingElevators()
	}
}

func UpdateLivingElevators() {
	checkLiving <- 1
}

func ConvertIDtoIP(id int) string {
	return "10.100.23." + fmt.Sprint(id)
}

/* func DialBackup() {
	time.Sleep(1500 * time.Millisecond) //WAY TOO LONG
	printList <- 1
	num := <-numberOfElevators
	if num < 2 {
		num = 2
	}
	for {
		for i := 1; i <= num; i++ {
			requestId <- i
			addr, err := net.ResolveTCPAddr("tcp", ConvertIDtoIP(<-idOfLivingElev)+":29506")
			if err != nil {
				fmt.Println(err)
				continue
			}
			time.Sleep(1500 * time.Millisecond)
			conn, err := net.DialTCP("tcp", nil, addr)
			if err != nil {
				fmt.Println(err)
				continue
			}
			time.Sleep(1 * time.Second)
			fmt.Println("Connected to backup")
			ConnectedToBackup = true
			go PrimaryAliveTCP(addr, conn)
			go BackupAliveListener(conn)
			go SendOrderToBackup(conn)
			//go TransferOrdersToBackup(conn)
			sendDataToNewBackup(conn)
			return
		}
		//defer conn.Close()
	}
} */

func DialBackup() {
	time.Sleep(1500 * time.Millisecond) //WAY TOO LONG
	for {
		requestId <- 1
		livingElevatorsMap := <-listOfLivingCh
		for id, _ := range livingElevatorsMap {
			addr, err := net.ResolveTCPAddr("tcp", ConvertIDtoIP(id)+":29506")
			if err != nil {
				fmt.Println(err)
				continue
			}
			time.Sleep(1500 * time.Millisecond)
			conn, err := net.DialTCP("tcp", nil, addr)
			if err != nil {
				fmt.Println(err)
				continue
			}
			time.Sleep(1 * time.Second)
			fmt.Println("Connected to backup")
			ConnectedToBackup = true
			go PrimaryAliveTCP(addr, conn)
			go BackupAliveListener(conn)
			go SendOrderToBackup(conn)
			//go TransferOrdersToBackup(conn)
			sendDataToNewBackup(conn)
			return
		}
		//defer conn.Close()
	}
}

func sendDataToNewBackup(conn *net.TCPConn) {
	fmt.Println("Sending data to new backup")
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
		_, err = conn.Read(buf)
		//n, err := conn.Read(buf)
		//fmt.Println("Message recieved: " + string(buf[:n]))
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

func BecomePrimary() {
	//29501
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
	rand.Seed(time.Now().UnixNano())
	deadlineTime := rand.Intn(10000)
	conn.SetReadDeadline(time.Now().Add(time.Duration(time.Duration(deadlineTime) * time.Millisecond)))
	buf := make([]byte, 1024)
	_, recievedAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		//BECOME PRIMARY
		fmt.Println("No Primary alive message recieved, Becoming primary")
		elevator.ElevatorType = el.Primary
		InitPrimary()
		conn.Close()
		return
	}
	senderID := int(recievedAddr.IP[3])
	go RecieveCabOrders(senderID)
	fmt.Printf("Recieved message: %s\n", buf[:])
	go AcceptPrimaryDial()
}

func ListenForOtherPrimary() {
	fmt.Println("number 1")
	addr, err := net.ResolveUDPAddr("udp4", ":29501")
	if err != nil {
		fmt.Println("Failed to resolve")
	}
	fmt.Println("number 2")
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen")
	}
	fmt.Println("number 3")
	buf := make([]byte, 1024)
	_, recievedAdress, _ := conn.ReadFromUDP(buf)
	otherPrimaryID = int(recievedAdress.IP[3])
	fmt.Println("Heard other primary")
	//DIE AND LIVE
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
		//fmt.Println("Sending alive message")
		conn.Write([]byte("Primary alive"))
		//fmt.Println("Message sent: Primary alive")
		time.Sleep(100 * time.Millisecond)
	}
}

func SendOrderToBackup(conn *net.TCPConn) {
	for {
		select {
		case order := <-newOrderCh:
			fmt.Println("1")
			reassignCh <- 1
			fmt.Println("2")
			_, err := conn.Write([]byte("n," + fmt.Sprint(order[0]) + "," + fmt.Sprint(order[1]) + "," + fmt.Sprint(order[2]) + ",:"))
			if err != nil {
				fmt.Print(err)
				return
			}
			if order[2] == 2 {
				UpdateCabRequests(order[0], order[1], 1)
			} else {
				UpdateHallRequests(order[2], order[1], 1)
			}
			go SendTurnOnOffLight(order, 1)
		case order := <-clearOrderCh:
			fmt.Println("1")
			fmt.Println("2")
			_, err := conn.Write([]byte("c," + fmt.Sprint(order[0]) + "," + fmt.Sprint(order[1]) + "," + fmt.Sprint(order[2]) + ",:"))
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

func TransferOrdersToBackup(conn *net.TCPConn) {
	for {
		order := <-orderTransferCh
		fmt.Println("Writing orders to backup" + fmt.Sprint(order))
		_, err := conn.Write(append([]byte("n,"+fmt.Sprint(order[0])+","+fmt.Sprint(order[1])+","+fmt.Sprint(order[2])+","), 0))
		fmt.Println("This is what i wrote: n," + fmt.Sprint(order[0]) + "," + fmt.Sprint(order[1]) + "," + fmt.Sprint(order[2]) + ",")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func UpdateHallRequests(btnType int, flr int, setType int) {
	hallRequests[flr][btnType] = setType
}

func UpdateCabRequests(elevatorID int, flr int, setType int) {
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
	DebugMaps()
}

func UpdateElevatorStates() {
	for {
		select {
		case newMessage := <-newStatesCh:
			elevatorID := newMessage[0]
			states := [3]int{newMessage[1], newMessage[2], newMessage[3]}

			elevatorStatesMap[elevatorID] = states //Tror det går, altså at den lager en ny key hvis det ikke finnes
			//MÅ SENDE VIDERE TIL BACKUP
		case <-retrieveElevatorStates:
			elevatorStates <- elevatorStatesMap
		}
	}
}

// func OrderListener() {
// 	//29503
// 	addr, err := net.ResolveUDPAddr("udp4", ":29503")
// 	if err != nil {
// 		fmt.Println("Could not connect")
// 	}
// 	conn, err := net.ListenUDP("udp4", addr)
// 	if err != nil {
// 		fmt.Println("Could not listen")
// 	}
// 	defer conn.Close()

// 	for {
// 		fmt.Print("Primary reading UDP ...")
// 		buf := make([]byte, 1024)
// 		n, _, err := conn.ReadFromUDP(buf)
// 		if err != nil {
// 			fmt.Println("Could not read")
// 		}
// 		orderStr := string(buf[:n])
// 		orderLst := strings.Split(orderStr, ",")
// 		floorIndex, _ := strconv.Atoi(orderLst[0])
// 		buttonIndex, _ := strconv.Atoi(orderLst[1])
// 		requests[floorIndex][buttonIndex] = 1
// 	}

// }

func FixNewElevatorLights() {
	for {
		id := <-newlyAliveID
		for i := range hallRequests {
			for j := range hallRequests[i] {
				if hallRequests[i][j] == 1 {
					order := [3]int{id, i, j}
					SendTurnOnOffLight(order, 1)
				}
			}
		}
		for i := range cabRequestMap[id] {
			if cabRequestMap[id][i] == 1 {
				SendTurnOnOffLight([3]int{id, i, 2}, 1)
			}
		}
	}
}

func SendTurnOnOffLight(order [3]int, turnOn int) {
	sendToIp := ConvertIDtoIP(order[0])
	if order[2] != 2 {
		sendToIp = ConvertIDtoIP(255)
	}
	fmt.Println("Sending light to: " + sendToIp)
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
	fmt.Println("turnOn is " + strconv.Itoa(turnOn))
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
			messageToSend += fmt.Sprint(boolToInt(req[flr][io.BT_HallUp])) + "," + fmt.Sprint(boolToInt(req[flr][io.BT_HallDown])) + ","
		}
		conn.Write([]byte(messageToSend))
		conn.Close()
	}
}

func boolToInt(n bool) int {
	if n {
		return 1
	}
	return 0
}

func ReassignRequests() {

	for {
		<-reassignCh
		requestId <- 1
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
			States:       map[string]HRAElevState{
				// "one": HRAElevState{
				// 	Behavior:    "moving",
				// 	Floor:       2,
				// 	Direction:   "up",
				// 	CabRequests: []bool{false, false, true, true},
				// },
				// "two": HRAElevState{
				// 	Behavior:    "idle",
				// 	Floor:       0,
				// 	Direction:   "stop",
				// 	CabRequests: []bool{false, false, false, false},
				// },
			},
		}
		for id, _ := range LivingElevators {
			s := ""
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
			fmt.Println("Floor: " + fmt.Sprint(f))
			f = 1
			boolCabRequests := [io.NumFloors]bool{}
			for i := range cabRequestMap[id] {
				if cabRequestMap[id][i] == 1 {
					boolCabRequests[i] = true
				} else {
					boolCabRequests[i] = false
				}
			}
			input.States[fmt.Sprint(id)] = HRAElevState{s, f, d, boolCabRequests}
		}
		for i := range hallRequests {
			for j := range hallRequests[i] {
				if hallRequests[i][j] == 1 {
					input.HallRequests[i][j] = true
				} else {
					input.HallRequests[i][j] = false
				}
			}
		}

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
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept")
		}
		for {
			n, err := conn.Read(buf)
			if err != nil {
				conn.Close()
				fmt.Println("Failed to read, TCP cab transmit")
				break
			}
			raw_recieved_message := strings.Split(string(buf[:n]), ":")
			for i := range raw_recieved_message {
				if raw_recieved_message[i] == "" {
					break
				}
				recieved_message := strings.Split(raw_recieved_message[i], ",")
				remoteIP := conn.RemoteAddr().(*net.TCPAddr).IP
				fmt.Println(remoteIP)
				btn, _ := strconv.Atoi(recieved_message[1])
				flr, _ := strconv.Atoi(recieved_message[0])
				IPString := fmt.Sprint(remoteIP)
				IpPieces := strings.Split(IPString, ".")
				id, _ := strconv.Atoi(IpPieces[3])
				fmt.Println("Recieved transmitted orders: " + fmt.Sprint([3]int{id, flr, btn}))
				transmittedCabOrderCh <- [2]int{id, flr}
				UpdateCabRequests(id, flr, 1)
			}
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
		fmt.Println(cabRequestMap[id])
		for i := 0; i < io.NumFloors; i++ {
			if cabRequestMap[id][i] == 1 {
				stringToSend += fmt.Sprint(i) + ":"
			}
		}
		if stringToSend == "" {
			stringToSend = ":"
		}
		_, err = conn.Write([]byte(stringToSend))
		if err != nil {
			fmt.Println("Failed to write, TCP cab transmit")
		}
	}
}

func RestartProgramme() {

	TransmitCabOrders(otherPrimaryID)
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run restartSelf.go")
	cmd.Run()
	panic("Dying")
}
