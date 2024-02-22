package PrimaryModules

import (
	"fmt"
	"net"
	"strings"
	"strconv"
)

func ListenUDP(port string, elevatorLives chan int, newOrderCh chan [3]int) {

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
		senderIP := recievedAddr.IP
		recieved_message := strings.Split(string(buf[:n]), ",")
		//fmt.Println("Read something")
		if recieved_message[0] == "n" {
			floor, _ := strconv.Atoi(recieved_message[2])
			btnType, _ := strconv.Atoi(recieved_message[3])
			order := [3]int{int(senderIP[3]), floor, btnType}
			newOrderCh <- order
		}
		if err != nil {
			fmt.Println("Failed to listen")
		}
		//fmt.Println(string(buf[:n])) 						//testing
		//SenderNumber := fmt.Sprintf("%d", senderIP[3])	//testing
		//fmt.Println(senderIP.String()) 					//testing
		//fmt.Println(SenderNumber) 						//testing
		elevatorLives <- int(senderIP[3])
	}
}

// func main() {
// 	ListenUDP("29505")
// }
