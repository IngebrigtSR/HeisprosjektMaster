package main

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	"../fsm"
	"../network/bcast"
	"../network/networkmanager"
	"../network/peers"
	"../orderhandler"
)

func main() {
	fmt.Println("Hello World")

	//Peers
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	id := "Something"
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	var newLog orderhandler.ElevLog

	logTx := make(chan orderhandler.ElevLog)
	logRx := make(chan orderhandler.ElevLog)
	go bcast.Transmitter(16569, logTx)
	go bcast.Receiver(16569, logRx)

	p := <-peerUpdateCh
	timer := time.NewTimer(5 * time.Second)
	peerInitDone := false
	for !peerInitDone {
		select{
		case p = <- peerUpdateCh:
		case <- timer.C:
			peerInitDone = true
		}
	}
	if len(p.Peers) == 1 {
		newLog = orderhandler.MakeEmptyLog()
		fmt.Println("No other peers on network. Created a new empty log")
	} else {
		newLog = <- logRx
		fmt.Println("Found other peer(s) on the network! Copied the already existing log")
	}

	//Network
	networkmanager.InitNewElevator(&newLog)
	logIndex := networkmanager.GetLogIndex(newLog, id)
	println("Local index: \t ", logIndex)

	elevio.Init("localhost:15657", NumFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	startUp := make(chan bool)
	logFromFSM := make(chan orderhandler.ElevLog)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	fsm.InitFSM(drv_floors, logIndex, logFromFSM)
	go fsm.ElevFSM(drv_buttons, drv_floors, startUp, logFromFSM)

	// elev := log[logIndex]

	orderhandler.TestCost(newLog)

	transmitter := time.NewTicker(1000 * time.Millisecond)
	//timer := time.NewTimer(5 * time.Second)
	transmit := false
	count := 0
	for {
		select {

		case newLog = <-logRx:
			orderhandler.SetLog(newLog)
			fsm.UpdateButtonLights(newLog)
			startUp <- true
			println("Recieved something")

		case <-transmitter.C:
			if transmit {
				logTx <- orderhandler.GetLog()
				println(count)
				count++
				transmit = false
			}

		case updatedLog := <-logFromFSM:
			if updatedLog != orderhandler.GetLog() {
				transmit = true
			}

			fsm.UpdateButtonLights(updatedLog)

			orderhandler.SetLog(updatedLog)

		case p = <-peerUpdateCh:
			if len(p.Lost) != 0 {
				for i := 0; i < len(p.Lost); i++ {
					fmt.Println(p.Lost[i])
					lostID := p.Lost[i]
					deadElevIndex := networkmanager.GetLogIndex(newLog, lostID)
					newLog = orderhandler.ReAssignOrders(newLog, deadElevIndex)
					newLog[deadElevIndex].State = DEAD
				}
				// Ta over ordrene fra alle heisene som har forsvunnet fra nettverket, og ikke assign nye ordre til disse tapte heisene
			}
			for i := 0; i < len(p.Peers); i++ {
				fmt.Println(p.Peers[i])
			}
			

		}

	}
}
