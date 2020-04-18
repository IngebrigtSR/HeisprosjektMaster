package orderhandler

import (
	"fmt"
	"math"

	. "../config"
	"../elevio"
)

//Elevator struct
type Elevator struct {
	Id     string
	Dir    elevio.MotorDirection
	Floor  int
	State  State
	Orders [NumFloors][NumButtons]OrderStatus
	Active bool
}

//ElevLog array of all system elevators
type ElevLog [NumElevators]Elevator

var localLog ElevLog

//GetLog return the current locally stored log
func GetLog() ElevLog {
	return localLog
}

//SetLog updates locally stored log with newLog
func SetLog(newLog ElevLog) {
	localLog = newLog
}

//ordersAbove checks if there are any orders above the elevators current position
func ordersAbove(elev Elevator) bool {
	for f := elev.Floor + 1; 0 <= f && f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[f][b] != Unassigned {
				return true
			}
		}
	}
	return false
}

//ordersBelow checks if there are any accepted orders below the elevator's current position
func ordersBelow(elev Elevator) bool {
	for f := elev.Floor - 1; 0 <= f && f < NumFloors; f-- {
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[f][b] != Unassigned {
				return true
			}
		}
	}
	return false
}

//OrdersInFront checks for any accepted orders in the direction of travel for a given elevator
func OrdersInFront(elev Elevator) bool {
	switch dir := elev.Dir; dir {
	case elevio.MD_Stop:
		return false
	case elevio.MD_Up:
		return ordersAbove(elev)
	case elevio.MD_Down:
		return ordersBelow(elev)
	}
	return false
}

//OrdersOnFloor checks if an elevator has accepted orders on a given floor
func OrdersOnFloor(floor int, elev Elevator) bool {

	cabOrder := (elev.Orders[floor][int(elevio.BT_Cab)] == Accepted)

	switch d := elev.Dir; d {
	case elevio.MD_Down:
		return elev.Orders[floor][int(elevio.BT_HallDown)] == Accepted || cabOrder

	case elevio.MD_Up:
		return elev.Orders[floor][int(elevio.BT_HallUp)] == Accepted || cabOrder

	case elevio.MD_Stop:
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[floor][b] == Accepted {
				return true
			}
		}
	}
	return false
}

func getCost(order elevio.ButtonEvent, elevator Elevator) int {

	elev := elevator //copy of elevator to simulate movement
	cost := 0        //Init value for cost

	switch S := elev.State; S {
	case DEAD:
		cost = math.MaxInt32
	case INIT:
		cost = math.MaxInt32
	case IDLE:
		cost = int(math.Abs(float64(elev.Floor - order.Floor))) //#floors between new order and elevator
	default:
		for elev.Floor != order.Floor {
			if OrdersOnFloor(elev.Floor, elev) {
				cost += 2
			} else {
				cost++
			}
			if !OrdersInFront(elev) {
				if elev.Dir == elevio.MD_Down && elev.Floor < order.Floor {
					elev.Dir = elevio.MotorDirection(-int(elev.Dir))
				}
				if elev.Dir == elevio.MD_Up && elev.Floor > order.Floor {
					elev.Dir = elevio.MotorDirection(-int(elev.Dir))
				}
			}
			if 0 == elev.Floor && elev.Dir == elevio.MD_Down || elev.Floor == NumFloors && elev.Dir == elevio.MD_Up {
				elev.Dir = elevio.MotorDirection(-int(elev.Dir))
			}
			elev.Floor += int(elev.Dir)
		}

		if elev.State == DOOROPEN {
			cost++
		}
	}
	return cost
}

//getCheapestElev returns the most suited elevator for an order
func getCheapestElev(order elevio.ButtonEvent, log ElevLog) int {
	cheapestElev := -1
	cheapestCost := 10000
	for elev := 0; elev < NumElevators; elev++ {
		cost := getCost(order, log[elev])
		if cost < cheapestCost && log[elev].State != DEAD && log[elev].Active {
			cheapestElev = elev
			cheapestCost = cost
		}
	}
	return cheapestElev
}

//DistributeOrder assigns a given order to the "closest" elevator
func DistributeOrder(order elevio.ButtonEvent, log ElevLog) ElevLog {

	if order.Button == elevio.BT_Cab {
		log[LogIndex].Orders[order.Floor][2] = Accepted
	} else {
		cheapestElev := getCheapestElev(order, log)
		if cheapestElev == -1 {
			println("No Elevators alive to take Order")
		} else if cheapestElev == LogIndex {
			log[cheapestElev].Orders[order.Floor][order.Button] = Accepted
		} else {
			log[cheapestElev].Orders[order.Floor][order.Button] = Assigned
		}
	}
	return log
}

//ReAssignOrders reassigns hall orders from a Dead elevator to the others
func ReAssignOrders(log ElevLog, deadElev int) ElevLog {
	if log[deadElev].State == DEAD || !log[deadElev].Active {
		for f := 0; f < NumFloors; f++ {
			for b := 0; b < NumButtons-1; b++ {
				if log[deadElev].Orders[f][b] != Unassigned {
					order := elevio.ButtonEvent{Floor: f, Button: elevio.ButtonType(b)}
					log = DistributeOrder(order, log)
					log[deadElev].Orders[f][b] = Unassigned
				}
			}
		}
	}
	return log
}

//AcceptOrders goes through the log and looks for orders assigned to the local Elevator and accepts them
func AcceptOrders(log ElevLog) (ElevLog, bool) {
	accepted := false
	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if log[LogIndex].Orders[f][b] == Assigned {
				log[LogIndex].Orders[f][b] = Accepted
				accepted = true
			}
		}
	}
	return log, accepted
}

//ClearOrdersFloor clears orders on a given floor with regards to direction
func ClearOrdersFloor(floor int, elevID int, log ElevLog) ElevLog {
	elev := log[elevID]

	//clear cab order
	log[elevID].Orders[floor][int(elevio.BT_Cab)] = Unassigned

	//clear hall orders
	if !OrdersInFront(elev) {
		for b := 0; b < NumButtons; b++ {
			log[elevID].Orders[floor][b] = Unassigned
		}
	} else {
		if elev.Dir == elevio.MD_Up {
			log[elevID].Orders[floor][int(elevio.BT_HallUp)] = Unassigned
		} else if elev.Dir == elevio.MD_Down {
			log[elevID].Orders[floor][int(elevio.BT_HallDown)] = Unassigned
		}
	}

	return log
}

//DetectDead checks a log for any dead elevators and returns the dead index
func DetectDead(log ElevLog) int {
	for i := 0; i < NumElevators; i++ {
		if log[i].State == DEAD {
			return i
		}
	}
	return -1
}

//MakeEmptyLog creates an empty ElevLog
func MakeEmptyLog() ElevLog {
	var log [NumElevators]Elevator

	for elev := 0; elev < NumElevators; elev++ {
		log[elev].Dir = elevio.MD_Stop
		log[elev].Floor = -1
		log[elev].State = DEAD
		log[elev].Id = ""
		log[elev].Active = false

		for i := 0; i < NumFloors; i++ {
			for j := 0; j < NumButtons; j++ {
				log[elev].Orders[i][j] = Unassigned
			}
		}

	}
	return log
}

//PrintOrders prints a given elevators Orders to terminal
func PrintOrders(elevIndex int, log ElevLog) {
	for i := 0; i < NumButtons; i++ {
		for j := 0; j < NumFloors; j++ {
			fmt.Print(int(log[elevIndex].Orders[j][i]), "\t")
		}
		println()
	}
	println()
}

//PrintElev print a given elevator to terminal
func PrintElev(elev Elevator) {
	println("Elevator:\t", elev.Id)
	println("Direction: \t", elev.Dir)
	println("State: \t", elev.State)
	println("Floor: \t", elev.Floor)
}

//TestCost tests the cost function
func TestCost(log ElevLog) {
	elev := log[0]

	elev.Dir = elevio.MD_Up
	elev.Floor = 1
	elev.State = MOVING

	elev.Orders[2][1] = 2
	elev.Orders[0][2] = 2

	testOrder := elevio.ButtonEvent{Floor: 3, Button: elevio.BT_Cab}

	//cost1 := oldCost(testOrder, elev)
	cost2 := getCost(testOrder, elev)

	//fmt.Println("Old cost fun: \t", cost1)
	fmt.Println("New cost fun: \t", cost2)

}
