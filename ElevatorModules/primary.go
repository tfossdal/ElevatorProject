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
	"time"
)

var hallRequests = make([][]int, io.NumFloors)
var cabRequestMap = make(map[int][io.NumFloors]int) //Format: {ID: {floor 0, ..., floor N-1}}
var elevatorStatesMap = make(map[int][3]int)        //Format: {ID: {state, direction, floor}}

// var elevatorAddresses = []string{"10.100.23.28", "10.100.23.34"}
var backupTimeoutTime = 3

var elevatorLives = make(chan int, 5)
var checkLiving = make(chan int)
var requestId = make(chan int, 5)
var idOfLivingElev = make(chan int, 5)
var printList = make(chan int)
var numberOfElevators = make(chan int, 5)
var newOrderCh = make(chan [3]int)
var newStatesCh = make(chan [4]int, 30)
var retrieveElevatorStates = make(chan int)
var elevatorStates = make(chan map[int][3]int)
var orderTransferCh = make(chan [3]int)
var terminateBackupConnection = make(chan int)
var newlyAliveID = make(chan int)
var listOfLivingCh = make(chan int)
var reassignCh = make(chan int, 5)

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

	time.Sleep(10 * time.Second)
}

func PrimaryRoutine() {
	go PrimaryAlive()
	go pm.ListenUDP("29503", elevatorLives, newOrderCh, newStatesCh)
	go pm.LivingElevatorHandler(elevatorLives, checkLiving, requestId, idOfLivingElev, printList, numberOfElevators, newlyAliveID, listOfLivingCh)
	go FixNewElevatorLights()
	go UpdateElevatorStates()
	go DialBackup()
	go ReassignRequests()

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

func DialBackup() {
	time.Sleep(1500 * time.Millisecond) //WAY TOO LONG
	printList <- 1
	num := <-numberOfElevators
	if num < 2 {
		num = 2
	}
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
		go PrimaryAliveTCP(addr, conn)
		go BackupAliveListener(conn)
		go SendOrderToBackup(conn)
		//go TransferOrdersToBackup(conn)
		sendDataToNewBackup(conn)
		break
	}
	//defer conn.Close()
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
		conn.SetReadDeadline(time.Now().Add(time.Duration(backupTimeoutTime) * time.Second))
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		//n, err := conn.Read(buf)
		//fmt.Println("Message recieved: " + string(buf[:n]))
		if err != nil {
			fmt.Println(err)
			fmt.Println("Backup Died")
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
	_, _, err = conn.ReadFromUDP(buf)
	if err != nil {
		//BECOME PRIMARY
		fmt.Println("No Primary alive message recieved, Becoming primary")
		elevator.ElevatorType = el.Primary
		InitPrimary()
		conn.Close()
		return
	}
	fmt.Printf("Recieved message: %s\n", buf[:])
	go AcceptPrimaryDial()
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
				UpdateCabRequests(order[0], order[1])
			} else {
				UpdateHallRequests(order[2], order[1])
			}
			go SendTurnOnLight(order)

		case <-terminateBackupConnection:
			return
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

func UpdateHallRequests(btnType int, flr int) {
	hallRequests[flr][btnType] = 1
}

func UpdateCabRequests(elevatorID int, flr int) {
	_, hasKey := cabRequestMap[elevatorID]
	if hasKey {
		cabRequests := cabRequestMap[elevatorID]
		cabRequests[flr] = 1
		cabRequestMap[elevatorID] = cabRequests
	} else {
		cabRequests := [io.NumFloors]int{}
		for i := 0; i < io.NumFloors; i++ {
			if i == flr {
				cabRequests[i] = 1
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
					SendTurnOnLight(order)
				}
			}
		}
		for i := range cabRequestMap[id] {
			if cabRequestMap[id][i] == 1 {
				SendTurnOnLight([3]int{id, i, 2})
			}
		}
	}
}

func SendTurnOnLight(order [3]int) {
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
	conn.Write([]byte(strconv.Itoa(order[0]) + "," + strconv.Itoa(order[1]) + "," + strconv.Itoa(order[2])))
}

func ReassignRequests() {
	for {
		<-reassignCh
		fmt.Println("3")
		requestId <- 3
		fmt.Println("4")
		θ := <-numberOfElevators
		fmt.Println("5")
		LivingElevators := make([]int, θ)
		for i := 1; i <= θ; i++ {
			fmt.Println("6")
			LivingElevators[i-1] = <-listOfLivingCh
		}
		fmt.Println("7")
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
		for i := range LivingElevators {
			id := LivingElevators[i]
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

		fmt.Printf("output: \n")
		for k, v := range *output {
			fmt.Printf("%6v :  %+v\n", k, v)
		}
	}
}
