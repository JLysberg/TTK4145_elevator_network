package types

import "../../../pkg/network/peers"

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
	ES_Error			  = 3
)

type FloorState struct {
	Up    bool
	Down  bool
	Cab   bool
	Clear bool
}

type NetworkChannels struct {
	MsgTransmitter chan GlobalInfo
	MsgReceiver    chan GlobalInfo
	PeerUpdate     chan peers.PeerUpdate
	PeerTxEnable   chan bool
	UpdateOrders   chan GlobalInfo
	OnlineElevators chan []bool
}

type NodeChannels struct {
	ButtonPress       chan ButtonEvent
	FloorSensor       chan int
	ObstructionSwitch chan bool
	UpdateQueue       chan []FloorState
	LightRefresh      chan GlobalInfo
	SetClearBit       chan int
	ClearQueue		  chan int
	DoorClose          chan bool
	UpdateLocal       chan LocalInfo
}

type LocalInfo struct {
	State   ElevatorState
	Dir     MotorDirection
	LastDir MotorDirection
	Floor   int
}

type GlobalInfo struct {
	ID     int
	Nodes  []LocalInfo
	Orders [][]FloorState
}
