package orderhandler

import (
	"fmt"
	"math"

	. "../config"
	"../elevio"
)

//Elevator struct
type Elevator struct {
	Dir    elevio.MotorDirection
	Floor  int
	State  State
	Orders [NumFloors][NumButtons]bool
}

//ElevLog array of all system elevators
type ElevLog [NumElevators]Elevator

func ordersAbove(elev Elevator) bool {
	for f := elev.Floor + 1; 0 <= f && f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[f][b] {
				return true
			}
		}
	}
	return false
}

func ordersBelow(elev Elevator) bool {
	for f := elev.Floor - 1; 0 <= f && f < NumFloors; f-- {
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[f][b] {
				return true
			}
		}
	}
	return false
}

func ordersInFront(elev Elevator) bool {
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

func ordersOnFloor(floor int, elev Elevator) bool {
	for b := 0; b < NumButtons; b++ {
		if elev.Orders[floor][b] {
			return true
		}
	}
	return false
}

func getCost(order elevio.ButtonEvent, elevator Elevator) int {
	//Passing floor = +1
	//Stopping at floor = +2

	elev := elevator //copy of elevator to simulate movement for cost calculation
	cost := 0

	switch S := elev.State; S {
	case DEAD:
		cost = 10000 //Infinity?
	case INIT:
		cost = 10000
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

			if !ordersInFront(elev) {
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

			if !ordersInFront(elev) {
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

func assignOrder(order elevio.ButtonEvent, log ElevLog) {
	cheapestElev := getCheapestElev(order, log)

	log[cheapestElev].Orders[order.Floor][order.Button] = true
}

func reAssignOrders(log ElevLog, deadElev int) {
	if log[deadElev].State != DEAD {
		log[deadElev].State = DEAD
	}
	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if log[deadElev].Orders[f][b] == true {
				order := elevio.ButtonEvent{Floor: f, Button: elevio.ButtonType(b)}
				assignOrder(order, log)
			}
		}
	}
}

func makeEmptyLog() ElevLog {
	var log [NumElevators]Elevator

	for elev := 0; elev < NumElevators; elev++ {
		log[elev].Dir = elevio.MD_Stop
		log[elev].Floor = 0
		log[elev].State = IDLE

		for i := 0; i < NumFloors; i++ {
			for j := 0; j < NumButtons; j++ {
				log[elev].Orders[i][j] = false
			}
		}

	}
	return log
}

func testCostFunction() {
	ElevLog := makeEmptyLog()
	elev := ElevLog[0]

	elev.Dir = elevio.MD_Down
	elev.Floor = 1
	elev.State = MOVING

	elev.Orders[2][2] = true
	elev.Orders[0][2] = true

	testOrder := elevio.ButtonEvent{Floor: 3, Button: elevio.BT_HallUp}

	cost := getCost(testOrder, elev)

	fmt.Println("Elevator cost: \t", cost)

}

// func main() {
// 	testCostFunction()
// }
