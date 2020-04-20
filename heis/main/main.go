package main

import (
	"fmt"
	"time"

	. "../config"
	"../elevio"
	"../fsm"
	"../network/bcast"
	"../logmanager"
	"../network/peers"
	"../orderhandler"
)

func main() {
	fmt.Println("Hello World")

	elevio.Init("localhost:15657", NumFloors)

	//Network & Peers
	logTx := make(chan logmanager.ElevLog)
	logRx := make(chan logmanager.ElevLog)
	go bcast.Transmitter(BcastPort, logTx)
	go bcast.Receiver(BcastPort, logRx)

	var p peers.PeerUpdate
	id := "Heis numero 1"

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(PeerPort, id, peerTxEnable)
	go peers.Receiver(PeerPort, peerUpdateCh)

	//Init log
	newLog := logmanager.InitLog(peerUpdateCh, logRx)
	logmanager.InitNewElevator(&newLog, id)

	LogIndex = logmanager.GetLogIndex(newLog, id)
	logmanager.SetLog(newLog)
	println("Local index: \t ", LogIndex)

	//FSM channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	startUp := make(chan bool)
	logFromFSMChan := make(chan logmanager.ElevLog)
	deadElev := make(chan int)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	fsm.InitFSM(drv_floors, LogIndex, logFromFSMChan)
	logTx <- logmanager.GetLog()
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
			logmanager.SetLog(newLog)
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
				fmt.Print("\n LOST:\t")
				for i := 0; i < len(p.Lost); i++ {
					fmt.Print("\t", p.Lost[i])
					lostID := p.Lost[i]
					lostElevIndex := logmanager.GetLogIndex(newLog, lostID)
					if lostElevIndex != -1 && lostElevIndex != LogIndex {
						fmt.Println("\nLog index for the lost elevator:", lostElevIndex)
						newLog[lostElevIndex].Online = false
						newLog = orderhandler.ReAssignOrders(newLog, lostElevIndex)
						logmanager.SetLog(newLog)
						transmit = true
					} else {
						fmt.Println("\n Did not find the lost elevator in the log")
					}
				}
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
				newElevIndex := logmanager.GetLogIndex(newLog, newID)
				if newElevIndex != -1 {
					fmt.Println("\n Log index for the new elevator:", newElevIndex)
					newLog[newElevIndex].Online = true
					newLog = logmanager.UpdateOnlineElevators(newLog)
					logmanager.SetLog(newLog)
				} else {
					fmt.Println("\n Did not find the new elevator in the log")
				}
				transmit = true

			}

		case newLog = <-logFromFSMChan:
			fsmWatchdog.Reset(ElevTimeout * time.Second)

			if newLog != logmanager.GetLog() {
				transmit = true
			}

			fsm.UpdateButtonLights(newLog)
			logmanager.SetLog(newLog)

		case <-fsmWatchdog.C:
			log := logmanager.GetLog()
			//If all Elevators are IDLE timer will never be reset
			if log[LogIndex].State != IDLE {
				log[LogIndex].State = DEAD
				fmt.Println("FSM has crashed")
				transmit = true
			}
		}
	}
}
