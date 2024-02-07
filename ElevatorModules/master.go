package ElevatorModules

import io "ElevatorProject/elevio"

func Master(){
	requests := make([][]int, io.NumFloors)
	
	for i := 0; i < io.NumFloors; i++{
		requests[i] = make([]int, io.NumButtons)
		for j := 0; j < io.NumButtons; j++{
			requests[i][j] = 0
		}
	}
}
