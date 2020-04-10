package fsm

import (
	"fmt"

	. "../config"
	"../elevio"
	"../orderhandler"
)

const numFloors int = 4
const numButtons int = 3

var activeOrders [numFloors][numButtons]bool

func printOrder() {
	fmt.Printf("Active Orders: \n")
	for f := 0; f < numFloors; f++ {
		for b := 0; b < numButtons; b++ {
			if activeOrders[f][b] {
				fmt.Printf("%d\t", 1)

			} else {
				fmt.Printf("%d\t", 0)
			}
		}
		fmt.Printf("\n")
	}
}

func shouldStop(elev orderhandler.Elevator) bool {
	if elev.Floor == 0 || elev.Floor == numFloors-1 {
		return true
	}
	if activeOrders[elev.Floor][elevio.BT_HallUp] && elev.Dir == elevio.MD_Up {
		return true
	}
	if activeOrders[elev.Floor][elevio.BT_HallDown] && elev.Dir == elevio.MD_Down {
		return true
	}
	if activeOrders[elev.Floor][elevio.BT_Cab] {
		return true
	}
	if !orderhandler.OrdersInFront(elev) {
		if activeOrders[elev.Floor][elevio.BT_HallUp] || activeOrders[elev.Floor][elevio.BT_HallDown] {
			return true
		}
	}
	return false
}

func anyActiveOrders() bool {
	for f := 0; f < numFloors; f++ {
		for b := 0; b < numButtons; b++ {
			if activeOrders[f][b] {
				return true
			}
		}
	}
	return false
}

func setDir(elev orderhandler.Elevator) elevio.MotorDirection {
	if !anyActiveOrders() {
		return elevio.MD_Stop
	}
	if elev.Dir == elevio.MD_Stop {
		for f := 0; f < numFloors; f++ {
			for b := 0; b < numButtons; b++ {
				if activeOrders[f][b] {
					if f < elev.Floor {
						return elevio.MD_Down
					}
					if f > elev.Floor {
						return elevio.MD_Up
					}
				}
			}
		}
	}
	if orderhandler.OrdersInFront(elev) {
		return elevio.MotorDirection(elev.Dir)
	}
	return elevio.MotorDirection(-elev.Dir)
}

func clearFloorOrders(floor int) {
	for button := elevio.BT_HallUp; button <= elevio.BT_Cab; button++ {
		activeOrders[floor][button] = false
		elevio.SetButtonLamp(button, floor, false)
	}
	// printOrder()
}

func takeOrder(floor int, button elevio.ButtonType) {
	activeOrders[floor][button] = true
	elevio.SetButtonLamp(button, floor, true)
	// printOrder()
}

func initFSM() {
	elevio.Init("localhost:15657", numFloors)
	//clears all orders
	for f := 0; f < numFloors; f++ {
		clearFloorOrders(f)
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
}

func updateLights(log orderhandler.ElevLog) {
	for i := 0; i < NumElevators; i++ {
		for b := 0; b < NumButtons-1; b++ {
			for f := 0; f < NumFloors; f++ {
				if log[i].Orders[f][b] {
					elevio.SetButtonLamp(elevio.ButtonType(b), f, true)
				}
			}
		}
		if i == LogIndex {
			for f := 0; f < NumFloors; f++ {
				if log[i].Orders[f][2] {
					elevio.SetButtonLamp(elevio.BT_Cab, f, true)
				}
			}
		}
	}
}

func ElevFSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int) {
	// elevio.Init("localhost:15657", numFloors)

	var elev orderhandler.Elevator
	initFSM()

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	for {
		select {
		case order := <-drv_buttons:
			fmt.Printf("Order:\t%+v\n", order)
			if !(order.Floor == elev.Floor && elev.Dir == elevio.MD_Stop) {
				takeOrder(order.Floor, order.Button)
			}
			elev.Dir = setDir(elev)
			elevio.SetMotorDirection(elev.Dir)

		case floor := <-drv_floors:
			fmt.Printf("Floor:\t%+v\n", floor)
			elevio.SetFloorIndicator(floor)
			elev.Floor = floor

			if shouldStop(elev) {
				clearFloorOrders(floor)
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
			elev.Dir = setDir(elev)
			elevio.SetMotorDirection(elev.Dir)

		}
	}
}
