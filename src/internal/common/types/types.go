package types

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type ElevatorState int

const (
	ES_Init ElevatorState = 0
	ES_Idle               = 1
	ES_Run                = 2
	ES_Stop               = 3
)

type FloorState struct {
	Up    bool
	Down  bool
	Cab   bool
	Clear bool
}

type NodeInfo struct {
	//ID    string
	State ElevatorState
	Dir   MotorDirection
	Floor int
	Queue []bool
}

type GlobalInfo struct {
	LocalID string
	Nodes   map[string]NodeInfo //Could possibly be exchanged with an array
	Orders  [][]FloorState
}
