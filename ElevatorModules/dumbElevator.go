package ElevatorModules

import (
	io "ElevatorProject/elevio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var IsObstructed = false

func IAmAlive() {
	var conn = &net.UDPConn{}
	for {
		addr, err := net.ResolveUDPAddr("udp4", "10.100.23.255:29503")
		if err != nil {
			fmt.Println("Failed to resolve, send order")
			continue
		}
		conn, err = net.DialUDP("udp4", nil, addr)
		if err != nil {
			fmt.Println("Failed to dial, send order")
			continue
		}
		defer conn.Close()
		break
	}
	for {
		UanbleToMoveMtx.Lock()
		localMovabiility := IsUnableToMove
		UanbleToMoveMtx.Unlock()
		if !localMovabiility {
			state := strconv.Itoa(int(elevator.State))
			direction := strconv.Itoa(int(elevator.Dirn))
			floor := strconv.Itoa(int(elevator.Floor))
			conn.Write([]byte("s," + state + "," + direction + "," + floor))
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func CheckGoneOffline() {
	for {
		if PingInternet() == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("Elevator went offline")
	time.Sleep(10 * time.Second)
	go ListenForOtherPrimary()
}

// func waitForAckUDP(message string, conn *net.UDPConn) bool {
// 	go readAckUDP(conn)
// 	startTime := time.Now().Unix()
// 	for {
// 		select {
// 		case ackMessage := <-ackCh:
// 			return ackMessage == message
// 		default:
// 			if time.Now().Unix() > startTime+1 {
// 				return false
// 			}
// 			time.Sleep(100 * time.Millisecond)
// 		}
// 	}
// }

// func readAckUDP(conn *net.UDPConn) {
// 	buf := make([]byte, 1024)
// 	n, _ := conn.Read(buf)
// 	ackCh <- string(buf[:n])
// }

func SendButtonPressUDP(btn io.ButtonEvent) {
	addr, err := net.ResolveUDPAddr("udp4", "10.100.23.255:29503")
	if err != nil {
		fmt.Println("Failed to resolve, send order")
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial, send order")
		return
	}
	messageToSend := "n," + fmt.Sprint(btn.Floor) + "," + fmt.Sprint(btn.Button)
	_, err = conn.Write([]byte(messageToSend))
	if err != nil {
		fmt.Println(err)
		fmt.Println("Failed to write, send order")
	}
	conn.Close()
}

func ClearRequestUDP(btn io.ButtonEvent) {
	addr, err := net.ResolveUDPAddr("udp4", "10.100.23.255:29503")
	if err != nil {
		fmt.Println("Failed to resolve, send order")
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial, send order")
		return
	}
	defer conn.Close()
	_, err = conn.Write([]byte("c," + fmt.Sprint(btn.Floor) + "," + fmt.Sprint(btn.Button)))
	if err != nil {
		fmt.Println(err)
	}
}

func RecieveTurnOnOffLight() {
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
			fmt.Println("Failed to read, recieve turn on/off light")
		}
		recievedMessage := string(buf[:n])
		messageList := strings.Split(recievedMessage, ",")
		btnInt, err := strconv.Atoi(messageList[2])
		if err != nil {
			fmt.Println("Failed to convert to btn, recieve turn on/off light")
		}
		btn := io.ButtonType(btnInt)
		floor, err := strconv.Atoi(messageList[1])
		turnOn := messageList[3] == "1"
		io.SetButtonLamp(btn, floor, turnOn)
		if err != nil {
			fmt.Println("Failed to convert to floor, recieve turn on/off light")
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
		fmt.Println("Failed to listen, recieve order matrix")
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
