package ElevatorModules

import (
	io "ElevatorProject/elevio"
	"fmt"
)

type DirnBehaviourPair struct {
	dirn  io.MotorDirection
	state State
}

func requests_above(e Elevator) int {
	for f := e.floor + 1; f < io.NumFloors; f++ {
		for btn := 0; btn < io.NumButtons; btn++ {
			if e.requests[f][btn] != 0 {
				return 1
			}
		}
	}
	return 0
}

func requests_below(e Elevator) int {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < io.NumButtons; btn++ {
			if e.requests[f][btn] != 0 {
				return 1
			}
		}
	}
	return 0
}

func requests_here(e Elevator) int {
	for btn := 0; btn < io.NumButtons; btn++ {
		if e.requests[e.floor][btn] != 0 {
			return 1
		}
	}
	return 0
}

func Requests_chooseDirection(e Elevator) DirnBehaviourPair {
	switch e.dirn {
	case io.MD_Up:
		if requests_above(e) != 0 {
			return DirnBehaviourPair{io.MD_Up, Moving}
		}
		if requests_here(e) != 0 {
			return DirnBehaviourPair{io.MD_Down, DoorOpen}
		}
		if requests_below(e) != 0 {
			return DirnBehaviourPair{io.MD_Down, Moving}
		}
		return DirnBehaviourPair{io.MD_Stop, Idle}
	case io.MD_Down:
		if requests_below(e) != 0 {
			return DirnBehaviourPair{io.MD_Down, Moving}
		}
		if requests_here(e) != 0 {
			return DirnBehaviourPair{io.MD_Up, DoorOpen}
		}
		if requests_above(e) != 0 {
			return DirnBehaviourPair{io.MD_Up, Moving}
		}
		return DirnBehaviourPair{io.MD_Stop, Idle}
	case io.MD_Stop:
		fmt.Println("yes")
		if requests_here(e) != 0 {
			return DirnBehaviourPair{io.MD_Stop, DoorOpen}
		}
		if requests_above(e) != 0 {
			return DirnBehaviourPair{io.MD_Up, Moving}
		}
		if requests_below(e) != 0 {
			return DirnBehaviourPair{io.MD_Down, Moving}
		}
		return DirnBehaviourPair{io.MD_Stop, Idle}
	default:
		return DirnBehaviourPair{io.MD_Stop, Idle}
	}
}

func Requests_shouldStop(e Elevator) int {
	switch elevator.dirn {
	case io.MD_Down:
		if e.requests[e.floor][io.BT_HallDown] != 0 {
			return 1
		}
		if e.requests[e.floor][io.BT_Cab] != 0 {
			return 1
		}
		if requests_below(e) == 0 {
			return 1
		}
		return 0
	case io.MD_Up:
		if e.requests[e.floor][io.BT_HallUp] != 0 {
			return 1
		}
		if e.requests[e.floor][io.BT_Cab] != 0 {
			return 1
		}
		if requests_above(e) == 0 {
			return 1
		}
		return 0
	default:
		return 1
	}
}

func Requests_ShouldClearImmediately(e Elevator, btn_floor int, btn_type io.ButtonType) int {
	switch e.config.clearRequestVariant {
	case CV_ALL:
		fmt.Println("Button floor: ", btn_floor)
		if e.floor == btn_floor {
			return 1
		}
		return 0
	case CV_InDirn:
		if e.floor == btn_floor &&
			(e.dirn == io.MD_Up && btn_type == io.BT_HallUp) ||
			(e.dirn == io.MD_Down && btn_type == io.BT_HallDown) ||
			e.dirn == io.MD_Stop ||
			btn_type == io.BT_Cab {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func Requests_clearAtCurrentFloor(e Elevator) Elevator {
	switch e.config.clearRequestVariant {
	case CV_ALL:
		for btn := 0; btn < io.NumButtons; btn++ {
			e.requests[e.floor][btn] = 0
		}
	case CV_InDirn:
		e.requests[e.floor][io.BT_Cab] = 0
		switch e.dirn {
		case io.MD_Up:
			if (requests_above(e) == 0) && (e.requests[e.floor][io.BT_HallUp] == 0) {
				e.requests[e.floor][io.BT_HallDown] = 0
			}
			e.requests[e.floor][io.BT_HallUp] = 0
		case io.MD_Down:
			if (requests_below(e) == 0) && (e.requests[e.floor][io.BT_HallDown] == 0) {
				e.requests[e.floor][io.BT_HallUp] = 0
			}
			e.requests[e.floor][io.BT_HallDown] = 0
		default:
			e.requests[e.floor][io.BT_HallUp] = 0
			e.requests[e.floor][io.BT_HallDown] = 0
		}
	default:
	}
	return e
}
