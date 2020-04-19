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

	elevio.Init("localhost:15657", NumFloors)
	var newLog orderhandler.ElevLog

	//Network & Peers
	logTx := make(chan orderhandler.ElevLog)
	logRx := make(chan orderhandler.ElevLog)
	go bcast.Transmitter(16569, logTx)
	go bcast.Receiver(16569, logRx)

	var p peers.PeerUpdate
	id := "Something"
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	timer := time.NewTimer(5 * time.Second)
	peerInitDone := false
	for !peerInitDone {
		select {
		case p = <-peerUpdateCh:
		case <-timer.C:
			peerInitDone = true
		}
	}

	if len(p.Peers) == 1 {
		newLog = orderhandler.MakeEmptyLog()
		fmt.Println("No other peers on network. Created a new empty log")
	} else {
		newLog = <-logRx
		fmt.Println("Found other peer(s) on the network! Copied the already existing log")
	}

	networkmanager.InitNewElevator(&newLog, id)
	LogIndex = networkmanager.GetLogIndex(newLog, id)
	orderhandler.SetLog(newLog)
	println("Local index: \t ", LogIndex)

	//FSM
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	startUp := make(chan bool)
	logFromFSMChan := make(chan orderhandler.ElevLog)
	deadElev := make(chan int)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	fsm.InitFSM(drv_floors, LogIndex, logFromFSMChan)
	logTx <- orderhandler.GetLog()
	go fsm.ElevFSM(drv_buttons, drv_floors, startUp, logFromFSMChan, deadElev)

	transmitter := time.NewTicker(10 * time.Millisecond)
	//timer := time.NewTimer(5 * time.Second)
	transmit := false
	println("Initialization completed")
	for {
		select {

		case newLog = <-logRx:

			newLog, accepted := orderhandler.AcceptOrders(newLog)
			orderhandler.SetLog(newLog)
			if accepted {
				transmit = true
			}

			fsm.UpdateButtonLights(newLog)
			startUp <- true
			println("Recieved new log")

		case <-transmitter.C:
			if transmit {
				logTx <- newLog
				println("Broadcasted log")
				transmit = false
			}

		case newLog = <-logFromFSMChan:

			if newLog != orderhandler.GetLog() {
				transmit = true
			}

			fsm.UpdateButtonLights(newLog)

			orderhandler.SetLog(newLog)

		case p = <-peerUpdateCh:

			if len(p.Lost) != 0 {
				fmt.Print("\n LOST:")
				for i := 0; i < len(p.Lost); i++ {
					fmt.Print("\t", p.Lost[i])
					lostID := p.Lost[i]
					lostElevIndex := networkmanager.GetLogIndex(newLog, lostID)
					if lostElevIndex != -1 && lostElevIndex != LogIndex {
						fmt.Println("Log index for the lost elevator:", lostElevIndex)
						newLog[lostElevIndex].Online = false
						newLog = orderhandler.ReAssignOrders(newLog, lostElevIndex)
						orderhandler.SetLog(newLog)
					} else {
						fmt.Println("Did not find the lost elevator in the log")
					}
				}
				transmit = true
				// Ta over ordrene fra alle heisene som har forsvunnet fra nettverket, og ikke assign nye ordre til disse tapte heisene
			}
			fmt.Print("\n PEERS:")
			for i := 0; i < len(p.Peers); i++ {
				fmt.Print("\t", p.Peers[i])
			}
			fmt.Print("\n IDS:")
			for i := 0; i < NumElevators; i++ {
				fmt.Print("\t", newLog[i].Id)
			}

			if len(p.New) != 0 {
				fmt.Print("\n NEW:", p.New)

				newID := p.New
				newElevIndex := networkmanager.GetLogIndex(newLog, newID)
				if newElevIndex != -1 {
					fmt.Println("Log index for the new elevator:", newElevIndex)
					newLog[newElevIndex].Online = true
					orderhandler.SetLog(newLog)
				} else {
					fmt.Println("Did not find the new elevator in the log")
				}
				transmit = true

			}

		case dead := <-deadElev:
			if dead != -1 {
				newLog = orderhandler.GetLog()
				newLog = orderhandler.ReAssignOrders(newLog, dead)
				if newLog != orderhandler.GetLog() {
					transmit = true
				}
			}
		}
	}
}
