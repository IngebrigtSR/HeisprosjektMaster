package config

const (
	NumFloors    = 4
	NumButtons   = 3
	NumElevators = 2

	DoorOpenTime = 3  //sec
	ElevTimeout  = 15 //sec

	//Ports
	BcastPort = 16569
	PeerPort  = 15647
)

var (
	LogIndex = 0
)

type State int

const (
	DEAD     State = 0
	INIT           = 1
	IDLE           = 2
	MOVING         = 3
	DOOROPEN       = 4
)

type OrderStatus int

const (
	Unassigned OrderStatus = 0
	Assigned               = 1
	Accepted               = 2
)
