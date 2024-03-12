package main

import (
	//io "ElevatorProject/elevio"
	"fmt"
	"net"
)

//const local_IPAdress = "10.24.18.4:33546"

func serverAccept() {
	addr, err := net.ResolveTCPAddr("tcp", ":33546")
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("Server is running...")
	conn, err := listener.AcceptTCP()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Connected %d", conn.RemoteAddr())
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	SendAck(string(buf[:n]), conn)

}

func SendAck(message string, conn *net.TCPConn) {
	fmt.Println(message)
	_, err := conn.Write([]byte(message))
	fmt.Println("Sent ack to primary")
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	serverAccept()
}
