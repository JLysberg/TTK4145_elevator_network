package monitor

import (
	"time"

	"pkg/elevio"
)

type FloorState struct {
	Up    bool
	Down  bool
	Cab   bool
	Clear bool
}

type ElevOperation int

const (
	EO_Up       ElevOperation = 1
	EO_Down                   = -1
	EO_Idle                   = 0
	EO_OpenDoor               = 2
)

type ElevState struct {
	Floor     int
	Operation ElevOperation
}

const (
	MFloors = 4
	NElevs  = 1

	_Button_UpdateRate = 20 * time.Millisecond
)

/************************************************************
 * Local elevator information
 ************************************************************/
var elev_floor int
var elev_id = 0 //Static ID for single operation

/************************************************************
 * Global network information
 ************************************************************
 OrderMatrix:
	M x N matrix with every FloorState of every floor
	corresponding to every elevator on the network.
	Reads column for column. That is, every FloorState for
	every floor of E1 first, then E2 etc.

	E.g.: E3, F2 is indexed as: OrderMatrix[(3-1)*NElevs + (2-1)]

			E1	E2	..	EN
		[
	F1		FS	FS	..	FS
	F2		FS	FS	..	FS
	..		..	..		..
	FM		FS	FS	..	FS
		]

 ElevStates:
	1 x N vector containing every ElevState of every elevator
	on the network.

			E1	E1	..	EN
		[
			ES	ES	..	ES
		]
*/
var OrderMatrix [MFloors * NElevs]FloorState
var ElevStates [NElevs]ElevState

func AddLocalOrder(buttonEvent elevio.ButtonEvent) {
	switch buttonEvent.Button {
	case elevio.BT_HallUp:
		OrderMatrix[(elev_id-1)*NElevs+(buttonEvent.Floor-1)].Up = true
	case elevio.BT_HallDown:
		OrderMatrix[(elev_id-1)*NElevs+(buttonEvent.Floor-1)].Down = true
	case elevio.BT_Cab:
		OrderMatrix[(elev_id-1)*NElevs+(buttonEvent.Floor-1)].Cab = true
	}
}

func UpdateButtonLights() {
	for {
		time.Sleep(_Button_UpdateRate)
		for index, floorState := range OrderMatrix {
			var floor = index / NElevs
			if floorState.Up {
				elevio.SetButtonLamp(elevio.BT_HallUp, floor+1, true)
			}
			if floorState.Down {
				elevio.SetButtonLamp(elevio.BT_HallDown, floor+1, true)
			}
			if floorState.Cab && floor == elev_id {
				elevio.SetButtonLamp(elevio.BT_Cab, floor+1, true)
			}
		}
	}
}

//TODO: Implement function that can poll the network for new orders
