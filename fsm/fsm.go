package fsm

import (
	"fmt"

	"../elevio"
)

const numFloors int = 4
const numButtons int = 3

var activeOrders [numFloors][numButtons]bool

type Elevator struct {
	floor  int
	dir    elevio.MotorDirection
	state  int
	orders [numFloors][numButtons]bool
}

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

func shouldStop(floor int, dir elevio.MotorDirection) bool {
	if floor == 0 || floor == numFloors-1 {
		return true
	}
	if activeOrders[floor][elevio.BT_HallUp] && dir == elevio.MD_Up {
		return true
	}
	if activeOrders[floor][elevio.BT_HallDown] && dir == elevio.MD_Down {
		return true
	}
	if activeOrders[floor][elevio.BT_Cab] {
		return true
	}
	if !ordersInFront(floor, int(dir)) {
		if activeOrders[floor][elevio.BT_HallUp] || activeOrders[floor][elevio.BT_HallDown] {
			return true
		}
	}
	return false
}

func ordersInFront(floor int, dir int) bool {
	if dir == int(elevio.MD_Stop) {
		return false
	}
	for f := floor + dir; 0 <= f && f < numFloors; f += dir {
		for b := 0; b < numButtons; b++ {
			if activeOrders[f][b] {
				return true
			}
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

func setDir(floor int, dir int) elevio.MotorDirection {
	if !anyActiveOrders() {
		return elevio.MD_Stop
	}
	if dir == int(elevio.MD_Stop) {
		for f := 0; f < numFloors; f++ {
			for b := 0; b < numButtons; b++ {
				if activeOrders[f][b] {
					if f < floor {
						return elevio.MD_Down
					}
					if f > floor {
						return elevio.MD_Up
					}
				}
			}
		}
	}
	if ordersInFront(floor, dir) {
		return elevio.MotorDirection(dir)
	}
	return elevio.MotorDirection(-dir)
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

func ElevFSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int) {
	// elevio.Init("localhost:15657", numFloors)

	var elev Elevator
	initFSM()

	// drv_buttons := make(chan elevio.ButtonEvent)
	// drv_floors := make(chan int)
	// drv_obstr := make(chan bool)
	// drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	// go elevio.PollObstructionSwitch(drv_obstr)
	// go elevio.PollStopButton(drv_stop)

	for {
		select {
		case order := <-drv_buttons:
			fmt.Printf("Order:\t%+v\n", order)
			if !(order.Floor == elev.floor && elev.dir == elevio.MD_Stop) {
				takeOrder(order.Floor, order.Button)
			}
			elev.dir = setDir(elev.floor, int(elev.dir))
			elevio.SetMotorDirection(elev.dir)

		case floor := <-drv_floors:
			fmt.Printf("Floor:\t%+v\n", floor)
			elevio.SetFloorIndicator(floor)
			elev.floor = floor

			if shouldStop(floor, elev.dir) {
				clearFloorOrders(floor)
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
			elev.dir = setDir(floor, int(elev.dir))
			elevio.SetMotorDirection(elev.dir)

			// case a := <-drv_obstr:
			// 	fmt.Printf("%+v\n", a)
			// 	if a {
			// 		elevio.SetMotorDirection(elevio.MD_Stop)
			// 	} else {
			// 		elevio.SetMotorDirection(elev.dir)
			// 	}

			// case a := <-drv_stop:
			// 	fmt.Printf("%+v\n", a)
			// 	for f := 0; f < numFloors; f++ {
			// 		for b := elevio.ButtonType(0); b < 3; b++ {
			// 			elevio.SetButtonLamp(b, f, false)
			// 		}
			// 	}
		}
	}
}
