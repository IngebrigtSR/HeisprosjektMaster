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
	if len(p.Peers) == 0 {
		newLog = orderhandler.MakeEmptyLog()
	} else {
		newLog = <-logRx
	}

	networkmanager.InitNewElevator(&newLog)
	localIndex := networkmanager.GetLogIndex(newLog, localIP)

	elevio.Init("localhost:15657", NumFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	startUp := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	fsm.InitFSM(drv_floors, localIndex)
	go fsm.ElevFSM(drv_buttons, drv_floors, startUp)

	// elev := log[localIndex]

	orderhandler.TestCost(newLog)

	transmitter := time.NewTicker(1000 * time.Millisecond)
	timer := time.NewTimer(5 * time.Second)
	transmit := true
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

		case p := <-peerUpdateCh:
			if len(p.Lost) != 0 {
				for i := 0; i < len(p.Lost); i++ {
					lostId := p.Lost[i]
					deadElevIndex := networkmanager.GetLogIndex(newLog, lostId)
					newLog = orderhandler.ReAssignOrders(newLog, deadElevIndex)
					newLog[deadElevIndex].State = DEAD
				}
				// Ta over ordrene fra alle heisene som har forsvunnet fra nettverket, og ikke assign nye ordre til disse tapte heisene
			}
		case newLog = <-logRx:

		}

	}
}
