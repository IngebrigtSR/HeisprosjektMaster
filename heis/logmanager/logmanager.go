package logmanager

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	"../network/peers"
)

//Elevator struct
type Elevator struct {
	Id     string
	Dir    elevio.MotorDirection
	Floor  int
	State  State
	Orders [NumFloors][NumButtons]OrderStatus
	Online bool
}

//ElevLog array of all system elevators
type ElevLog [NumElevators]Elevator

var localLog ElevLog

//InitLog initializes the log either recieved from potensial peers or make a new empty log
func InitLog(peerUpdate chan peers.PeerUpdate, logRx chan ElevLog) ElevLog {
	timer := time.NewTimer(5 * time.Second)
	peerInitDone := false
	var p peers.PeerUpdate
	for !peerInitDone {
		select {
		case p = <-peerUpdate:
		case <-timer.C:
			peerInitDone = true
		}
	}

	var newLog ElevLog
	if len(p.Peers) == 1 {
		newLog = MakeEmptyLog()
		fmt.Println("No other peers on network. Created a new empty log")
	} else {
		fmt.Println("Waiting on log from other peer(s)")
		newLog = <-logRx
		fmt.Println("Found other peer(s) on the network! Copied the already existing log")
	}
	return newLog
}

//InitNewElevator initializes a new elevator in the log
func InitNewElevator(logPtr *ElevLog, id string) {
	elev := Elevator{}
	elev.Id = id
	elev.Dir = elevio.MD_Stop
	elev.Floor = -1
	elev.State = INIT
	elev.Online = true
	for i := 0; i < NumElevators; i++ {
		if (*logPtr)[i].Id == "" && (*logPtr)[i].State == DEAD {
			(*logPtr)[i] = elev
			return
		}
	}
}

//GetLog returns the current locally stored log
func GetLog() ElevLog {
	return localLog
}

//SetLog updates locally stored log with newLog
func SetLog(newLog ElevLog) {
	localLog = newLog
}

//GetLogIndex returns the log index of the local elevator
func GetLogIndex(log ElevLog, id string) int {
	for index := 0; index < len(log); index++ {
		if log[index].Id == id {
			return index
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
		log[elev].Online = false

		for i := 0; i < NumFloors; i++ {
			for j := 0; j < NumButtons; j++ {
				log[elev].Orders[i][j] = Unassigned
			}
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

//UpdateOnlineElevators checks if new elevators has come online and updates the log with them
func UpdateOnlineElevators(newLog ElevLog) ElevLog {
	log := GetLog()
	for elev := 0; elev < NumElevators; elev++ {
		if newLog[elev].Online && !log[elev].Online {
			log[elev] = newLog[elev]
		}
	}
	return log
}

//AcceptOrders accepts orders assigned by external elevators (returns true if any are accepted)
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

//UpdateLog updates the local log when a new one is recieved from the network
func UpdateLog(newLog ElevLog) (ElevLog, bool) {
	log := GetLog()
	newLog, changed := AcceptOrders(newLog)

	for i := 0; i < NumElevators; i++ {
		if i == LogIndex {
			log[i].Orders = newLog[i].Orders
		} else {
			log[i] = newLog[i]
		}
	}

	SetLog(log)

	return log, changed
}
