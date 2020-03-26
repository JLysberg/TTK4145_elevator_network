package types

import (
	/* LAB setup */
	// "../config"

	/* GOPATH setup */
	"internal/common/config"
)

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
	State        ElevatorState
	Dir          MotorDirection
	LastDir      MotorDirection
	Floor        int
	Queue        [config.MFloors]bool
	OnlineList   [config.NElevs]bool
	ElevLastSent [config.NElevs]int
}

type GlobalInfo struct {
	ID int
	Nodes   [config.NElevs]NodeInfo //Could possibly be exchanged with an array
	Orders  [config.MFloors][config.NElevs]FloorState
}
