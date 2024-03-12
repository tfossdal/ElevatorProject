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
		for {
			if io.GetFloor() != -1{
				io.SetMotorDirection(io.MD_Stop)
				break
			}
		}
	}

	module.InitLights()
	go module.IAmAlive()
	go module.RecieveTurnOnOffLight()
	go ElevatorModules.RecieveOrderMatrix()
	go module.Fsm_Obstructed()
	go module.CheckMoveAvailability()

	for {
		select {
		case a := <-drv_buttons:
			if int(a.Button) == 2 {
				if module.PingInternet() == 0 {
					module.Fsm_OnRequestButtonPress(a.Floor, a.Button)
				} else {
					ElevatorModules.SendButtonPressUDP(a)
					ElevatorModules.AddCabRequest(a.Floor, a.Button)
				}
			} else {
				ElevatorModules.SendButtonPressUDP(a)
			}
		case a := <-drv_floors:
			module.Fsm_OnFloorArrival(a)
		case a := <-drv_obstr:
			module.IsObstructed = a
		}
	}

}
