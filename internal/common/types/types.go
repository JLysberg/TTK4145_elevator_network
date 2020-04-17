package types

/* LAB setup */
// "../config"

/* Setup desc. in main */
//"github.com/JLysberg/TTK4145_elevator_network/internal/common/config"

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
	ES_Idle ElevatorState = 0
	ES_Run                = 1
	ES_Stop               = 2
)

type FloorState struct {
	Up    bool
	Down  bool
	Cab   bool
	Clear bool
}

type NodeChannels struct {
	ButtonPress       chan ButtonEvent
	FloorSensor       chan int
	ObstructionSwitch chan bool
	UpdateQueue       chan []FloorState
	LightRefresh      chan GlobalInfo
	ClearOrder        chan int
	DoorOpen          chan bool
}

type LocalInfo struct {
	State   ElevatorState
	Dir     MotorDirection
	LastDir MotorDirection
	Floor   int
	// Queue        [config.MFloors]FloorState
	// OnlineList   [config.NElevs]bool
	// ElevLastSent [config.NElevs]int
}

type GlobalInfo struct {
	ID     int
	Nodes  []LocalInfo
	Orders [][]FloorState
}
