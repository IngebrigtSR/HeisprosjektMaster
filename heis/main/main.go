package main

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	"../fsm"
	"../network/bcast"
	"../network/localip"
	"../network/networkmanager"
	"../network/peers"
	"../orderhandler"
)

func main() {
	fmt.Println("Hello World")

	//Peers
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	localIP, _ := localip.LocalIP()
	go peers.Transmitter(15647, localIP, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	var newLog orderhandler.ElevLog

	logTx := make(chan orderhandler.ElevLog)
	logRx := make(chan orderhandler.ElevLog)
	go bcast.Transmitter(16569, logTx)
	go bcast.Receiver(16569, logRx)
	p := <-peerUpdateCh
	if len(p.Peers) == 1 {
		newLog = orderhandler.MakeEmptyLog()
	} else {
		newLog = <-logRx
	}

	//Network
	networkmanager.InitNewElevator(&newLog)
	localIndex := networkmanager.GetLogIndex(newLog, localIP)
	println("Local index: \t ", localIndex)

	elevio.Init("localhost:15657", NumFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	startUp := make(chan bool)
	logFromFSM := make(chan orderhandler.ElevLog)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	fsm.InitFSM(drv_floors, localIndex, logFromFSM)
	go fsm.ElevFSM(drv_buttons, drv_floors, startUp, logFromFSM)

	// elev := log[localIndex]

	orderhandler.TestCost(newLog)

	transmitter := time.NewTicker(1000 * time.Millisecond)
	timer := time.NewTimer(5 * time.Second)
	transmit := false
	count := 0
	for {
		select {

		case <-timer.C:
			println("walla walla bing bang")

		case <-transmitter.C:
			if transmit {
				count++
				println("transmitted: \t", count)
			}
		case updatedLog := <-logFromFSM:
			orderhandler.SetLog(updatedLog)
			if updatedLog != orderhandler.GetLog() {
				transmit = true
			}

		case p := <-peerUpdateCh:
			if len(p.Lost) != 0 {
				for i := 0; i < len(p.Lost); i++ {
					lostID := p.Lost[i]
					deadElevIndex := networkmanager.GetLogIndex(newLog, lostID)
					newLog = orderhandler.ReAssignOrders(newLog, deadElevIndex)
					newLog[deadElevIndex].State = DEAD
				}
				// Ta over ordrene fra alle heisene som har forsvunnet fra nettverket, og ikke assign nye ordre til disse tapte heisene
			}
		case newLog = <-logRx:

		}

	}
}
