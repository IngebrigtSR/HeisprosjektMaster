package fsm

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	"../orderhandler"
	"../logmanager"
)

func shouldStop(elev logmanager.Elevator, floor int) bool {
	if elev.Floor == 0 || elev.Floor == NumFloors-1 {
		return true
	}
	if orderhandler.OrdersOnFloor(floor, elev) {
		return true
	}
	if !orderhandler.OrdersInFront(elev) {
		if elev.Orders[floor][elevio.BT_HallUp] == Accepted || elev.Orders[elev.Floor][elevio.BT_HallDown] == Accepted {
			return true
		}
	}
	return false
}

func anyActiveOrders(elev logmanager.Elevator) bool {
	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[f][b] == Accepted {
				return true
			}
		}
	}
	return false
}

func getDir(elev logmanager.Elevator) elevio.MotorDirection {
	if !anyActiveOrders(elev) {
		return elevio.MD_Stop
	}

	if elev.Dir == elevio.MD_Stop {
		for f := 0; f < NumFloors; f++ {
			for b := 0; b < NumButtons; b++ {
				if elev.Orders[f][b] == Accepted {
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

//UpdateButtonLights sets all Elevator button lights on/off depending on accepted orders
func UpdateButtonLights(log logmanager.ElevLog) {

	var lights [NumFloors][NumButtons]bool

	//Hall order lights
	for i := 0; i < NumElevators; i++ {
		for f := 0; f < NumFloors; f++ {
			for b := 0; b < NumButtons-1; b++ {
				if log[i].Orders[f][b] == Accepted {
					lights[f][b] = true
				}
			}
		}
	}

	//Cab order lights
	for f := 0; f < NumFloors; f++ {
		if log[LogIndex].Orders[f][2] == Accepted {
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

//InitFSM initializes the FSM
func InitFSM(drv_floors chan int, localIndex int, newLogChan chan logmanager.ElevLog) {
	log := logmanager.GetLog()
	elev := log[localIndex]
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(elevio.MD_Down)

	floor := <-drv_floors
	elevio.SetFloorIndicator(floor)
	elevio.SetMotorDirection(elevio.MD_Stop)

	elev.State = IDLE
	elev.Dir = elevio.MD_Stop
	elev.Floor = floor

	log[localIndex] = elev
	logmanager.SetLog(log)
}

//ElevFSM handles logic used run the elevator and handle events like buttonpresses and floorpassing
func ElevFSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, startUp chan bool, newLogChan chan logmanager.ElevLog, deadElev chan int) {

	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	motorTimer := time.NewTimer(ElevTimeout * time.Second) //Timer to check for motor malfunction
	doorTimer := time.NewTimer(DoorOpenTime * time.Second) //Timer for closing the door after opening
	doorTimer.Stop()
	for {
		select {

		case <-startUp: //Detects if main recieves new log from Network
			log := logmanager.GetLog()
			dead := logmanager.DetectDead(log)

			if dead != -1 {
				log = orderhandler.ReAssignOrders(log, dead)
			}

			println(log[LogIndex].Floor)
			if orderhandler.OrdersOnFloor(log[LogIndex].Floor, log[LogIndex]) {
				if log[LogIndex].State != MOVING {
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorTimer.Reset(DoorOpenTime * time.Second)
					log[LogIndex].State = DOOROPEN
					elevio.SetDoorOpenLamp(true)
					log = orderhandler.ClearOrdersFloor(log[LogIndex].Floor, LogIndex, log)
				}
			} else if log[LogIndex].State == IDLE {
				dir := getDir(log[LogIndex])
				log[LogIndex].Dir = dir

				if dir != elevio.MD_Stop {
					log[LogIndex].State = MOVING
					motorTimer.Reset(ElevTimeout * time.Second)
				}
				elevio.SetMotorDirection(dir)
			}
			newLogChan <- log
			println("Fetched log from network")

		case order := <-drv_buttons:
			//watchdog.Reset(ElevTimeout * time.Second)
			log := logmanager.GetLog()
			fmt.Printf("Order:\t%+v\n", order)

			log = orderhandler.DistributeOrder(order, log)
			dir := getDir(log[LogIndex])
			log[LogIndex].Dir = dir

			//To prevent Elev from moving with the door open or dead hardware
			if log[LogIndex].State != DOOROPEN && log[LogIndex].State != DEAD {
				if dir == elevio.MD_Stop {
					log[LogIndex].State = IDLE
				} else {
					log[LogIndex].State = MOVING
					motorTimer.Reset(ElevTimeout * time.Second)
				}
				elevio.SetMotorDirection(dir)
			}

			newLogChan <- log

		case floor := <-drv_floors:
			motorTimer.Reset(ElevTimeout * time.Second)

			fmt.Printf("Floor:\t%+v\n", floor)
			elevio.SetFloorIndicator(floor)

			log := logmanager.GetLog()
			log[LogIndex].Floor = floor

			if shouldStop(log[LogIndex], floor) {
				log = orderhandler.ClearOrdersFloor(floor, LogIndex, log)
				elevio.SetMotorDirection(elevio.MD_Stop)

				doorTimer.Reset(DoorOpenTime * time.Second)
				log[LogIndex].State = DOOROPEN
				elevio.SetDoorOpenLamp(true)

			}

			newLogChan <- log

		case <-doorTimer.C:
			log := logmanager.GetLog()
			dir := getDir(log[LogIndex])

			elevio.SetDoorOpenLamp(false)

			log[LogIndex].Dir = dir
			if dir == elevio.MD_Stop {
				log[LogIndex].State = IDLE
			} else {
				log[LogIndex].State = MOVING
				motorTimer.Reset(ElevTimeout * time.Second)
			}

			elevio.SetMotorDirection(dir)

			newLogChan <- log

		case <-motorTimer.C:
			log := logmanager.GetLog()
			if log[LogIndex].State == IDLE || log[LogIndex].State == DOOROPEN {
				motorTimer.Reset(ElevTimeout * time.Second)
			} else {
				log[LogIndex].State = DEAD

				println("Motor Failure")
			}

			newLogChan <- log

		case <-drv_obstr:
			orderhandler.PrintOrders(LogIndex, logmanager.GetLog())

		case <-drv_stop:
			orderhandler.PrintElev(logmanager.GetLog()[LogIndex])
		}
	}
}
