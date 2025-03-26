package clock

import "fmt"

var clock int = 0

func UpdateClock() int {
	clock++
	fmt.Println("\t=> Atualizando relogio para ", clock)
	return clock
}

func ResetClock() {
	clock = 0
}
