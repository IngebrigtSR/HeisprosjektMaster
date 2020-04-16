package fsm

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	"../orderhandler"
)

func printOrder(elev orderhandler.Elevator) {
	fmt.Printf("Active Orders: \n")
	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[f][b] == 2 {
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
		if elev.Orders[floor][elevio.BT_HallUp] == 2 || elev.Orders[elev.Floor][elevio.BT_HallDown] == 2 {
			return true
		}
	}
	return false
}

func anyActiveOrders(elev orderhandler.Elevator) bool {
	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[f][b] == 2 {
				return true
			}
		}
	}
	return false
}

func getDir(elev orderhandler.Elevator) elevio.MotorDirection {
	if !anyActiveOrders(elev) {
		return elevio.MD_Stop
	}

	if elev.Dir == elevio.MD_Stop {
		for f := 0; f < NumFloors; f++ {
			for b := 0; b < NumButtons; b++ {
				if elev.Orders[f][b] == 2 {
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
	//Turns around if there are no orders in front of Elevator
	return elevio.MotorDirection(-elev.Dir)
}

func updateButtonLights(log orderhandler.ElevLog) {

	var lights [NumFloors][NumButtons]bool

	//Hall order lights
	for i := 0; i < NumElevators; i++ {
		for f := 0; f < NumFloors; f++ {
			for b := 0; b < NumButtons-1; b++ {
				if log[i].Orders[f][b] == 2 {
					lights[f][b] = true
				}
			}
		}
	}

	//Cab order lights
	for f := 0; f < NumFloors; f++ {
		if log[LogIndex].Orders[f][2] == 2 {
			lights[f][2] = true
		}
	}

	//Setting all light values
	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if lights[f][b] {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
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

//InitFSM initializes the FSM
func InitFSM(drv_floors chan int, localIndex int, newLogChan chan orderhandler.ElevLog) {
	log := orderhandler.GetLog()
	elev := log[localIndex]
	elevio.SetMotorDirection(elevio.MD_Down)

	floor := <-drv_floors
	elevio.SetFloorIndicator(floor)
	elevio.SetMotorDirection(elevio.MD_Stop)

	elev.State = IDLE
	elev.Dir = elevio.MD_Stop
	elev.Floor = floor

	log[localIndex] = elev
	orderhandler.SetLog(log)

	//newLogChan <- log
}

//ElevFSM handles logic used to execute waiting orders and run the elevator
func ElevFSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, startUp chan bool, newLogChan chan orderhandler.ElevLog) {

	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	watchdog := time.NewTimer(ElevTimeout * time.Second)   //Timer to check for hardware malfunction
	doorTimer := time.NewTimer(DoorOpenTime * time.Second) //Timer for closing door after opening
	doorTimer.Stop()
	for {
		select {

		case order := <-drv_buttons:
			watchdog.Reset(ElevTimeout * time.Second)

			log := orderhandler.GetLog()
			fmt.Printf("Order:\t%+v\n", order)

			//Locally executes any orders on the same floor as Elevator
			if log[LogIndex].Floor == order.Floor && log[LogIndex].State != MOVING {
				if log[LogIndex].State == DOOROPEN {
					doorTimer.Reset(DoorOpenTime * time.Second)
				}
				if log[LogIndex].State == IDLE {
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(DoorOpenTime * time.Second)
					log[LogIndex].State = DOOROPEN
				}
			} else {
				log = orderhandler.DistributeOrder(order, log)

				dir := getDir(log[LogIndex])
				log[LogIndex].Dir = dir

				//To prevent Elev from moving with the door open
				if log[LogIndex].State != DOOROPEN {
					elevio.SetMotorDirection(dir)
				}

				if dir == elevio.MD_Stop {
					log[LogIndex].State = IDLE
				} else {
					log[LogIndex].State = MOVING
				}
			}

			updateButtonLights(log)
			//newLogChan <- log
			orderhandler.SetLog(log)

		case floor := <-drv_floors:
			watchdog.Reset(ElevTimeout * time.Second)

			fmt.Printf("Floor:\t%+v\n", floor)
			elevio.SetFloorIndicator(floor)

			log := orderhandler.GetLog()
			log[LogIndex].Floor = floor

			if shouldStop(log[LogIndex], floor) {
				log = orderhandler.ClearOrdersFloor(floor, LogIndex, log)
				elevio.SetMotorDirection(elevio.MD_Stop)

				elevio.SetDoorOpenLamp(true)
				doorTimer.Reset(DoorOpenTime * time.Second)
				log[LogIndex].State = DOOROPEN
			}

			updateButtonLights(log)
			//newLogChan <- log
			orderhandler.SetLog(log)

		case <-doorTimer.C:
			log := orderhandler.GetLog()

			elevio.SetDoorOpenLamp(false)
			dir := getDir(log[LogIndex])
			log[LogIndex].Dir = dir
			elevio.SetMotorDirection(dir)
			if dir != elevio.MD_Stop {
				log[LogIndex].State = MOVING
			} else {
				log[LogIndex].State = IDLE
			}
			updateButtonLights(log)
			//newLogChan <- log
			orderhandler.SetLog(log)

		case <-startUp: //Detects if main recieves new log from Network (only needed to get Elev out of IDLE)
			log := orderhandler.GetLog()
			updateButtonLights(log)

			if log[LogIndex].State == IDLE {

				dir := getDir(log[LogIndex])
				elevio.SetMotorDirection(dir)
				log[LogIndex].Dir = dir

				if dir != elevio.MD_Stop {
					log[LogIndex].State = MOVING
				}
			}
			//newLogChan <- log
			orderhandler.SetLog(log)

		case <-watchdog.C:
			log := orderhandler.GetLog()
			if log[LogIndex].State == IDLE {
				watchdog.Reset(ElevTimeout * time.Second)
			} else {
				log[LogIndex].State = DEAD
			}
			//newLogChan <- log
			orderhandler.SetLog(log)

		case <-drv_obstr:
			orderhandler.PrintOrders(0, orderhandler.GetLog())

		case <-drv_stop:
			orderhandler.PrintElev(orderhandler.GetLog()[0])
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
