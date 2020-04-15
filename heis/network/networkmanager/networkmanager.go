package networkmanager

import (
	. "../../config"
	"../../elevio"
	"../../orderhandler"
	"../localip"
)

//InitNewElevator initializes a new elevator on the network
func InitNewElevator(logPtr *orderhandler.ElevLog) {
	elev := Elevator{}
	ip, err = localip.LocalIp()
	if err != nil {
		fmt.println(err)
	}
	elev.Id = ip
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
func GetLogIndex(log orderhandler.ElevLog, ip string) int {
	index := 0
	for log[index].Id != ip {
		index++
	}
	return index
}

//UpdateLog updates the log
func UpdateLog(logChan chan orderhandler.ElevLog, log *orderhandler.ElevLog) {
	*log <- logChan
}
