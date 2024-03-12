package main

import (
	"ElevatorProject/ElevatorModules"
	module "ElevatorProject/ElevatorModules"
	io "ElevatorProject/elevio"
	"fmt"
)

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

	module.CheckForPrimary()

	go module.CheckForDoorTimeout()

	if io.GetFloor() == -1 {
		module.Fsm_onInitBetweenFloors()
	}

	//elev_init()
	module.InitLights()
	go module.IAmAlive()
	go module.RecieveTurnOnOffLight()
	go ElevatorModules.RecieveOrderMatrix()
	go module.Fsm_Obstructed()

	for {
		select {
		case a := <-drv_buttons:
			if int(a.Button) == 2 {
				if module.PingInternet() == 0 {
					module.Fsm_OnRequestButtonPress(a.Floor, a.Button)
				} else {
					fmt.Println("sending button press")
					ElevatorModules.SendButtonPressUDP(a)
					ElevatorModules.AddCabRequest(a.Floor, a.Button)
				}
			} else {
				fmt.Println("sending button press")
				ElevatorModules.SendButtonPressUDP(a)
			}
		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			module.Fsm_OnFloorArrival(a)
		case a := <-drv_obstr:
			//Obstruction
			fmt.Printf("%+v\n", a)
			fmt.Println("OBSTRUUUUUUUCTING!!!!!!!!!!")
			module.IsObstructed = a
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
