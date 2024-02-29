package ElevatorModules

import (
	io "ElevatorProject/elevio"
	"fmt"
	el "ElevatorProject/Elevator"
)

var elevator el.Elevator = el.Elevator{Floor: -1,Dirn: io.MD_Stop,Requests: [io.NumFloors][io.NumButtons]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, State: el.Idle, ElevatorType: el.None,Config: el.Config{ClearRequestVariant: el.CV_ALL,DoorOpenDuration_s: 3.0}}

func PrintState() {
	fmt.Println(el.StateToString(elevator.State))
	fmt.Println("Direction: ", elevator.Dirn)
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

func InitLights() {
	io.SetDoorOpenLamp(false)
	SetAllLights(elevator)
}

func Fsm_onInitBetweenFloors() {
	io.SetMotorDirection(io.MD_Down)
	elevator.Dirn = io.MD_Down
	elevator.State = el.Moving
}

func Fsm_OnRequestButtonPress(btn_Floor int, btn_type io.ButtonType) {
	switch elevator.State {
	case el.DoorOpen:
		if Requests_ShouldClearImmediately(elevator, btn_Floor, btn_type) != 0 {
			Timer_start(elevator.Config.DoorOpenDuration_s)
		} else {
			if btn_type == 2 {
				//Btn_type += elevatornumber
			} else {
				//Update master matrix
			}
			elevator.Requests[btn_Floor][btn_type] = 1
		}
	case el.Moving:
		elevator.Requests[btn_Floor][btn_type] = 1
	case el.Idle:
		elevator.Requests[btn_Floor][btn_type] = 1
		var pair DirnBehaviourPair = Requests_chooseDirection(elevator)
		elevator.Dirn = pair.dirn
		elevator.State = pair.state
		switch pair.state {
		case el.DoorOpen:
			io.SetDoorOpenLamp(true)
			Timer_start(elevator.Config.DoorOpenDuration_s)
			elevator = Requests_clearAtCurrentFloor(elevator)
		case el.Moving:
			io.SetMotorDirection(elevator.Dirn)
		case el.Idle:
			break
		}
	}
	SetAllLights(elevator)
}

func Fsm_OnFloorArrival(newFloor int) {
	elevator.Floor = newFloor
	io.SetFloorIndicator(elevator.Floor)

	switch elevator.State {
	case el.Moving:
		if Requests_shouldStop(elevator) != 0 {
			io.SetMotorDirection(io.MD_Stop)
			io.SetDoorOpenLamp(true)
			elevator = Requests_clearAtCurrentFloor(elevator)
			Timer_start(elevator.Config.DoorOpenDuration_s)
			SetAllLights(elevator)
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
	switch elevator.State {
	case el.DoorOpen:
		var pair DirnBehaviourPair = Requests_chooseDirection(elevator)
		elevator.Dirn = pair.dirn
		elevator.State = pair.state

		switch elevator.State {
		case el.DoorOpen:
			Timer_start(elevator.Config.DoorOpenDuration_s)
			elevator = Requests_clearAtCurrentFloor(elevator)
			SetAllLights(elevator)
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
