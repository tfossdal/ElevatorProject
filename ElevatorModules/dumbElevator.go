package ElevatorModules

import (
	io "ElevatorProject/elevio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

//var QueueHasBeenUpdated = make(chan int)

func IAmAlive() {
	state := strconv.Itoa(int(elevator.State))
	direction := strconv.Itoa(int(elevator.Dirn))
	floor := strconv.Itoa(int(elevator.Floor))
	addr, err := net.ResolveUDPAddr("udp4", "10.100.23.255:29503")
	if err != nil {
		fmt.Println("Failed to resolve, send order")
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial, send order")
	}
	defer conn.Close()
	for {
		//fmt.Println("Sending message")
		conn.Write([]byte("s," + state + "," + direction + "," + floor))
		//fmt.Println("Message sent: Elevator alive")
		time.Sleep(100 * time.Millisecond)
	}
}

func SendButtonPressUDP(btn io.ButtonEvent) {
	addr, err := net.ResolveUDPAddr("udp4", "10.100.23.255:29503")
	if err != nil {
		fmt.Println("Failed to resolve, send order")
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial, send order")
	}
	defer conn.Close()
	_, err = conn.Write([]byte("n," + fmt.Sprint(btn.Floor) + "," + fmt.Sprint(btn.Button)))
	if err != nil {
		fmt.Println(err)
	}
}

func RecieveTurnOnLight() {
	addr, err := net.ResolveUDPAddr("udp4", ":29505")
	if err != nil {
		fmt.Println("Failed to resolve, recieve turn on light")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen, recieve turn on light")
	}
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Failed to read, recieve turn on light")
		}
		recievedMessage := string(buf[:n])
		fmt.Printf("Read " + recievedMessage)
		messageList := strings.Split(recievedMessage, ",")
		btnInt, err := strconv.Atoi(messageList[2])
		if err != nil {
			fmt.Println("Failed to convert to btn, recieve turn on light")
		}
		btn := io.ButtonType(btnInt)
		floor, err := strconv.Atoi(messageList[1])
		io.SetButtonLamp(btn, floor, true)
		if err != nil {
			fmt.Println("Failed to convert to floor, recieve turn on light")
		}
	}

}

func RecieveOrderMatrix() {
	addr, err := net.ResolveUDPAddr("udp4", ":29504")
	if err != nil {
		fmt.Println("Failed to resolve, recieve order matrix")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to resolve listen, recieve order matrix")
	}
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
		}
		recieved_message := strings.Split(string(buf[:n]), ",")
		newMatrix := [io.NumFloors][2]int{}
		for flr := range newMatrix {
			upOrder, _ := strconv.Atoi(recieved_message[2*flr])
			downOrder, _ := strconv.Atoi(recieved_message[2*flr+1])
			newMatrix[flr][0] = upOrder
			newMatrix[flr][1] = downOrder
		}
		UpdateLocalRequestMatrix(newMatrix)
	}
}
