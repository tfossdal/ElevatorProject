package ElevatorModules

import(
	"fmt"
	"net"
)

func CheckForPrimary(){
	addr, err := net.ResolveUDPAddr("udp4", ":29501")
	if err != nil {
		fmt.Println("Failed to resolve")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen")
	}
	defer conn.Close()
	buf := make([]byte, 1024)
	for{
	_, _, err = conn.ReadFromUDP(buf)
	fmt.Println("Read something")

		fmt.Println(buf[:])
	}
}