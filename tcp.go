package main

import (
	"fmt"
	"net"
)

func TCPtest() {
	addr, err := net.ResolveTCPAddr("tcp", ":33546")
	if err != nil {
		panic(err)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Connected?")
}

