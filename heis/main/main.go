package main

import (
	"fmt"
	"time"

	"../orderhandler"
	"../network/networkmanager"
	"../network/peers"
	"../network/localip"
	. "../config"
)

func main() {
	fmt.Println("Hello World")

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	const localIP := localip.LocalIP()
	go peers.Transmitter(15647, localIP, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)
	
	var log orderhandler.ElevLog

	logTx := make(chan orderhandler.ElevLog)
	logRx := make(chan orderhandler.ElevLog)
	go bcast.Transmitter(16569, logTx)
	go bcast.Receiver(16569, logRx)
	
	
	p := <- peerUpdateCh
	if p.Peers == "" {
		log = orderhandler.MakeEmptyLog()
	}
	else {
		log = <- logRx
	}
	
	networkmanager.InitNewElevator(&log)
	const localIndex := networkmanager.GetLogIndex(log, localip.LocalIP())




	elevio.Init("localhost:15657", config.NumFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	startUp := make(chan bool)


	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	InitFSM(drv_floors, localIndex)
	go fsm.ElevFSM(drv_buttons, drv_floors, startUp)

	elev := log[localIndex]

	orderhandler.TestCost(log)

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

		case p := <- peerUpdateCh:
			if p.Lost != ""{
				lostId := p.Lost
				deadElevIndex := networkmanager.GetLogIndex(log, lostId)
				deadElev := log[deadElevIndex]
				log = orderhandler.ReAssignOrders(log, deadElev)
				log[deadElevIndex].State = DEAD
				// Ta over ordrene fra alle heisene som har forsvunnet fra nettverket, og ikke assign nye ordre til disse tapte heisene
			}
		case log <- logRx:

		}
		
	}
}
