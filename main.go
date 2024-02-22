package main

import (
	"ElevatorProject/ElevatorModules"
	module "ElevatorProject/ElevatorModules"
	io "ElevatorProject/elevio"
	"fmt"
)

func elev_init() {
	numFloors := 4

	io.Init("localhost:15657", numFloors)

	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan io.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go io.PollButtons(drv_buttons)
	go io.PollFloorSensor(drv_floors)
	go io.PollObstructionSwitch(drv_obstr)
	go io.PollStopButton(drv_stop)

	go module.CheckForTimeout()

	if io.GetFloor() == -1 {
		module.Fsm_onInitBetweenFloors()
	}

	module.InitLights()

	for {
		module.PrintState()
		select {
		case a := <-drv_buttons:
			//Button signal
			fmt.Printf("%+v\n", a)
			//io.SetButtonLamp(a.Button, a.Floor, true)
			module.Fsm_OnRequestButtonPress(a.Floor, a.Button)

		case a := <-drv_floors:
			//Floor signal
			fmt.Printf("%+v\n", a)
			module.Fsm_OnFloorArrival(a)

		case a := <-drv_obstr:
			//Obstruction
			fmt.Printf("%+v\n", a)

		case a := <-drv_stop:
			//Stop button signal
			fmt.Printf("%+v\n", a)
			//Turn all buttons off
			// for f := 0; f < numFloors; f++ {
			// 	for b := io.ButtonType(0); b < 3; b++ {
			// 		module.SetButtonLamp(b, f, false)
			// 	}
			// }
		}
	}
}

func main() {
	numFloors := 4

	io.Init("localhost:15657", numFloors)

	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan io.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go io.PollButtons(drv_buttons)
	go io.PollFloorSensor(drv_floors)
	go io.PollObstructionSwitch(drv_obstr)
	go io.PollStopButton(drv_stop)

	//elev_init()
	go module.IAmAlive()
	module.BecomePrimary()
	for {
		select {
		case a := <-drv_buttons:
			ElevatorModules.SendButtonPressUDP(a)
		}
	}

}

// Here I'm trying to test the backup and primary alive functions, but it is not working
/* import (
	module "ElevatorProject/ElevatorModules"
	"fmt"
)

func main() {
	go module.PrimaryAlive()
	go module.PrimaryAliveListener()
	fmt.Println("nothing happens")
}
*/
