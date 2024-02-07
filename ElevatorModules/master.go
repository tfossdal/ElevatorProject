package ElevatorModules

import io "ElevatorProject/elevio"

func Master(){
	requests := make([][]int, io.numFloors)

	for i := 0; i < io.numFloors; i++{
		requests[i] = make([]int, io.NumButtons)
		for j := 0; j < io.NumButtons; j++{
			requests[i][j] = 0
		}
	}
}
