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
	go bcast.Transmitter(BcastPort, logTx)
	go bcast.Receiver(BcastPort, logRx)

	var p peers.PeerUpdate
	id := "Heis numero 1"
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(PeerPort, id, peerTxEnable)
	go peers.Receiver(PeerPort, peerUpdateCh)

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
		fmt.Println("Waiting on log from other peer(s)")
		newLog = <-logRx
		fmt.Println("Found other peer(s) on the network! Copied the already existing log")
	}

	networkmanager.InitNewElevator(&newLog, id)
	LogIndex = networkmanager.GetLogIndex(newLog, id)
	orderhandler.SetLog(newLog)
	println("Local index: \t ", LogIndex)

	//FSM channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	startUp := make(chan bool)
	logFromFSMChan := make(chan orderhandler.ElevLog)
	deadElev := make(chan int)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	fsm.InitFSM(drv_floors, LogIndex, logFromFSMChan)
	logTx <- orderhandler.GetLog()
	time.Sleep(1 * time.Second)
	go fsm.ElevFSM(drv_buttons, drv_floors, startUp, logFromFSMChan, deadElev)

	transmitter := time.NewTicker(100 * time.Millisecond)
	transmit := false
	fsmWatchdog := time.NewTimer(ElevTimeout * time.Second)
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
				println("Broadcasting log")
				logTx <- newLog
				println("Broadcasted log")
				transmit = false
			}

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
						transmit = true
					} else {
						fmt.Println("\n Did not find the lost elevator in the log")
					}
				}
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
					fmt.Println("\n Log index for the new elevator:", newElevIndex)
					newLog[newElevIndex].Online = true
					newLog = orderhandler.UpdateOnlineElevators(newLog)
					orderhandler.SetLog(newLog)
				} else {
					fmt.Println("\n Did not find the new elevator in the log")
				}
				transmit = true

			}

		case newLog = <-logFromFSMChan:
			fsmWatchdog.Reset(ElevTimeout * time.Second)

			if newLog != orderhandler.GetLog() {
				transmit = true
			}

			fsm.UpdateButtonLights(newLog)
			orderhandler.SetLog(newLog)

		case <-fsmWatchdog.C:
			log := orderhandler.GetLog()
			//If all Elevators IDLE timer will never be reset
			if log[LogIndex].State != IDLE {
				log[LogIndex].State = DEAD
				fmt.Println("FSM has crashed")
				transmit = true
			}
		}
	}
}
