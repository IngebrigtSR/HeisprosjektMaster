package networkmanager

import (
	. "../../config"
	"../../elevio"
	"../../orderhandler"
)

func InitNewElevator(logPtr *orderhandler.ElevLog) {
	elev := orderhandler.Elevator{}
	id := "Something"
	elev.Id = id
	elev.Dir = elevio.MD_Stop
	elev.Floor = -1
	elev.State = INIT
	for i := 0; i < NumElevators; i++ {
		if (*logPtr)[i].Id == "" && (*logPtr)[i].State == DEAD {
			(*logPtr)[i] = elev
			return
		}
	}
}

//GetLogIndex returns the index of the elevator
func GetLogIndex(log orderhandler.ElevLog, id string) int {
	for index := 0; index < len(log); index++ {
		if log[index].Id == id {
			return index
		}
	}
	return -1
}

func UpdateLog(logChan chan orderhandler.ElevLog, log *orderhandler.ElevLog) {
	(*log) = <-logChan
}
