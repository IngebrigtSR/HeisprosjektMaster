package networkmanager

import (
	"../localip"
	"../peers"
	"../../orderhandler"
	"../../elevio"
	"fmt"
)

func InitNewElevator(logPtr* orderhandler.ElevLog){
	elev := Elevator{}
	ip, err = localip.LocalIp()
	if err != nil{
		fmt.println(err)
	}
	elev.Id = ip
	elev.Dir = elevio.MD_Stop
	elev.Floor = -1
	elev.State = INIT
	*logPtr = append(*logPtr, elev)
}

func GetLocalIndex(log orderhandler.ElevLog) int {
	index := 0
	for log[index].Id != localip.LocalIp() {
		index++
	}
	return index
}

func UpdateLog(logChan chan orderhandler.ElevLog, log* orderhandler.ElevLog) {
	*log <- logChan
}