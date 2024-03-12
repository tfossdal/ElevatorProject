package main

import (
	"fmt"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("Restarting the server ...")
	time.Sleep(5 * time.Second)
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go")
	cmd.Run()
	panic("Restarting")
}