// package main

// import (
// 	//io "ElevatorProject/elevio"
// 	"fmt"
// 	"net"
// )

// //const local_IPAdress = "10.24.18.4:33546"

// func serverAccept() {
// 	addr, err := net.ResolveTCPAddr("tcp", ":33546")
// 	if err != nil {
// 		panic(err)
// 	}
// 	listener, err := net.ListenTCP("tcp", addr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer listener.Close()
// 	fmt.Println("Server is running...")
// 	conn, err := listener.AcceptTCP()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("Connected %d", conn.RemoteAddr())
// 	//_, err = s2c_conn.Write(append([]byte("Accepted from group 12"), 0))
// }

// func main(){
// 	serverAccept();
// }