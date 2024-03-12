package ElevatorModules

import (
	"fmt"

	"github.com/go-ping/ping"
)

func ConvertIDtoIP(id int) string {
	return "10.100.23." + fmt.Sprint(id)
}

func BoolToInt(n bool) int {
	if n {
		return 1
	}
	return 0
}

func PingInternet() int {
	_, err := ping.NewPinger("www.google.com")
	if err != nil {
		return 0
	}
	return 1
}
