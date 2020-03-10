package monitor

import (
	//"fmt"
	"time"
	//"encoding/json"

	/* LAB setup */
	// . "../common/types"
	// "../common/config"
	// "../../pkg/elevio"

	/* GOPATH setup */
	//"internal/common/config"
	. "internal/common/types"
)

var Node = NodeInfo{
	State:   ES_Init,
	Dir:     MD_Stop,
	LastDir: MD_Stop,
	Floor:   0,
	ID:      0,
}

var Global = GlobalInfo{
	LocalID: 0,
}

const (
	_UpdateRate = 20 * time.Millisecond
)

/*func KingOfOrders(btnsPressedLocal chan buttonPress, newPackets chan packetReceiver,
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

			if msg.Orders != OrdersLocal{
				for floors := 0; floors < MFloors; floors++ {
					for elevs := 0; elevs < NElevs; elevs++

						if msg.Orders[floors][elevs].Clear
							//HER MÅ DET IMPLEMENTERES: En ticker som venter i f.eks. 2 sekunder før vi clearer
							OrdersLocal[floors][elevs].Up = false
							OrdersLocal[floors][elevs].Down = false
							OrdersLocal[floors][elevs].Cab = false
						}
						if msg.Orders[floors][elevs].Cab && elevs == id{
							OrdersLocal[floors][elevs].Cab = true
						}

						if msg.Orders[floors][elevs].Up{
							OrdersLocal[floors][elevs].Up = true
						}

						if msg.Orders[floors][elevs].Down{
							OrdersLocal[floors][key].Down = true
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

	for floor := 0; floor < MFloors; floor++ {
		for elev := 0; elev < NElevs; elev++ {
			for _, NodeInfo := range GlobalInfo.Nodes {

				//HER MÅ DET IMPLEMENTERES: Noe som sjekker bare heisene som er på nettverket.

				if OrdersLocal[floor][elev].Cab && elev == id{
					FSMQueue = append(FSMQueue, floor)
				}

				///Only take down-orders if we're going down/standing still
				if OrdersLocal[floor][elev].Down && LastDir == -1{
					if abs(NodeInfo.Floor - floor) < floordifferenceDOWN{
						floordifferenceDOWN = NodeInfo.Floor - floor
						bestchoiceDOWN = elev
					}
				}

				//Only take up-orders if we're going up/standing still
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

			//HER MÅ DET IMPLEMENTERES: Funksjonalitet for å skru AV floorindicator på alle andre etasjer
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

func setLastDirection(i int){
	last_dir = i
}
*/