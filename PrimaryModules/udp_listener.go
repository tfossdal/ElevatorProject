package PrimaryModules

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func ListenUDP(port string, elevatorLives chan int, newOrderCh chan [3]int, newStatesCh chan [4]int) {

	addr, err := net.ResolveUDPAddr("udp4", ":"+port)
	if err != nil {
		fmt.Println("Failed to resolve")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen")
	}
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, recievedAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		}
		senderIP := recievedAddr.IP
		recieved_message := strings.Split(string(buf[:n]), ",")
		fmt.Println(string(buf[:n]) + "yo")
		if recieved_message[0] == "n" {
			fmt.Println("test")
			floor, _ := strconv.Atoi(recieved_message[1])
			btn, _ := strconv.Atoi(recieved_message[2])
			order := [3]int{int(senderIP[3]), floor, btn}
			newOrderCh <- order
		}
		if recieved_message[0] == "s" {
			stateInt, _ := strconv.Atoi(recieved_message[1])
			directionInt, _ := strconv.Atoi(recieved_message[2])
			floorInt, _ := strconv.Atoi(recieved_message[3])
			// state := el.State(stateInt)
			// direction := io.MotorDirection(directionInt)
			newStates := [4]int{int(senderIP[3]), stateInt, directionInt, floorInt}
			newStatesCh <- newStates
		}
		if err != nil {
			fmt.Println("Failed to listen")
		}
		//fmt.Println(string(buf[:n])) 						//testing
		//SenderNumber := fmt.Sprintf("%d", senderIP[3])	//testing
		//fmt.Println(senderIP.String()) 					//testing
		//fmt.Println(SenderNumber)
		fmt.Println("If this is last, elevatorLives is causing trouble") //testing
		elevatorLives <- int(senderIP[3])
		fmt.Println("I even got down here")
	}
}

// func main() {
// 	ListenUDP("29505")
// }
