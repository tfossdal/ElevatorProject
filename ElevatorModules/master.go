package ElevatorModules

import (
	io "ElevatorProject/elevio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

var requests = make([][]int, io.NumFloors)

func InitMaster() {

	for i := 0; i < io.NumFloors; i++ {
		requests[i] = make([]int, io.NumButtons)
		for j := 0; j < io.NumButtons; j++ {
			requests[i][j] = 0
		}
	}
}

func OrderListener() {
	//29503
	addr, err := net.ResolveUDPAddr("udp4", ":29503")
	if err != nil {
		fmt.Println("Could not connect")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Could not listen")
	}
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Could not read")
		}
		orderStr := string(buf[:n])
		orderLst := strings.Split(orderStr, ",")
		floorIndex, _ := strconv.Atoi(orderLst[0])
		buttonIndex, _ := strconv.Atoi(orderLst[1])
		requests[floorIndex][buttonIndex] = 1
	}

}
