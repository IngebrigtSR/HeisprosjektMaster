package fsm

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	"../orderhandler"
)

var activeOrders [NumFloors][NumButtons]bool

func printOrder() {
	fmt.Printf("Active Orders: \n")
	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if activeOrders[f][b] {
				fmt.Printf("%d\t", 1)

			} else {
				fmt.Printf("%d\t", 0)
			}
		}
		fmt.Printf("\n")
	}
}

func shouldStop(elev orderhandler.Elevator, floor int) bool {
	if elev.Floor == 0 || elev.Floor == NumFloors-1 {
		return true
	}
	if elev.Orders[floor][elevio.BT_HallUp] != 0 && elev.Dir == elevio.MD_Up {
		return true
	}
	if elev.Orders[floor][elevio.BT_HallDown] != 0 && elev.Dir == elevio.MD_Down {
		return true
	}
	if elev.Orders[floor][elevio.BT_Cab] != 0 {
		return true
	}
	if !orderhandler.OrdersInFront(elev) {
		if activeOrders[floor][elevio.BT_HallUp] || activeOrders[elev.Floor][elevio.BT_HallDown] {
			return true
		}
	}
	return false
}

func anyActiveOrders() bool {
	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if activeOrders[f][b] {
				return true
			}
		}
	}
	return false
}

func getDir(elev orderhandler.Elevator) elevio.MotorDirection {
	if !anyActiveOrders() {
		return elevio.MD_Stop
	}
	if elev.Dir == elevio.MD_Stop {
		for f := 0; f < NumFloors; f++ {
			for b := 0; b < NumButtons; b++ {
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

func updateButtonLights(log orderhandler.ElevLog) {
	for i := 0; i < NumElevators; i++ {
		for b := 0; b < NumButtons-1; b++ {
			for f := 0; f < NumFloors; f++ {
				if log[i].Orders[f][b] == 2 {
					elevio.SetButtonLamp(elevio.ButtonType(b), f, true)
				}
			}
		}
		if i == LogIndex {
			for f := 0; f < NumFloors; f++ {
				if log[i].Orders[f][2] == 2 {
					elevio.SetButtonLamp(elevio.BT_Cab, f, true)
				}
			}
		}
	}
}

func floorArrival(log orderhandler.ElevLog, floor int) orderhandler.ElevLog {
	fmt.Printf("Floor:\t%+v\n", floor)
	elevio.SetFloorIndicator(floor)
	log[LogIndex].Floor = floor

	if shouldStop(log[LogIndex], floor) {
		log = orderhandler.ClearOrdersFloor(floor, LogIndex, log)
		elevio.SetMotorDirection(elevio.MD_Stop)

		elevio.SetDoorOpenLamp(true)

		doorTimer := time.NewTimer(DoorOpenTime * time.Second)
		<-doorTimer.C

		elevio.SetDoorOpenLamp(false)

	}

	return log
}

func InitFSM(drv_floors chan int) {
	log := orderhandler.GetLog()
	elev := log[LogIndex]
	elevio.SetMotorDirection(elevio.MD_Down)

	floor := <-drv_floors
	elevio.SetFloorIndicator(floor)
	elevio.SetMotorDirection(elevio.MD_Stop)

	elev.State = IDLE
	elev.Dir = elevio.MD_Stop
	elev.Floor = floor

	log[LogIndex] = elev
	orderhandler.SetLog(log)
}

func ElevFSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, startUp chan bool) {

	InitFSM(drv_floors)

	watchdog := time.NewTimer(ElevTimeout * time.Second)

	for {
		select {
		case <-startUp: //Detects if main recieves new log from Network (only needed to get Elev out of IDLE)
			log := orderhandler.GetLog()
			if log[LogIndex].State == IDLE {
				dir := getDir(log[LogIndex])
				elevio.SetMotorDirection(dir)
				log[LogIndex].Dir = dir

				if dir != elevio.MD_Stop {
					log[LogIndex].States = MOVING
				}
			}

		case order := <-drv_buttons:
			log := orderhandler.GetLog()
			fmt.Printf("Order:\t%+v\n", order)

			log = orderhandler.DistributeOrder(order, log)

			dir := getDir(log[LogIndex])
			log[LogIndex].Dir = dir
			elevio.SetMotorDirection(dir)
			watchdog.Reset(ElevTimeout * time.Second)

		case floor := <-drv_floors:
			fmt.Printf("Floor:\t%+v\n", floor)
			elevio.SetFloorIndicator(floor)
			watchdog.Reset(ElevTimeout * time.Second)

			log := orderhandler.GetLog()
			log[LogIndex].Floor = floor

			if shouldStop(log[LogIndex], floor) {
				log = orderhandler.ClearOrdersFloor(floor, LogIndex, log)
				elevio.SetMotorDirection(elevio.MD_Stop)

				elevio.SetDoorOpenLamp(true)
				doorTimer := time.NewTimer(DoorOpenTime * time.Second)
				<-doorTimer.C
				elevio.SetDoorOpenLamp(false)

				dir := getDir(log[LogIndex])
				elevio.SetMotorDirection(dir)
				log[LogIndex].Dir = dir
			}

			orderhandler.SetLog(log)
		}
	}
}

// func ElevFSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, ) {
// 	// elevio.Init("localhost:15657", NumFloors)

// 	var elev orderhandler.Elevator
// 	initFSM()

// 	go elevio.PollButtons(drv_buttons)
// 	go elevio.PollFloorSensor(drv_floors)

// 	for {
// 		select {
// 		case order := <-drv_buttons:
// 			fmt.Printf("Order:\t%+v\n", order)
// 			if !(order.Floor == elev.Floor && elev.Dir == elevio.MD_Stop) {
// 				takeOrder(order.Floor, order.Button)
// 			}
// 			elev.Dir = getDir(elev)
// 			elevio.SetMotorDirection(elev.Dir)

// 		case floor := <-drv_floors:
// 			fmt.Printf("Floor:\t%+v\n", floor)
// 			elevio.SetFloorIndicator(floor)
// 			elev.Floor = floor

// 			if shouldStop(elev) {
// 				clearFloorOrders(floor)
// 				elevio.SetMotorDirection(elevio.MD_Stop)
// 			}
// 			elev.Dir = getDir(elev)
// 			elevio.SetMotorDirection(elev.Dir)

// 		}
// 	}
// }
