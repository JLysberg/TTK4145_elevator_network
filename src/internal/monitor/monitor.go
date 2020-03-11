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


func KingOfOrders(btnsPressedLocal chan buttonPress, newPackets chan packetReceiver, 
	id int, LastDir int, OrdersLocal [][]FloorState, FSMQueue []bool){

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
				fmt.Println("Error with unmarshaling message:", err)
			}
			newOrders := msg.Orders
			//Is it important to only add new orders?
			if newOrders != Orders{
				for floors := 0; floors < MFloors; floors++ {
					for elevs := 0; elevs < NElevs; elevs++
						
						if newOrders[floors][elevs].Clear
							//HER MÅ DET IMPLEMENTERES: En ticker som venter i f.eks. 2 sekunder før vi clearer
							Orders[floors][elevs].Up = false
							Orders[floors][elevs].Down = false
							Orders[floors][elevs].Cab = false			
						}
						if newOrders[floors][id].Cab{ //&& elevs == id{
							Orders[floors][id].Cab = true				
						}	
							
						if newOrders[floors][elevs].Up{ 
							Orders[floors][elevs].Up = true
						}
				
						if newOrders[floors][elevs].Down{
							Orders[floors][elevs].Down = true
						}
					}
				}
			}
		}
	}

	
	// ---- Cost Estimator ---- //

	floordifferenceUP := MFloors
	floordifferenceDOWN := MFloors

	bestchoiceDOWN := id
	bestchoiceUP := id

	//If we're offline, only check own column in matrix


	for floor := 0; floor < MFloors; floor++ {
		for elev := 0; elev < NElevs; elev++ { 
			for _, NodeInfo := range GlobalInfo.Nodes {
				
				//HER MÅ DET IMPLEMENTERES: Noe som sjekker bare heisene som er på nettverket.

				if OrdersLocal[floor][elev].Cab && elev == id{
					FSMQueue = append(FSMQueue, floor)
				}

				///Only take down-orders if we're going down
				if OrdersLocal[floor][elev].Down && LastDir == -1{ 
					if abs(NodeInfo.Floor - floor) < floordifferenceDOWN{
						floordifferenceDOWN = NodeInfo.Floor - floor
						bestchoiceDOWN = elev
					} 
				}

				//Only take up-orders if we're going up
				if OrdersLocal[floor][elev].Up && LastDir == 1{ 
					if abs(NodeInfo.Floor - floor) < floordifferenceUP{
						floordifferenceUP = NodeInfo.Floor - floor
						bestchoiceUP = elev
					}
				}
			} 
			floordifferenceUP, floordifferenceDOWN = MFloors, MFloors
			
			if bestchoiceUP == id{
				FSMQueue = append(FSMQueue, floor)
			}
			if bestchoiceDOWN == id{
				FSMQueue = append(FSMQueue, floor)
			}
		}
	}


	//HER MÅ DET IMPLEMENTERES: En ticker for å sende FSMQueue til FSM. 
		//F.eks.: //Lag en kanal som har typen FSMQueue, 
				//FSM sjekker ved en select: case hvert 3. millisekund om noe nytt har kommet inn på kanalen  


	//Denne nye ordrematrisa skal inn i lightSetter og til Network
}




func lightSetter(id string, sensor chan floorsensor){
	select{
		case thisfloor := <- sensor
			SetFloorIndicator(thisfloor)

			//HER MÅ DET IMPLEMENTERES: Funksjonalitet for å skru AV floorindicator på alle andre etasjer enn thisfloor
		}
	}

	for floor := 0; floor < MFloors; floor++ {
		for elev := 0; elev < NElevs; elev++
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
		}
	}
}