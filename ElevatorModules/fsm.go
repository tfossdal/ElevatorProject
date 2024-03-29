package ElevatorModules

import (
	el "ElevatorProject/Elevator"
	io "ElevatorProject/elevio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var elevator el.Elevator = el.Elevator{Floor: -1, Dirn: io.MD_Stop, Requests: [io.NumFloors][io.NumButtons]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, State: el.Idle, ElevatorType: el.None, Config: el.Config{ClearRequestVariant: el.CV_ALL, DoorOpenDuration_s: 3.0}}
var OrderMtx = sync.Mutex{}
var TimeStartedMoving time.Time

var IsUnableToMove = false
var UanbleToMoveMtx = sync.Mutex{}

var ackedCabMessageCh = make(chan string)

func CheckMoveAvailability() {
	for {
		if time.Since(TimeStartedMoving) > 3*io.NumFloors*time.Second && elevator.State == el.Moving {
			UanbleToMoveMtx.Lock()
			IsUnableToMove = true
			UanbleToMoveMtx.Unlock()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func AddCabRequest(btn_floor int, btn_type io.ButtonType) {
	if btn_type != io.BT_Cab {
		return
	}
	OrderMtx.Lock()
	elevator.Requests[btn_floor][io.BT_Cab] = 1
	OrderMtx.Unlock()
	Fsm_OnRequestButtonPress(-2, 0)
}

func UpdateLocalRequestMatrix(newMatrix [io.NumFloors][2]int) {
	OrderMtx.Lock()
	for flr := range newMatrix {
		elevator.Requests[flr][0] = newMatrix[flr][0]
		elevator.Requests[flr][1] = newMatrix[flr][1]
	}
	OrderMtx.Unlock()
	Fsm_OnRequestButtonPress(-2, 0)
}

func SetAllLights(es el.Elevator) {
	for floor := 0; floor < io.NumFloors; floor++ {
		for btn := 0; btn < io.NumButtons; btn++ {
			if es.Requests[floor][btn] != 0 {
				io.SetButtonLamp(io.ButtonType(btn), floor, true)
			} else {
				io.SetButtonLamp(io.ButtonType(btn), floor, false)
			}
		}
	}
}

func SetAllCabLights(es el.Elevator) {
	for floor := 0; floor < io.NumFloors; floor++ {
		if es.Requests[floor][io.BT_Cab] != 0 {
			io.SetButtonLamp(io.ButtonType(io.BT_Cab), floor, true)
		} else {
			io.SetButtonLamp(io.ButtonType(io.BT_Cab), floor, false)
		}
	}
}

func InitLights() {
	io.SetDoorOpenLamp(false)
	SetAllLights(elevator)
}

func Fsm_onInitBetweenFloors() {
	TimeStartedMoving = time.Now()
	io.SetMotorDirection(io.MD_Down)
	elevator.Dirn = io.MD_Down
	elevator.State = el.Moving
}

func Fsm_OnRequestButtonPress(btn_Floor int, btn_type io.ButtonType) {
	OrderMtx.Lock()
	defer OrderMtx.Unlock()
	online := btn_Floor == -2
	switch elevator.State {
	case el.DoorOpen:
		if online {
			Requests_ClearImmediately_Online(elevator)
		} else {
			if Requests_ShouldClearImmediately(elevator, btn_Floor, btn_type) != 0 {
				Timer_start(elevator.Config.DoorOpenDuration_s)
			} else {
				elevator.Requests[btn_Floor][btn_type] = 1
			}
		}
	case el.Moving:
		if !online {
			elevator.Requests[btn_Floor][btn_type] = 1
		}
	case el.Idle:
		if !online {
			elevator.Requests[btn_Floor][btn_type] = 1
		}
		OrderMtx.Unlock()
		var pair DirnBehaviourPair = Requests_chooseDirection(elevator)
		OrderMtx.Lock()
		elevator.Dirn = pair.dirn
		elevator.State = pair.state
		switch pair.state {
		case el.DoorOpen:
			io.SetDoorOpenLamp(true)
			Timer_start(elevator.Config.DoorOpenDuration_s)
			OrderMtx.Unlock()
			elevator = Requests_clearAtCurrentFloor(elevator)
			OrderMtx.Lock()
		case el.Moving:
			TimeStartedMoving = time.Now()
			io.SetMotorDirection(elevator.Dirn)
		case el.Idle:
			break
		}
	}
	SetAllCabLights(elevator)
}

func Fsm_OnFloorArrival(newFloor int) {
	elevator.Floor = newFloor
	io.SetFloorIndicator(elevator.Floor)
	UanbleToMoveMtx.Lock()
	IsUnableToMove = false
	UanbleToMoveMtx.Unlock()

	switch elevator.State {
	case el.Moving:
		if Requests_shouldStop(elevator) != 0 {
			io.SetMotorDirection(io.MD_Stop)
			io.SetDoorOpenLamp(true)
			elevator = Requests_clearAtCurrentFloor(elevator)
			Timer_start(elevator.Config.DoorOpenDuration_s)
			SetAllCabLights(elevator)
			elevator.State = el.DoorOpen
		}
	default:
		break
	}
}

func Fsm_OnDoorTimeout() {
	switch elevator.State {
	case el.DoorOpen:
		var pair DirnBehaviourPair = Requests_chooseDirection(elevator)
		elevator.Dirn = pair.dirn
		elevator.State = pair.state

		switch elevator.State {
		case el.DoorOpen:
			Timer_start(elevator.Config.DoorOpenDuration_s)
			elevator = Requests_clearAtCurrentFloor(elevator)
			SetAllCabLights(elevator)
		case el.Idle:
			io.SetDoorOpenLamp(false)
			io.SetMotorDirection(elevator.Dirn)
		case el.Moving:
			io.SetDoorOpenLamp(false)
			TimeStartedMoving = time.Now()
			io.SetMotorDirection(elevator.Dirn)
		}
	default:
		break
	}
}

func WaitForCabAck(message string) bool {
	startTime := time.Now().Unix()
	for {
		select {
		case recievedAck := <-ackedCabMessageCh:
			return recievedAck == message
		default:
			if time.Now().Unix() > startTime+1 {
				return false
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func TransmitCabOrders(primaryID int) {
	addr, err := net.ResolveTCPAddr("tcp", ConvertIDtoIP(primaryID)+":29507")
	if err != nil {
		fmt.Println("Failed to resolve, send order")
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial, send order")
	}
	OrderMtx.Lock()
	defer OrderMtx.Unlock()
	stringToSend := ""
	for i := 0; i < io.NumFloors; i++ {
		if elevator.Requests[i][2] == 1 {
			stringToSend += fmt.Sprint(i) + ":"
		}
	}
	if stringToSend == "" {
		stringToSend = ":"
		conn.Close()
		return
	}
	go RecieveAck(conn)
	for {
		_, err = conn.Write([]byte(stringToSend))
		if err != nil {
			fmt.Println(err)
		}
		if WaitForCabAck(stringToSend) {
			break
		}
	}
	conn.Close()
}

func RecieveAck(conn *net.TCPConn) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Failed to read, TCP cab recieve")
	}
	ackedCabMessageCh <- string(buf[:n])
}

func sendCabRetransmittAck(ackMessage string, conn *net.TCPConn) {
	conn.Write([]byte(ackMessage))
}

func RecieveCabOrders(primaryID int) {
	var conn = &net.TCPConn{}
	for {
		addr, err := net.ResolveTCPAddr("tcp", ConvertIDtoIP(primaryID)+":29508")
		if err != nil {
			fmt.Println("Failed to resolve, recieve cab order")
			continue
		}
		conn, err = net.DialTCP("tcp", nil, addr)
		if err != nil {
			fmt.Println("Failed to dial, recieve cab order")
			continue
		}
		break
	}
	OrderMtx.Lock()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	go sendCabRetransmittAck(string(buf[:n]), conn)
	if err != nil {
		fmt.Println("Failed to read, TCP cab recieve")
	}
	raw_recieved_message := strings.Split(string(buf[:n]), ":")
	for i := range raw_recieved_message {
		if raw_recieved_message[i] == "" {
			break
		}
		floor, _ := strconv.Atoi(raw_recieved_message[i])
		elevator.Requests[floor][io.BT_Cab] = 1
	}
	SetAllCabLights(elevator)
	OrderMtx.Unlock()
	time.Sleep(100 * time.Millisecond)
	Fsm_OnRequestButtonPress(-2, 0)
	conn.Close()
}

func Fsm_Obstructed() {
	for {
		if IsObstructed && elevator.State == el.DoorOpen {
			Timer_start(elevator.Config.DoorOpenDuration_s)
		}
	}
}
