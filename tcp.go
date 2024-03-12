package main

// import (
// 	"fmt"
// 	"net"

// 	"github.com/go-ping/ping"
// )

// func WaitForAck(message string, conn *net.TCPConn) bool {
// 	buf := make([]byte, 1024)
// 	fmt.Println("Message to ack: " + message)
// 	n, err := conn.Read(buf)
// 	fmt.Println("Read ack")
// 	if err != nil {
// 		fmt.Println("Failed to read ack")
// 		return false
// 	}
// 	fmt.Println("Recieved ack message: " + string(buf[:n]))
// 	return string(buf[:n]) == message
// }

// func TCPtest() {
// 	addr, err := net.ResolveTCPAddr("tcp", "10.100.23.192:33546")
// 	if err != nil {
// 		panic(err)
// 	}
// 	conn, err := net.DialTCP("tcp", nil, addr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()
// 	fmt.Println("Connected?")
// 	stringToSend := "HELLO"
// 	for {
// 		_, err = conn.Write([]byte(stringToSend))
// 		if err != nil {
// 			fmt.Print(err)
// 			return
// 		}
// 		if WaitForAck(stringToSend, conn) {
// 			break
// 		}
// 	}
// 	fmt.Println("DONE")

// }

// func main() {
// 	TCPtest()
// }

// func PingInternet() int {
// 	_, err := ping.NewPinger("www.google.com")
// 	if err != nil {
// 		return 0
// 	}
// 	return 1
// }
