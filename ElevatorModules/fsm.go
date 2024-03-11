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

func PrintState() {
	fmt.Println(el.StateToString(elevator.State))
	fmt.Println("Direction: ", elevator.Dirn)
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

func debugRequestMatrix() {
	fmt.Println(elevator.Requests)
}

func UpdateLocalRequestMatrix(newMatrix [io.NumFloors][2]int) {
	OrderMtx.Lock()
	for flr := range newMatrix {
		elevator.Requests[flr][0] = newMatrix[flr][0]
		elevator.Requests[flr][1] = newMatrix[flr][1]
	}
	OrderMtx.Unlock()
	debugRequestMatrix()
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
	io.SetMotorDirection(io.MD_Down)
	elevator.Dirn = io.MD_Down
	elevator.State = el.Moving
}

// func Fsm_OnNewOrder(btn_floor int, btn_type io.ButtonType) {
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
			io.SetMotorDirection(elevator.Dirn)
		case el.Idle:
			break
		}
	}
	SetAllCabLights(elevator)
	// if !online {
	// 	SetAllLights(elevator)
	// }
}

func Fsm_OnFloorArrival(newFloor int) {
	elevator.Floor = newFloor
	io.SetFloorIndicator(elevator.Floor)

	switch elevator.State {
	case el.Moving:
		fmt.Println("test1")
		if Requests_shouldStop(elevator) != 0 {
			fmt.Println("test2")
			io.SetMotorDirection(io.MD_Stop)
			fmt.Println("HELLO1")
			io.SetDoorOpenLamp(true)
			fmt.Println("HELLO2")
			elevator = Requests_clearAtCurrentFloor(elevator)
			fmt.Println("HELLO3")
			Timer_start(elevator.Config.DoorOpenDuration_s)
			// if !ConnectedToBackup {
			// 	SetAllLights(elevator)
			// }
			SetAllCabLights(elevator)
			elevator.State = el.DoorOpen
		}
	default:
		break
	}
}

// func Fsm_OnStopButtonpress() {
// 	switch elevator.state {
// 	case Moving:
// 		elevator.dirn = io.MD_Stop
// 		elevator.state = Idle
// 	case Idle:
// 	}
// }

func Fsm_OnDoorTimeout() {
	fmt.Println("Door timed out")
	switch elevator.State {
	case el.DoorOpen:
		var pair DirnBehaviourPair = Requests_chooseDirection(elevator)
		elevator.Dirn = pair.dirn
		elevator.State = pair.state

		switch elevator.State {
		case el.DoorOpen:
			Timer_start(elevator.Config.DoorOpenDuration_s)
			elevator = Requests_clearAtCurrentFloor(elevator)
			// if !ConnectedToBackup {
			// 	SetAllLights(elevator)
			// }
			SetAllCabLights(elevator)
		case el.Idle:
			io.SetDoorOpenLamp(false)
			io.SetMotorDirection(elevator.Dirn)
		case el.Moving:
			io.SetDoorOpenLamp(false)
			io.SetMotorDirection(elevator.Dirn)
		}
	default:
		break
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
		conn.Close()
		return
	}
	//defer conn.Close()
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
	}
	fmt.Println("Transmitting: " + stringToSend)
	_, err = conn.Write([]byte(stringToSend))
	if err != nil {
		fmt.Println(err)
	}
	conn.Close()
}

func RecieveCabOrders(primaryID int) {
	for {
		var conn = &net.TCPConn{}
		for {
			fmt.Println("LOOKING FOR PRIMARY TO SEND ORDERS")
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
		if err != nil {
			fmt.Println("Failed to read, TCP cab recieve")
		}
		fmt.Println("Recieved stuff: " + string(buf[:n]))
		raw_recieved_message := strings.Split(string(buf[:n]), ":")
		for i := range raw_recieved_message {
			if raw_recieved_message[i] == "" {
				break
			}
			floor, _ := strconv.Atoi(raw_recieved_message[i])
			fmt.Println("Recieved cab at floor: " + fmt.Sprint(floor))
			elevator.Requests[floor][io.BT_Cab] = 1
		}
		//SetAllCabLights(elevator)
		InitLights()
		OrderMtx.Unlock()
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Called FSM_OnRequestButtonPress")
		fmt.Println("Requests: " + fmt.Sprint(elevator.Requests))
		Fsm_OnRequestButtonPress(-2, 0)
		conn.Close()
	}
}

func Fsm_Obstructed(){
	for{
		if IsObstructed && elevator.State == el.DoorOpen {
			Timer_start(elevator.Config.DoorOpenDuration_s)
		}
	}
}