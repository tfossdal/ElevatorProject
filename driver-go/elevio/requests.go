package elevio

import "fmt"

type DirnBehaviourPair struct {
	dirn  MotorDirection
	state State
}

func requests_above(e Elevator) int {
	for f := e.floor + 1; f < _numFloors; f++ {
		for btn := 0; btn < _numButtons; btn++ {
			if e.requests[f][btn] != 0 {
				return 1
			}
		}
	}
	return 0
}

func requests_below(e Elevator) int {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < _numButtons; btn++ {
			if e.requests[f][btn] != 0 {
				return 1
			}
		}
	}
	return 0
}

func requests_here(e Elevator) int {
	for btn := 0; btn < _numButtons; btn++ {
		if e.requests[e.floor][btn] != 0 {
			return 1
		}
	}
	return 0
}

func Requests_chooseDirection(e Elevator) DirnBehaviourPair {
	switch e.dirn {
	case MD_Up:
		if requests_above(e) != 0 {
			return DirnBehaviourPair{MD_Up, Moving}
		}
		if requests_here(e) != 0 {
			return DirnBehaviourPair{MD_Down, DoorOpen}
		}
		if requests_below(e) != 0 {
			return DirnBehaviourPair{MD_Down, Moving}
		}
		return DirnBehaviourPair{MD_Stop, Idle}
	case MD_Down:
		if requests_below(e) != 0 {
			return DirnBehaviourPair{MD_Down, Moving}
		}
		if requests_here(e) != 0 {
			return DirnBehaviourPair{MD_Up, DoorOpen}
		}
		if requests_above(e) != 0 {
			return DirnBehaviourPair{MD_Up, Moving}
		}
		return DirnBehaviourPair{MD_Stop, Idle}
	case MD_Stop:
		fmt.Println("yes")
		if requests_here(e) != 0 {
			return DirnBehaviourPair{MD_Stop, DoorOpen}
		}
		if requests_above(e) != 0 {
			return DirnBehaviourPair{MD_Up, Moving}
		}
		if requests_below(e) != 0 {
			return DirnBehaviourPair{MD_Down, Moving}
		}
		return DirnBehaviourPair{MD_Stop, Idle}
	default:
		return DirnBehaviourPair{MD_Stop, Idle}
	}
}

func Requests_shouldStop(e Elevator) int {
	switch elevator.dirn {
	case MD_Down:
		if e.requests[e.floor][BT_HallDown] != 0 {
			return 1
		}
		if e.requests[e.floor][BT_Cab] != 0 {
			return 1
		}
		if requests_below(e) == 0 {
			return 1
		}
		return 0
	case MD_Up:
		if e.requests[e.floor][BT_HallUp] != 0 {
			return 1
		}
		if e.requests[e.floor][BT_Cab] != 0 {
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

func Requests_ShouldClearImmediately(e Elevator, btn_floor int, btn_type ButtonType) int {
	switch e.config.clearRequestVariant {
	case CV_ALL:
		fmt.Println("Button floor: ", btn_floor)
		if e.floor == btn_floor {
			return 1
		}
		return 0
	case CV_InDirn:
		if e.floor == btn_floor &&
			(e.dirn == MD_Up && btn_type == BT_HallUp) ||
			(e.dirn == MD_Down && btn_type == BT_HallDown) ||
			e.dirn == MD_Stop ||
			btn_type == BT_Cab {
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
		for btn := 0; btn < _numButtons; btn++ {
			e.requests[e.floor][btn] = 0
		}
	case CV_InDirn:
		e.requests[e.floor][BT_Cab] = 0
		switch e.dirn {
		case MD_Up:
			if (requests_above(e) == 0) && (e.requests[e.floor][BT_HallUp] == 0) {
				e.requests[e.floor][BT_HallDown] = 0
			}
			e.requests[e.floor][BT_HallUp] = 0
		case MD_Down:
			if (requests_below(e) == 0) && (e.requests[e.floor][BT_HallDown] == 0) {
				e.requests[e.floor][BT_HallUp] = 0
			}
			e.requests[e.floor][BT_HallDown] = 0
		default:
			e.requests[e.floor][BT_HallUp] = 0
			e.requests[e.floor][BT_HallDown] = 0
		}
	default:
	}
	return e
}
