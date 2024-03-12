package PrimaryModules

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func ListenUDP(port string, elevatorLives chan int, newOrderCh, clearOrderCh chan [3]int, newStatesCh chan [4]int) {

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
		if recieved_message[0] == "n" {
			floor, _ := strconv.Atoi(recieved_message[1])
			btn, _ := strconv.Atoi(recieved_message[2])
			order := [3]int{int(senderIP[3]), floor, btn}
			select {
			case newOrderCh <- order:
			default:
				fmt.Println("Order not accepted, buffer full")
			}

		}
		if recieved_message[0] == "s" {
			stateInt, _ := strconv.Atoi(recieved_message[1])
			directionInt, _ := strconv.Atoi(recieved_message[2])
			floorInt, _ := strconv.Atoi(recieved_message[3])
			newStates := [4]int{int(senderIP[3]), stateInt, directionInt, floorInt}
			select {
			case newStatesCh <- newStates:
			default:
				fmt.Println("New states not accepted, buffer full")
			}
		}
		if recieved_message[0] == "c" {
			floor, _ := strconv.Atoi(recieved_message[1])
			btn, _ := strconv.Atoi(recieved_message[2])
			order := [3]int{int(senderIP[3]), floor, btn}
			select {
			case clearOrderCh <- order:
			default:
				fmt.Println("New clear not accepted, buffer full")
			}
		}
		if err != nil {
			fmt.Println("Failed to listen")
		}
		elevatorLives <- int(senderIP[3])
	}
}
