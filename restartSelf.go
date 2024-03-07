package Restarting

import (
	"fmt"
	"os/exec"
	"time"
)

func main() {
	time.Sleep(5 * time.Second)
	cmd := exec.Command("go", "run", "main.go")
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
	}
}