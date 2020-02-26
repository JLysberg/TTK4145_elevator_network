package monitor

type FloorState struct {
	Up   bool
	Down bool
	Cab  bool
}

type ElevOperation int

const (
	EO_Up       ElevOperation = 1
	EO_Down                   = -1
	EO_Idle                   = 0
	EO_OpenDoor               = 1
)

type ElevState struct {
	ID        int
	Floor     int
	Operation ElevOperation
}

const (
	NumFloors = 4
	NumElevs  = 1
)

var elev_floor int
var elev_id int
var elevStates [NumElevs]ElevState
var OrderMatrix [NumFloors * NumElevs]FloorState

func GetOrderMatrix() [NumFloors * NumElevs]FloorState {
	return orderMatrix
}
