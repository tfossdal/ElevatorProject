package main

// import (
// 	"fmt"

// 	"github.com/go-ping/ping"
// )

// func SendOrder() {

// 	addr, err := net.ResolveUDPAddr("udp4", ":29505")
// 	if err != nil {
// 		fmt.Println("Failed to resolve, send order")
// 	}
// 	conn, err := net.DialUDP("udp4", nil, addr)
// 	if err != nil {
// 		fmt.Println("Failed to dial, send order")
// 	}
// 	defer conn.Close()

// 	fmt.Println("Sending message")
// 	conn.Write([]byte("4"))
// 	//fmt.Println("Message sent: Primary alive")
// 	time.Sleep(10 * time.Millisecond)

// }

// func main() {
// 	//SendOrder();
// 	_, err := ping.NewPinger("www.google.com")
// 	if err != nil {
// 		fmt.Println("test")
// 	}
// 	fmt.Println("testtastsata")
// }
