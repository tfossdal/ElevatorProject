package main

import (
	"fmt"
	"sanntid/elevio"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go elevio.CheckForTimeout()

	if elevio.GetFloor() == -1 {
		elevio.Fsm_onInitBetweenFloors()
	}

	elevio.InitLights()

	for {
		elevio.PrintState()
		select {
		case a := <-drv_buttons:
			//Button signal
			fmt.Printf("%+v\n", a)
			//elevio.SetButtonLamp(a.Button, a.Floor, true)
			elevio.Fsm_OnRequestButtonPress(a.Floor, a.Button)

		case a := <-drv_floors:
			//Floor signal
			fmt.Printf("%+v\n", a)
			elevio.Fsm_OnFloorArrival(a)

		case a := <-drv_obstr:
			//Obstruction
			fmt.Printf("%+v\n", a)

		case a := <-drv_stop:
			//Stop button signal
			fmt.Printf("%+v\n", a)
			//Turn all buttons off
			// for f := 0; f < numFloors; f++ {
			// 	for b := elevio.ButtonType(0); b < 3; b++ {
			// 		elevio.SetButtonLamp(b, f, false)
			// 	}
			// }
		}
	}
}
