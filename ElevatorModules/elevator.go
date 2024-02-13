package ElevatorModules

import (
	io "ElevatorProject/elevio"
)

type ElevatorType int

const (
	Primary ElevatorType = 0
	Backup  ElevatorType = 1
	None ElevatorType = 2
)

type State int

const (
	Idle     State = 0
	Moving   State = 1
	DoorOpen State = 2
)

type ClearRequestVariant int

const (
	CV_ALL    ClearRequestVariant = 0
	CV_InDirn ClearRequestVariant = 1
)

type Config struct {
	clearRequestVariant ClearRequestVariant
	doorOpenDuration_s  float64
}

type Elevator struct {
	floor    int
	dirn     io.MotorDirection
	requests [io.NumFloors][io.NumButtons]int
	state    State
	elevatorType ElevatorType

	config Config
}

func StateToString(state State) string {
	switch state {
	case Idle:
		return "State Idle"
	case Moving:
		return "State Moving"
	case DoorOpen:
		return "State DoorOpen"
	default:
		return "State Unknown"
	}
}
