package config

const (
	NumFloors    = 4
	NumButtons   = 3
	NumElevators = 3

	DoorTimer  = 3  //sec
	ElevTimout = 10 //sec

	//Div porter for kommunikajson osv

	// LogBCPort = 20000
)

type State int

const (
	DEAD     State = 0
	INIT           = 1
	IDLE           = 2
	MOVING         = 3
	DOOROPEN       = 4
)
