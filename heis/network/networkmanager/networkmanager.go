package networkmanager

import (
<<<<<<< HEAD
	"../localip"
	"../../orderhandler"
	"../../elevio"
	"fmt"
=======
>>>>>>> f003f64784ac5d30bbf9b87e30c70d0d97ec11d4
	. "../../config"
	"../../elevio"
	"../../orderhandler"
	"../localip"
)

<<<<<<< HEAD
func InitNewElevator(logPtr* orderhandler.ElevLog){
	elev := orderhandler.Elevator{}
	ip, err := localip.LocalIP()
	if err != nil{
		fmt.Println(err)
=======
//InitNewElevator initializes a new elevator on the network
func InitNewElevator(logPtr *orderhandler.ElevLog) {
	elev := Elevator{}
	ip, err = localip.LocalIp()
	if err != nil {
		fmt.println(err)
>>>>>>> f003f64784ac5d30bbf9b87e30c70d0d97ec11d4
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

<<<<<<< HEAD
func UpdateLog(logChan chan orderhandler.ElevLog, log* orderhandler.ElevLog) {
	(*log) = <- logChan
}
=======
//UpdateLog updates the log
func UpdateLog(logChan chan orderhandler.ElevLog, log *orderhandler.ElevLog) {
	*log <- logChan
}
>>>>>>> f003f64784ac5d30bbf9b87e30c70d0d97ec11d4
