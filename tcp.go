package main

/*  import (
 	"fmt"
	"time"
	"github.com/go-ping/ping"
 ) */

// func TCPtest() {
// 	addr, err := net.ResolveTCPAddr("tcp", "10.100.23.28:33546")
// 	if err != nil {
// 		panic(err)
// 	}
// 	conn, err := net.DialTCP("tcp", nil, addr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()
// 	fmt.Println("Connected?")
// }

// func main() {
// 	TCPtest()
// }

/* func PingInternet() int {
	_, err := ping.NewPinger("www.google.com")
	if err != nil {
		return 0
	}
	return 1
}

func main() {
for {
	if PingInternet() == 1 {
		fmt.Println("Connected")
	} else {
		fmt.Println("Not connected")
	}
	time.Sleep(1 * time.Second)
}

} */