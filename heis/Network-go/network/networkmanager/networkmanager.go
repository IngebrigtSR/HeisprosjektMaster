package networkmanager

import (
	"../bcast"
	"../conn"
	"../localip"
	"../peers"
	"../../../orderhandler"
	"../../../fsm"
	"../../../elevio"
	"fmt"
)

func InitNewElevator(id int, logPtr* orderhandler.ElevLog){
	elev := Elevator{}
	ip, err = localip.LocalIp()
	if err != nil{
		fmt.println(err)
	}
	elev.Id = ip
	elev.Dir = elevio.MD_Stop
	elev.State = INIT
	*logPtr = append(*logPtr, elev)
	fsm.InitFSM()
}

func GetLocalIndex(log orderhandler.ElevLog) int {
	index = 0
	for log[index].Id != localip.LocalIp() {
		index++
	}
	return index
}