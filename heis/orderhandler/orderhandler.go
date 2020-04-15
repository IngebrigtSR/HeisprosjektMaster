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
	Orders [NumFloors][NumButtons]int
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
			if elev.Orders[f][b] != 0 {
				return true
			}
		}
	}
	return false
}

//ordersBelow checks if there are any orders below the elevators current position
func ordersBelow(elev Elevator) bool {
	for f := elev.Floor - 1; 0 <= f && f < NumFloors; f-- {
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[f][b] != 0 {
				return true
			}
		}
	}
	return false
}

//OrdersInFront checks for any active orders in the direction of travel for a given elevator
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

//Checks if an elevator has orders on a given floor, used to for calculations in cost function
func ordersOnFloor(floor int, elev Elevator) bool {

	switch d := elev.Dir; d {
	case elevio.MD_Down:
		return elev.Orders[floor][int(elevio.BT_HallDown)] != 0

	case elevio.MD_Up:
		return elev.Orders[floor][int(elevio.BT_HallUp)] != 0

	case elevio.MD_Stop:
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[floor][b] != 0 {
				return true
			}
		}
	}
	return false
}

func getCost(order elevio.ButtonEvent, elevator Elevator) int {

	elev := elevator //copy of elevator to simulate movement for cost calculation
	cost := 0        //Init value for cost

	switch S := elev.State; S {
	case DEAD:
		cost = math.MaxInt32 //Se kommentar over getCost
	case INIT:
		cost = math.MaxInt32
	case IDLE:
		cost = int(math.Abs(float64(elev.Floor - order.Floor))) //#floors between new order and elevator
	default:
		startFloor := elev.Floor
		println(startFloor)

		for elev.Floor != order.Floor {
			if ordersOnFloor(elev.Floor, elev) {
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
	}
	return cost
}

func oldCost(order elevio.ButtonEvent, elevator Elevator) int {
	//Passing floor = +1
	//Stopping at floor = +2

	elev := elevator //copy of elevator to simulate movement for cost calculation
	cost := 0        //Init value for cost

	//
	switch S := elev.State; S {
	case DEAD:
		cost = math.MaxInt32
	case INIT:
		cost = math.MaxInt32
	case IDLE:
		cost = int(math.Abs(float64(elev.Floor - order.Floor))) //#floors between new order and elevator
	default:
		startFloor := elev.Floor
		println(startFloor)

		//Cost of orders and floors in direction of travel
		for 0 <= elev.Floor && elev.Floor < NumFloors {
			if elev.Floor == order.Floor {
				return cost
			} else if ordersOnFloor(elev.Floor, elev) {
				cost += 2
			} else {
				cost++
			}

			if !OrdersInFront(elev) {
				break
			}

			elev.Floor += int(elev.Dir)
		}
		//Adding cost of traveling back to "start"
		cost += int(math.Abs(float64(startFloor - elev.Floor)))

		//Turn elevator around and move to next floor
		elev.Dir = elevio.MotorDirection(-int(elev.Dir))
		elev.Floor = startFloor + int(elev.Dir)

		//Cost of orders and floors in oppsite direction
		for 0 <= elev.Floor && elev.Floor < NumFloors {
			println(elev.Floor)
			if elev.Floor == order.Floor {
				return cost
			} else if ordersOnFloor(elev.Floor, elev) {
				cost += 2
			} else {
				cost++
			}

			if !OrdersInFront(elev) {
				break
			}

			elev.Floor += int(elev.Dir)
		}

		//Adding 1 if elevator is currently executing an order
		if elev.State == DOOROPEN {
			cost++
		}
	}
	return cost
}

func getCheapestElev(order elevio.ButtonEvent, log ElevLog) int {
	cheapestElev := -1
	cheapestCost := 10000
	for elev := 0; elev < NumElevators; elev++ {
		cost := getCost(order, log[elev])
		if cost < cheapestCost && log[elev].State != DEAD {
			cheapestElev = elev
			cheapestCost = cost
		}
	}
	return cheapestElev
}

//AllElevatorsDead checks if all the elevators are dead
func AllElevatorsDead(log ElevLog) bool {
	for elev := 0; elev < NumElevators; elev++ {
		if log[elev].State != DEAD {
			return false
		}
	}
	return true
}

//DistributeOrder assigns a given order to the "closest" elevator
func DistributeOrder(order elevio.ButtonEvent, log ElevLog) ElevLog {

	cheapestElev := getCheapestElev(order, log)
	if cheapestElev == LogIndex {
		log[cheapestElev].Orders[order.Floor][order.Button] = 2
	} else {
		log[cheapestElev].Orders[order.Floor][order.Button] = 1
	}
	return log
}

//ReAssignOrders reassigns dead elevator orders to other elevators
func ReAssignOrders(log ElevLog, deadElev int) ElevLog {
	if log[deadElev].State == DEAD {
		for f := 0; f < NumFloors; f++ {
			for b := 0; b < NumButtons; b++ {
				if log[deadElev].Orders[f][b] != 0 {
					order := elevio.ButtonEvent{Floor: f, Button: elevio.ButtonType(b)}
					log = DistributeOrder(order, log)
				}
			}
		}
	}
	return log
}

//AcceptOrders goes through the log and looks for orders assigned to the local Elevator and accepts them
func AcceptOrders(log ElevLog) ElevLog {
	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if log[LogIndex].Orders[f][b] == 1 {
				log[LogIndex].Orders[f][b] = 2
			}
		}
	}
	return log
}

//ClearOrdersFloor clears all orders on a given floor for a given elevator
func ClearOrdersFloor(floor int, elevID int, log ElevLog) ElevLog {
	elev := log[elevID]

	switch S := elev.State; S {
	case IDLE:
		for b := 0; b < NumButtons; b++ {
			elev.Orders[floor][b] = 0
		}
	case MOVING:
		if elev.Orders[floor][elevio.BT_Cab] != 0 {
			elev.Orders[floor][elevio.BT_Cab] = 0
		}
		if elev.Orders[floor][elevio.BT_HallUp] != 0 && elev.Dir == elevio.MD_Up {
			elev.Orders[floor][elevio.BT_HallUp] = 0
		}
		if elev.Orders[floor][elevio.BT_HallDown] != 0 && elev.Dir == elevio.MD_Down {
			elev.Orders[floor][elevio.BT_HallDown] = 0
		}
	}

	log[elevID] = elev
	return log
}

//MakeEmptyLog creates an empty ElevLog
func MakeEmptyLog() ElevLog {
	var log [NumElevators]Elevator

	for elev := 0; elev < NumElevators; elev++ {
		log[elev].Dir = elevio.MD_Stop
		log[elev].Floor = -1
		log[elev].State = DEAD
		log[elev].Id = ""

		for i := 0; i < NumFloors; i++ {
			for j := 0; j < NumButtons; j++ {
				log[elev].Orders[i][j] = 0
			}
		}

	}
	return log
}

func PrintOrders(elevIndex int, log ElevLog) {
	for i := 0; i < NumButtons; i++ {
		for j := 0; j < NumFloors; j++ {
			fmt.Print(log[elevIndex].Orders[j][i], "\t")
		}
		println()
	}
	println()
}

func TestCost(log ElevLog) {
	elev := log[0]

	elev.Dir = elevio.MD_Up
	elev.Floor = 1
	elev.State = MOVING

	elev.Orders[2][1] = 2
	//elev.Orders[0][2] = 2

	testOrder := elevio.ButtonEvent{Floor: 3, Button: elevio.BT_Cab}

	cost1 := oldCost(testOrder, elev)
	cost2 := getCost(testOrder, elev)

	fmt.Println("Old cost fun: \t", cost1)
	fmt.Println("New cost fun: \t", cost2)

}

// func main() {
// 	log := MakeEmptyLog()
// 	TestCost(log)
// }
