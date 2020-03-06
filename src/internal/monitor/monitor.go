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

			E1	E2  ..	EN
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


func KingOfOrders(btnsPressedLocal chan buttonPress, newPackets chan packetReceiver, id string, OrdersLocal [][]FloorState){

	select{
		case btn := <- btnsPressedLocal
			switch btn.Button {
				case 0: //HallUp
					OrdersLocal[btn.Floor][id].Up = true
				case 1: //HallDown
					OrdersLocal[btn.Floor][id].Down = true
				case 2: //Cab
					OrdersLocal[btn.Floor][id].Cab = true
			}
	
		case packet := <- newPackets
			var msg GlobalInfo
			err := json.Unmarshal(packet, &msg)
			if err != nil {
				fmt.Println("error with unmarshaling message:", err)
			}

			if msg.Orders != OrdersLocal{  //if GlobalInfo order matrix != Local order matrix:
				for floors := 0; floors < MFloors; i++ {
					for key, value := range msg.Nodes {
						
						if msg.Orders[floors][key].Clear
							//VENT I 2 SEKUNDER FØR VI CLEARER ORDREN!
							OrdersLocal[floors][value.ID].Up = false
							OrdersLocal[floors][key].Down = false
							OrdersLocal[floors][key].Cab = false
										
						}
						if msg.Orders[floors][key].Cab{
							OrdersLocal[floors][key].Cab = true				
						}	
							
						if msg.Orders[floors][key].Up{ 
							OrdersLocal[floors][key].Up = true
						}
				
						if msg.Orders[floors][key].Down{
							OrdersLocal[floors][key].Down = true
						}
					}
				}
			}
		}
	}
	//HER MÅ VI:

	//kjør kostfunksjon på LOKAL ordrematrise
	//send inn queue te fsm:
		//lag en ticker før å send te FSM!!
	//default 
	floordifference := MFloors


	for floor := 0; floor < MFloors; floor++ {
		for key, value := range msg.Nodes{ //key: IP	 value: NodeInfo struct
			
			//if det e nån som ikje e på nettverke:
			if OrdersLocal[floor][value.ID].Down{
				if value.Floor - 
			}
			if OrdersLocal[floor][value.ID].Up{
				if value.Floor - 
			}
			if OrdersLocal[floor][value.ID].Cab{ //Vårres heis har en cab order -> den må tas
				
				if value.Floor - 
			} 
		}
	}

	//HAN E GLOBAL  -  BEHØV IKJE Å SEND
	//Send ut ny, oppdatert ordrematrise tebake te network
	//send den tel func lightSetter()
}




func lightSetter(id string, sensor chan floorsensor){
	select{
		case thisfloor := <- sensor
			SetFloorIndicator(thisfloor)
			for floor := 0; floor < MFloors; floor++ {
				if floor != thisfloor{
					//Skru av floorindicator på alle andre etasja
					//SetFloorIndicator(thisfloor)
				}
			}
		}
	}

	for floor := 0; floor < MFloors; floor++ {
		for elev := range OrdersLocal[floor] {
			if OrdersLocal[floor][elev].Cab && elev == id{
				elevio.SetButtonLamp(2, floor, true)
			}
			if OrdersLocal[floors][elev].Up{
				elevio.SetButtonLamp(0, floor, true)
			}
			if OrdersLocal[floors][elev].Down{
				elevio.SetButtonLamp(1, floor, true)
			}

			if OrdersLocal[floors][elev].Clear{
				elevio.SetButtonLamp(0, floor, false)
				elevio.SetButtonLamp(1, floor, false)
				if elevs == id{
					elevio.SetButtonLamp(2, floor, false)
				}
			}

		//for floors
			//for elevators
				//if cab button e blitt trykt OG det e våres egen heis (ikje ta andres cab lys)
					//sett lyse høyt hos mæ sjøl
			
				//if hall button e blitt trykt (hvis en ener e blitt satt hos up/down på en rad)
					//sett lyse høyt i min etasje der

				//if clear nån steds
				//skru av mitt lys i den etasjn

		}
	}
}