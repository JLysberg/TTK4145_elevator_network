package monitor

import (
	//"fmt"
	"fmt"
	"time"
	"encoding/json"
	"math"

	//"encoding/json"

	/* LAB setup */
	. "../common/types"
	"../common/config"
	"../../pkg/elevio"

	/* GOPATH setup */
	// . "internal/common/config"
	// . "internal/common/types"
)

//type ElevOperation int

//	"pkg/elevio"
//)

var Node = NodeInfo{
	State:   ES_Init,
	Dir:     MD_Stop,
	LastDir: MD_Stop,
	Floor:   0,
}

var Global = GlobalInfo{
	ID: 0,
}

const (
	_UpdateRate = 500 * time.Millisecond
)

func KingOfOrders(btnsPressedLocal <-chan ButtonEvent, newPackets <-chan []byte,
				  newOrderLocal chan<- bool, clearCab <-chan bool) {
	for {
		time.Sleep(_UpdateRate)

		select {
		case btn := <-btnsPressedLocal:
			switch btn.Button {
			case 0: //HallUp
				Global.Orders[btn.Floor][Global.ID].Up = true
			case 1: //HallDown
				Global.Orders[btn.Floor][Global.ID].Down = true
			case 2: //Cab
				Global.Orders[btn.Floor][Global.ID].Cab = true
			}
		case packet := <-newPackets:
			var msg GlobalInfo
			err := json.Unmarshal(packet, &msg)
			if err != nil {
				fmt.Println("Error with unmarshaling message in Monitor:", err)
			}
			newOrders := msg.Orders
			//Is it important to only add new orders?
			if newOrders != Global.Orders {
				for floors := 0; floors < config.MFloors; floors++ {
					for elevs := 0; elevs < config.NElevs; elevs++ {
						if newOrders[floors][elevs].Clear {
							//HER MÅ DET IMPLEMENTERES: En ticker som venter i f.eks. 2 sekunder før vi clearer
							Global.Orders[floors][elevs].Up = false
							Global.Orders[floors][elevs].Down = false
							Global.Orders[floors][elevs].Cab = false
						}
						if newOrders[floors][Global.ID].Cab {
							Global.Orders[floors][Global.ID].Cab = true
						}

						if newOrders[floors][elevs].Up {
							Global.Orders[floors][elevs].Up = true
						}

						if newOrders[floors][elevs].Down {
							Global.Orders[floors][elevs].Down = true
						}
					}
				}
			}
		case <-clearCab:

		default:
			// ---- Cost Estimator ---- //
			// Consider assigning designated thread
			// TODO: Implement support for watchdog elev timeout table

			/* Pre-check for cab orders */
			for floor, floorStates := range Global.Orders {
				if floorStates[Global.ID].Cab && !floorStates[Global.ID].Clear{
					Node.Queue[floor] = true
					newOrderLocal <- true
				}
			}
			/* Cost calculation for non-cab orders */
			for floor, floorStates := range Global.Orders {
				for orderOwner, floorState := range floorStates {
					if floorState.Clear {
						if orderOwner == Global.ID {
							Global.Orders[floor][Global.ID].Cab = false
							Node.Queue[floor] = false
						}
						for elevID := 0; elevID < config.NElevs; elevID++ {
							Global.Orders[floor][elevID].Up = false
							Global.Orders[floor][elevID].Down = false
						}
					} else if floorState.Up || floorState.Down{
						bestCost := int(math.Inf(1))
						bestID := 0
						cost := 0
						for elevID, node := range Global.Nodes {
							floorDiff := int(math.Abs(float64(node.Floor - floor)))

							/*	Calculate floor distance cost
								Distance	Cost
								0			+0
								1			+2
								2			+3
								..
								M 			+(M + 1)
							*/
							if floorDiff != 0 {
								cost += floorDiff + 1
							}

							/*	Calculate pass cost
								Condition	Cost
								Will pass	+0
								Stopped		+1
								Has passed	+5
							*/
							switch node.Dir{
							case MD_Down:
								if floorDiff >= 0 && floorState.Down {
									break
								} else {
									cost += 5
								}
							case MD_Up:
								if floorDiff >= 0 && floorState.Up {
									break
								} else {
									cost += 5
								}
							case MD_Stop:
								cost += 1
							}

							if cost < bestCost {
								bestCost = cost
								bestID = elevID
							}
						}
						if bestID == Global.ID {
							Node.Queue[floor] = true
							newOrderLocal <- true
						}
					}
				}
			}
					//HER MÅ DET IMPLEMENTERES: Noe som sjekker bare heisene som er på nettverket.
				// 	// if Global.Orders[floor][elev].Up{
				// 		upOrder = true	
				// 	}

				// 	if Global.Orders[floor][elev].Down{
				// 		downOrder = true
				// 	}
				
				// 	for elevID, node := range Global.Nodes {
				// 		if  downOrder && node.LastDir == MD_Down {
				// 			if math.Abs(float64(node.Floor-floor)) < float64(floordifferenceDOWN) {
				// 				floordifferenceDOWN = int(math.Abs(float64(node.Floor - floor)))
				// 				bestchoiceDOWN = elevID
				// 			}
				// 		}

				// 		//Only take up-orders if we're going up
				// 		if upOrder && node.LastDir == MD_Up {
				// 			if math.Abs(float64(node.Floor-floor)) < float64(floordifferenceUP) {
				// 				floordifferenceUP = int(math.Abs(float64(node.Floor - floor)))
				// 				bestchoiceUP = elevID
				// 			}
				// 		}
				// 	}
				// 	upOrder, downOrder = false, false
				// 	floordifferenceUP, floordifferenceDOWN = config.MFloors, config.MFloors
				// 	if bestchoiceUP == Global.ID {
				// 		Node.Queue[floor] = true
				// 		newOrderLocal <- true
				// 	}
				// 	if bestchoiceDOWN == Global.ID {
				// 		Node.Queue[floor] = true
				// 		newOrderLocal <- true
				// 	}


			//HER MÅ DET IMPLEMENTERES: En ticker for å sende FSMQueue til FSM.
			//F.eks.: //Lag en kanal som har typen FSMQueue,
			//FSM sjekker ved en select: case hvert 3. millisekund om noe nytt har kommet inn på kanalen

			//Denne nye ordrematrisa skal inn i lightSetter og til Network
		}
	}
}

func lightSetter() {
	for floor := 0; floor < config.MFloors; floor++ {
		for elev := 0; elev < config.NElevs; elev++ {
			if Global.Orders[floor][Global.ID].Cab {
				elevio.SetButtonLamp(2, floor, true)
			}
			if Global.Orders[floor][elev].Up {
				elevio.SetButtonLamp(0, floor, true)
			}
			if Global.Orders[floor][elev].Down {
				elevio.SetButtonLamp(1, floor, true)
			}

			if Global.Orders[floor][elev].Clear {
				elevio.SetButtonLamp(0, floor, false)
				elevio.SetButtonLamp(1, floor, false)
				if elev == Global.ID {
					elevio.SetButtonLamp(2, floor, false)
				}
			}
		}
	}
}
