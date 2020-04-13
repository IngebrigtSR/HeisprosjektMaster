package main

import (
	"fmt"
	"time"

	"../orderhandler"
	"../network/networkmanager"
)

func main() {
	fmt.Println("Hello World")

	// TODO: Sjekk om det allerede finnes andre heiser i loggen fra før av
	
	var log []Elevator
	networkmanager.InitNewElevator(&log)
	localIndex := networkmanager.GetLocalIndex(log)

	// TODO: Endre på InitFSM slik at den opererer med localIndex

	// elevio.Init("localhost:15657", config.NumFloors)

	// drv_buttons := make(chan elevio.ButtonEvent)
	// drv_floors := make(chan int)
	// startUp := make(chan bool)

	// go elevio.PollButtons(drv_buttons)
	// go elevio.PollFloorSensor(drv_floors)

	// go fsm.ElevFSM(drv_buttons, drv_floors, startUp)

	log := orderhandler.MakeEmptyLog()
	orderhandler.TestCost(log)

	transmitter := time.NewTicker(1000 * time.Millisecond)
	timer := time.NewTimer(5 * time.Second)
	transmit := true
	count := 0
	for {
		select {

		case <-timer.C:
			println("walla walla bing bang")

		case <-transmitter.C:
			if transmit {
				count++
				println("transmitted: \t", count)
			}
		}
	}
}
