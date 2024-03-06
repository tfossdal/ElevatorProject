package ElevatorModules

import (
	el "ElevatorProject/Elevator"
	io "ElevatorProject/elevio"
	"fmt"
)

type DirnBehaviourPair struct {
	dirn  io.MotorDirection
	state el.State
}

func requests_above(e el.Elevator) int {
	OrderMtx.Lock()
	defer OrderMtx.Unlock()
	for f := e.Floor + 1; f < io.NumFloors; f++ {
		for btn := 0; btn < io.NumButtons; btn++ {
			if e.Requests[f][btn] != 0 {
				return 1
			}
		}
	}
	return 0
}

func requests_below(e el.Elevator) int {
	OrderMtx.Lock()
	defer OrderMtx.Unlock()
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < io.NumButtons; btn++ {
			if e.Requests[f][btn] != 0 {
				return 1
			}
		}
	}
	return 0
}

func requests_here(e el.Elevator) int {
	OrderMtx.Lock()
	defer OrderMtx.Unlock()
	for btn := 0; btn < io.NumButtons; btn++ {
		if e.Requests[e.Floor][btn] != 0 {
			return 1
		}
	}
	return 0
}

func Requests_chooseDirection(e el.Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case io.MD_Up:
		if requests_above(e) != 0 {
			return DirnBehaviourPair{io.MD_Up, el.Moving}
		}
		if requests_here(e) != 0 {
			return DirnBehaviourPair{io.MD_Down, el.DoorOpen}
		}
		if requests_below(e) != 0 {
			return DirnBehaviourPair{io.MD_Down, el.Moving}
		}
		return DirnBehaviourPair{io.MD_Stop, el.Idle}
	case io.MD_Down:
		if requests_below(e) != 0 {
			return DirnBehaviourPair{io.MD_Down, el.Moving}
		}
		if requests_here(e) != 0 {
			return DirnBehaviourPair{io.MD_Up, el.DoorOpen}
		}
		if requests_above(e) != 0 {
			return DirnBehaviourPair{io.MD_Up, el.Moving}
		}
		return DirnBehaviourPair{io.MD_Stop, el.Idle}
	case io.MD_Stop:
		fmt.Println("yes")
		if requests_here(e) != 0 {
			return DirnBehaviourPair{io.MD_Stop, el.DoorOpen}
		}
		if requests_above(e) != 0 {
			return DirnBehaviourPair{io.MD_Up, el.Moving}
		}
		if requests_below(e) != 0 {
			return DirnBehaviourPair{io.MD_Down, el.Moving}
		}
		return DirnBehaviourPair{io.MD_Stop, el.Idle}
	default:
		return DirnBehaviourPair{io.MD_Stop, el.Idle}
	}
}

func Requests_shouldStop(e el.Elevator) int {
	OrderMtx.Lock()
	switch elevator.Dirn {
	case io.MD_Down:
		if e.Requests[e.Floor][io.BT_HallDown] != 0 {
			return 1
		}
		if e.Requests[e.Floor][io.BT_Cab] != 0 {
			return 1
		}
		OrderMtx.Unlock()
		if requests_below(e) == 0 {
			return 1
		}
		return 0
	case io.MD_Up:
		if e.Requests[e.Floor][io.BT_HallUp] != 0 {
			return 1
		}
		if e.Requests[e.Floor][io.BT_Cab] != 0 {
			return 1
		}
		OrderMtx.Unlock()
		if requests_above(e) == 0 {
			return 1
		}
		return 0
	default:
		OrderMtx.Unlock()
		return 1
	}
}

func Requests_ShouldClearImmediately(e el.Elevator, btn_floor int, btn_type io.ButtonType) int {
	switch e.Config.ClearRequestVariant {
	case el.CV_ALL:
		fmt.Println("Button floor: ", btn_floor)
		if e.Floor == btn_floor {
			return 1
		}
		return 0
	case el.CV_InDirn:
		if e.Floor == btn_floor &&
			(e.Dirn == io.MD_Up && btn_type == io.BT_HallUp) ||
			(e.Dirn == io.MD_Down && btn_type == io.BT_HallDown) ||
			e.Dirn == io.MD_Stop ||
			btn_type == io.BT_Cab {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func Requests_ClearImmediately_Online(e el.Elevator) {
	cleared := false
	for i := range e.Requests[e.Floor] {
		if e.Requests[e.Floor][i] == 1 {
			if i == int(io.BT_Cab) {
				e.Requests[e.Floor][i] = 0
				cleared = true
			}
			if e.Dirn == io.MD_Up && i == int(io.BT_HallUp) {
				e.Requests[e.Floor][i] = 0
				cleared = true
				break
			}
			if e.Dirn == io.MD_Down && i == int(io.BT_HallDown) {
				e.Requests[e.Floor][i] = 0
				cleared = true
				break
			}
		}
	}
	if cleared {
		Timer_start(elevator.Config.DoorOpenDuration_s)
	}
}

func Requests_clearAtCurrentFloor(e el.Elevator) el.Elevator {
	OrderMtx.Lock()
	defer OrderMtx.Unlock()
	e.Requests[e.Floor][io.BT_Cab] = 0
	switch e.Dirn {
	case io.MD_Up:
		if (requests_above(e) == 0) && (e.Requests[e.Floor][io.BT_HallUp] == 0) {
			e.Requests[e.Floor][io.BT_HallDown] = 0
		}
		e.Requests[e.Floor][io.BT_HallUp] = 0
	case io.MD_Down:
		if (requests_below(e) == 0) && (e.Requests[e.Floor][io.BT_HallDown] == 0) {
			e.Requests[e.Floor][io.BT_HallUp] = 0
		}
		e.Requests[e.Floor][io.BT_HallDown] = 0
	default:
		e.Requests[e.Floor][io.BT_HallUp] = 0
		e.Requests[e.Floor][io.BT_HallDown] = 0
	}
	return e
}
