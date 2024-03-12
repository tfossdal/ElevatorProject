package PrimaryModules

import (
	"fmt"
	"time"
)


var livingElevatorMap = make(map[int]time.Time)

func LivingElevatorHandler(elevatorLives, checkLiving, retrieveId, idOfLivingElev, printList, newlyAlive chan int, listOfLiving chan map[int]time.Time) {
	timeout := 4 * time.Second
	for {
		select {
		case elevId := <-elevatorLives:
			duplicate := livingElevatorMap[elevId]
			livingElevatorMap[elevId] = time.Now()
			if duplicate.IsZero() {
				fmt.Println("Newly alive elevator", elevId)
				newlyAlive <- elevId
			}
		case <-checkLiving:
			for k, v := range livingElevatorMap {
				if v.Add(timeout).Before(time.Now()) {
					fmt.Println("Removing", k)
					delete(livingElevatorMap, k)
				}
			}
		case <-retrieveId:
			for k, v := range livingElevatorMap {
				if v.Add(timeout).Before(time.Now()) {
					fmt.Println("Removing", k)
					delete(livingElevatorMap, k)
				}
			}
			listOfLiving <- livingElevatorMap
		case <-printList:
			fmt.Println("Printing list:")
			fmt.Println(livingElevatorMap)

		}
	}
}

