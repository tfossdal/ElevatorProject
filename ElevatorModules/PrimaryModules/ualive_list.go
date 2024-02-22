package PrimaryModules

import (
	"container/list"
	"fmt"
	"time"
)

type Node struct {
	id       int
	lastSeen time.Time
}

func LivingElevatorHandler(elevatorLives, checkLiving, retrieveId, idOfLivingElev, printList, numberOfElevators chan int) {
	living := list.New()
	timeout := 1 * time.Second
	for {
		select {
		case elevId := <-elevatorLives:
			duplicate := false
			for e := living.Front(); e != nil; e = e.Next() {
				node := e.Value.(*Node)
				if elevId == node.id {
					node.lastSeen = time.Now()
					duplicate = true
					break
				}
			}
			if !duplicate {
				living.PushBack((&Node{id: elevId, lastSeen: time.Now()}))
			}
			break
		case <-checkLiving:

			for e := living.Front(); e != nil; e = e.Next() {
				node := e.Value.(*Node)
				if node.lastSeen.Add(timeout).Before(time.Now()) {
					fmt.Println("Removing")
					fmt.Println(node.lastSeen.String())
					living.Remove(e)
				}
			}
			break
		case whatToRetrieve := <-retrieveId:
			//check for living before extracting
			for e := living.Front(); e != nil; e = e.Next() {
				node := e.Value.(*Node)
				if node.lastSeen.Add(timeout).Before(time.Now()) {
					fmt.Println("Removing")
					fmt.Println(node.lastSeen.String())
					living.Remove(e)
				}
			}

			if whatToRetrieve == 1 {
				fmt.Println("Retrieving firster")
				e := living.Front() //Må finne løsning på ka som skjer
				if e == nil {       //om e ikke finnes eller er seg selv
					retrieveId <- 1
					fmt.Println("No living elevators")
					break
				}
				fmt.Println("Retrieving", e.Value.(*Node).id)
				idOfLivingElev <- e.Value.(*Node).id
				break
			} else {
				fmt.Println("Retrieving next")
				e := living.Front().Next() //Må finne løsning på ka som skjer
				if e == nil {              //om e ikke finnes eller er seg selv
					retrieveId <- 2
					break
				}
				idOfLivingElev <- e.Value.(*Node).id
				break
			}
		case <-printList:
			fmt.Println("Printing list:")
			θ := 0
			for e := living.Front(); e != nil; e = e.Next() {
				node := e.Value.(*Node)
				fmt.Println(node.id)
				θ++
			}
			numberOfElevators <- θ

		}

	}

}

// func main() {

// 	elevatorLives := make(chan int)
// 	checkLiving := make(chan int)
// 	retrieveId := make(chan int)
// 	idOfLivingElev := make(chan int)
// 	printList := make(chan int)
// 	go LivingElevatorHandler(elevatorLives, checkLiving, retrieveId, idOfLivingElev, printList)
// 	elevatorLives <- 1
// 	elevatorLives <- 2
// 	elevatorLives <- 1
// 	time.Sleep(6 * time.Second)
// 	elevatorLives <- 1
// 	time.Sleep(1 * time.Second)
// 	printList <- 1
// 	time.Sleep(1 * time.Second)
// 	checkLiving <- 1
// 	time.Sleep(1 * time.Second)
// 	retrieveId <- 1
// 	println(<-idOfLivingElev)
// 	printList <- 1
// 	time.Sleep(5 * time.Second)
// }
