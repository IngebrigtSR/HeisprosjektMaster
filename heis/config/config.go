package config

const (
	NumFloors    = 4
	NumButtons   = 3
	NumElevators = 3

	DoorOpenTime = 3  //sec
	ElevTimeout  = 20 //sec

	LogIndex = 0

	//Div porter for kommunikajson osv

	// LogBCPort = 20000
)

var ()

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
