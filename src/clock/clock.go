package clock

import "fmt"

var clock int = 0

func UpdateClock() {
	clock++
	fmt.Println("\t=> Atualizando relogio para ", clock)
}
