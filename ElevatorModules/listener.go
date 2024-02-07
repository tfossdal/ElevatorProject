package ElevatorModules

import(
	"time"
)

func CheckForMaster(){
	for{
		//Set up Lister-Timer
		//Wait for "Im primary" message
		//If timeout - Become primary
		time.Sleep(1000000000)
	}
}