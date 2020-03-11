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
	eS_Init ElevatorState = 0
	eS_Idle               = 1
	eS_Run                = 2
	eS_Stop               = 3
)

type FloorState struct {
	Up    bool
	Down  bool
	Cab   bool
	Clear bool
}

type NodeInfo struct {
	IP      string
	id      int
	State   ElevatorState
	Dir     MotorDirection
	LastDir int
	Floor   int
	Queue   []bool
	//OrdersLocal [][]FloorState
}

type GlobalInfo struct {
	LocalID      int
	LocalIP      string
	OnlineList   [NElevs]bool
	ElevLastSent [NElevs]int
	Nodes        map[int]NodeInfo
	Orders       [][]FloorState
}
