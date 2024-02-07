package ElevatorModules

import (
	io "ElevatorProject/elevio"
	"fmt"
)

var elevator Elevator = Elevator{-1, io.MD_Stop, [io.NumFloors][io.NumButtons]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, Idle, Config{CV_ALL, 3.0}}

func PrintState() {
	fmt.Println(StateToString(elevator.state))
	fmt.Println("Direction: ", elevator.dirn)
}

func SetAllLights(es Elevator) {
	for floor := 0; floor < io.NumFloors; floor++ {
		for btn := 0; btn < io.NumButtons; btn++ {
			if es.requests[floor][btn] != 0 {
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
	elevator.dirn = io.MD_Down
	elevator.state = Moving
}

func Fsm_OnRequestButtonPress(btn_Floor int, btn_type io.ButtonType) {
	switch elevator.state {
	case DoorOpen:
		if Requests_ShouldClearImmediately(elevator, btn_Floor, btn_type) != 0 {
			Timer_start(elevator.config.doorOpenDuration_s)
		} else {
			elevator.requests[btn_Floor][btn_type] = 1
		}
	case Moving:
		elevator.requests[btn_Floor][btn_type] = 1
	case Idle:
		elevator.requests[btn_Floor][btn_type] = 1
		var pair DirnBehaviourPair = Requests_chooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.state = pair.state
		switch pair.state {
		case DoorOpen:
			io.SetDoorOpenLamp(true)
			Timer_start(elevator.config.doorOpenDuration_s)
			elevator = Requests_clearAtCurrentFloor(elevator)
		case Moving:
			io.SetMotorDirection(elevator.dirn)
		case Idle:
			break
		}
	}
	SetAllLights(elevator)
}

func Fsm_OnFloorArrival(newFloor int) {
	elevator.floor = newFloor
	io.SetFloorIndicator(elevator.floor)

	switch elevator.state {
	case Moving:
		if Requests_shouldStop(elevator) != 0 {
			io.SetMotorDirection(io.MD_Stop)
			io.SetDoorOpenLamp(true)
			elevator = Requests_clearAtCurrentFloor(elevator)
			Timer_start(elevator.config.doorOpenDuration_s)
			SetAllLights(elevator)
			elevator.state = DoorOpen
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
	switch elevator.state {
	case DoorOpen:
		var pair DirnBehaviourPair = Requests_chooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.state = pair.state

		switch elevator.state {
		case DoorOpen:
			Timer_start(elevator.config.doorOpenDuration_s)
			elevator = Requests_clearAtCurrentFloor(elevator)
			SetAllLights(elevator)
		case Idle:
			io.SetDoorOpenLamp(false)
			io.SetMotorDirection(elevator.dirn)
		case Moving:
			io.SetDoorOpenLamp(false)
			io.SetMotorDirection(elevator.dirn)
		}
	default:
		break
	}
}
