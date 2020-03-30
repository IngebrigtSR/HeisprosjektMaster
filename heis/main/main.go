package main

import (
	"fmt"

	"../elevio"
	"../fsm"
)

func main() {
	fmt.Println("Hello World")

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)

	go fsm.ElevFSM(drv_buttons, drv_floors)

	a := 1
	for {
		a++
	}
}
