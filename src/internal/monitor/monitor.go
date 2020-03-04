package monitor

import (
	//"fmt"
	"time"
	"encoding/json"
	"pkg/elevio"
)

type FloorState struct {
	Up    bool
	Down  bool
	Cab   bool
	Clear bool
}

type ElevOperation int

	"pkg/elevio"
)

const (
	_UpdateRate = 20 * time.Millisecond
)

/************************************************************
 * Local elevator information
 ************************************************************/
var Elev_id = 0 //Static ID for single operation

/************************************************************
 * Global network information
 ************************************************************
 OrderMatrix:
	M x N matrix with every FloorState of every floor
	corresponding to every elevator on the network.
	Reads column for column. That is, every FloorState for
	every floor of E1 first, then E2 etc.

	E.g.: E2, F2 is indexed as: OrderMatrix[2*NElevs + 2]

			E0	E1	..	EN
		[
	F0		FS	FS	..	FS
	F1		FS	FS	..	FS
	..		..	..		..
	FM		FS	FS	..	FS
		]

 ElevStates:
	1 x N vector containing every ElevState of every elevator
	on the network.

			E0	E1	..	EN
		[
			ES	ES	..	ES
		]
*/
var OrderMatrix [MFloors * NElevs]FloorState
var ElevStates [NElevs]ElevState

func RemoveGlobalOrder() {
	OrderMatrix[Elev_id*NElevs+ElevStates[Elev_id].Floor].Clear = true
	time.Sleep(500 * time.Millisecond)
	OrderMatrix[Elev_id*NElevs+ElevStates[Elev_id].Floor].Clear = false
}

func PollOrders(sender <-chan elevio.ButtonEvent) {
	for {
		time.Sleep(_Generic_UpdateRate)
		select {
		case b := <-sender:
			switch b.Button {
			case elevio.BT_HallUp:
				OrderMatrix[Elev_id*NElevs+b.Floor].Up = true
			case elevio.BT_HallDown:
				OrderMatrix[Elev_id*NElevs+b.Floor].Down = true
			case elevio.BT_Cab:
				OrderMatrix[Elev_id*NElevs+b.Floor].Cab = true
			}
		
		default:
			for index, floorState := range OrderMatrix {
				var floor = index % MFloors
				var id = 0 //Temporary solution for single elevator operation
				if floorState.Up {
					elevio.SetButtonLamp(elevio.BT_HallUp, floor, true)
				}
				if floorState.Down {
					elevio.SetButtonLamp(elevio.BT_HallDown, floor, true)
				}
				if floorState.Cab && id == Elev_id {
					elevio.SetButtonLamp(elevio.BT_Cab, floor, true)
				}
				if floorState.Clear {
					elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
					elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
					elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
					
					for elev := 0; elev < NElevs; elev++ {
						OrderMatrix[elev*NElevs + floor].Up = false
						OrderMatrix[elev*NElevs + floor].Down = false
					}
					OrderMatrix[Elev_id*NElevs + floor].Cab = false
				}
			}
		}
	}
	//TODO: Implement functionality to poll the network for new orders
}


//Make a struct with what needs to be sent; ElevStates and OrderMatrix (I called it ElevStates_OrderMatrix here)
//Make a function that Marshals and transmits the messages through the network

IncomingMsg := make(chan ElevStates_OrderMatrix)
go bcast.Receiver(0, IncomingMsg)

func PollOrdersNetwork(packet <-chan IncomingMsg) {
	var msg ElevStates_OrderMatrix
	err := json.Unmarshal(packet, &msg)
	if err != nil {
		fmt.Println("error with unmarshaling message:", err)
	}

	for floors := 0; floors < MFloors; i++ {
		for elevs := 0; elevs < NElevs; j++ {
			if elevs == Elev_id && msg.OrderMatrix[floors][elevs].Clear{
			
			OrderMatrix[floors][elevs].Up = false
			OrderMatrix[floors][elevs].Down = false
			OrderMatrix[floors][elevs].Cab = false
							
			}
			if elevs == Elev_id && msg.OrderMatrix[floors][elevs].Cab{
				OrderMatrix[floors][elevs].Cab = true				
			}	
				
			if elevs == Elev_id && msg.OrderMatrix[floors][elevs].Up{ 
				OrderMatrix[floors][elevs].Up = true
			}
	
			if elevs == Elev_id && msg.OrderMatrix[floors][elevs].Down) {
				OrderMatrix[floors][elevs].Down = true
			}
		}
	}
}
