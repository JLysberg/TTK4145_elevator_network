package monitor

import (
	//"fmt"
	"fmt"
	"time"

	//"encoding/json"

	/* LAB setup */
	// . "../common/types"
	// "../common/config"
	// "../../pkg/elevio"

	/* GOPATH setup */
	"encoding/json"

	. "internal/common/config"
	. "internal/common/types"
	//"pkg/elevio"
)

//type ElevOperation int

//	"pkg/elevio"
//)

var Node = types.NodeInfo{
	State:   ES_Init,
	Dir:     MD_Stop,
	LastDir: MD_Stop,
	Floor:   0,
	ID:      0,
}

var Global = types.GlobalInfo{
	LocalID: 0,
}

const (
	_UpdateRate = 20 * time.Millisecond
)

func KingOfOrders(btnsPressedLocal <-chan ButtonPress, newPackets <-chan PacketReceiver) {
	select {
	case btn := <-btnsPressedLocal:
		switch btn.Button {
		case 0: //HallUp
			types.GlobalInfo.Orders[btn.Floor][id].Up = true
		case 1: //HallDown
			types.GlobalInfo.Orders[btn.Floor][id].Down = true
		case 2: //Cab
			types.GlobalInfo.Orders[btn.Floor][id].Cab = true
		}
	case packet := <-newPackets:
		var msg types.GlobalInfo
		err := json.Unmarshal(packet, &msg)
		if err != nil {
			fmt.Println("Error with unmarshaling message in Monitor:", err)
		}
		newOrders := msg.Orders
		//Is it important to only add new orders?
		if newOrders != types.GlobalInfo.Orders {
			for floors := 0; floors < config.MFloors; floors++ {
				for elevs := 0; elevs < config.NElevs; elevs++ {
					if newOrders[floors][elevs].Clear {
						//HER MÅ DET IMPLEMENTERES: En ticker som venter i f.eks. 2 sekunder før vi clearer
						types.GlobalInfo.Orders[floors][elevs].Up = false
						types.GlobalInfo.Orders[floors][elevs].Down = false
						types.GlobalInfo.Orders[floors][elevs].Cab = false
					}
					if newOrders[floors][id].Cab {
						types.GlobalInfo.Orders[floors][id].Cab = true
					}

					if newOrders[floors][elevs].Up {
						types.GlobalInfo.Orders[floors][elevs].Up = true
					}

					if newOrders[floors][elevs].Down {
						types.GlobalInfo.Orders[floors][elevs].Down = true
					}
				}
			}
		}
	}

	// ---- Cost Estimator ---- //

	var floordifferenceUP = config.MFloors
	var floordifferenceDOWN = config.MFloors

	var bestchoiceDOWN = id
	var bestchoiceUP = id

	//If we're offline, only check own column in matrix

	for floor = 0; floor < config.MFloors; floor++ {
		for elev := 0; elev < config.NElevs; elev++ {
			for _, NodeInfo := range types.GlobalInfo.Orders {

				//HER MÅ DET IMPLEMENTERES: Noe som sjekker bare heisene som er på nettverket.

				if types.GlobalInfo.Orders[floor][elev].Cab && elev == id {
					FSMQueue = append(FSMQueue, floor)
				}

				///Only take down-orders if we're going down
				if types.GlobalInfo.Orders[floor][elev].Down && types.NodeInfo.LastDir == -1 {
					if abs(types.NodeInfo.Floor-floor) < floordifferenceDOWN {
						floordifferenceDOWN = types.NodeInfo.Floor - floor
						bestchoiceDOWN = elev
					}
				}

				//Only take up-orders if we're going up
				if types.GlobalInfo.Orders[floor][elev].Up && types.NodeInfo.LastDir == 1 {
					if abs(types.NodeInfo.Floor-floor) < floordifferenceUP {
						floordifferenceUP = types.NodeInfo.Floor - floor
						bestchoiceUP = elev
					}
				}
			}
			floordifferenceUP, floordifferenceDOWN = config.MFloors, config.MFloors
			if bestchoiceUP == id {
				FSMQueue = append(FSMQueue, floor)
			}
			if bestchoiceDOWN == id {
				FSMQueue = append(FSMQueue, floor)
			}
		}
	}

	//HER MÅ DET IMPLEMENTERES: En ticker for å sende FSMQueue til FSM.
	//F.eks.: //Lag en kanal som har typen FSMQueue,
	//FSM sjekker ved en select: case hvert 3. millisekund om noe nytt har kommet inn på kanalen

	//Denne nye ordrematrisa skal inn i lightSetter og til Network
}

func lightSetter(id int, sensor chan FloorSensor) {
	select {
	case thisfloor := <-sensor:
		SetFloorIndicator(thisfloor)

		//HER MÅ DET IMPLEMENTERES: Funksjonalitet for å skru AV floorindicator på alle andre etasjer enn thisfloor
	}

	for floor := 0; floor < config.MFloors; floor++ {
		for elev := 0; elev < config.NElevs; elev++ {
			if types.GlobalInfo.Orders[floor][id].Cab {
				elevio.SetButtonLamp(2, floor, true)
			}
			if types.GlobalInfo.Orders[floors][elev].Up {
				elevio.SetButtonLamp(0, floor, true)
			}
			if types.GlobalInfo.Orders[floors][elev].Down {
				elevio.SetButtonLamp(1, floor, true)
			}

			if types.GlobalInfo.Orders[floors][elev].Clear {
				elevio.SetButtonLamp(0, floor, false)
				elevio.SetButtonLamp(1, floor, false)
				if elevs == id {
					elevio.SetButtonLamp(2, floor, false)
				}
			}
		}
	}
}
