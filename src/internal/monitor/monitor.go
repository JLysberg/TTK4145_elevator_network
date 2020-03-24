package monitor

import (
	//"fmt"
	"encoding/json"
	"fmt"
	"math"
	"time"

	//"encoding/json"

	/* LAB setup */
	// . "../common/types"
	// "../common/config"
	// "../../pkg/elevio"

	/* GOPATH setup */
	"internal/common/config"
	. "internal/common/types"
	"pkg/elevio"
)

var Node = NodeInfo{
	State:   ES_Init,
	Dir:     MD_Stop,
	LastDir: MD_Stop,
	Floor:   0,
}

var Global = GlobalInfo{
	ID: 0,
}

func CostEstimator(newOrderLocal chan<- bool) {
	// TODO: Implement support for watchdog elev timeout table
	for {
		time.Sleep(config.CostEstimator_UpdateRate)
		/* Pre-check for cab orders */
		for floor, floorStates := range Global.Orders {
			if floorStates[Global.ID].Cab && !floorStates[Global.ID].Clear{
				Node.Queue[floor] = true
				newOrderLocal <- true
			}
		}
		/* Cost calculation for non-cab orders */
		for floor, floorStates := range Global.Orders {
			for elevID, floorState := range floorStates {
				if floorState.Clear {
					if elevID == Global.ID {
						Node.Queue[floor] = false
					}
				} else if floorState.Up || floorState.Down {
					bestCost := int(math.Inf(1))
					bestID := 0
					cost := 0
					for nodeID, node := range Global.Nodes {
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
							NOTE: (Has passed includes case
									of passing order in
									opposite direction)
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
							bestID = nodeID
						}
					}
					if bestID == Global.ID {
						Node.Queue[floor] = true
						newOrderLocal <- true
					}
				}
			}
		}
	}
}

func clearTimeout(floor int) {
	fmt.Println("Clearing on floor", floor)
	Global.Orders[floor][Global.ID].Clear = true
	timeout := time.NewTimer(config.ClearTimeout)
	<-timeout.C
	Global.Orders[floor][Global.ID].Clear = false
}

func KingOfOrders(btnsPressedLocal <-chan ButtonEvent, newPackets <-chan []byte,
				  refreshButtonLights chan<- int, clearOrderLocal chan int) {
	/* Initially refresh all button lights */
	refreshButtonLights <- -1
	for {
		//TODO: Send order matrix to network
		select {
		case btn := <-btnsPressedLocal:
			switch btn.Button {
			case BT_HallUp:
				Global.Orders[btn.Floor][Global.ID].Up = true
			case BT_HallDown:
				Global.Orders[btn.Floor][Global.ID].Down = true
			case BT_Cab:
				Global.Orders[btn.Floor][Global.ID].Cab = true
			}
			refreshButtonLights <- btn.Floor
		case packet := <-newPackets:
			var msg GlobalInfo
			err := json.Unmarshal(packet, &msg)
			if err != nil {
				fmt.Println("Error with unmarshaling message in Monitor:", err)
			}
			/* Only update local Global.Orders if it differs from msg.Orders */
			if msg.Orders != Global.Orders {
				for msgFloor, msgFloorStates := range msg.Orders {
					for msgElevID, msgFloorState := range msgFloorStates {
						if !msgFloorState.Clear {
							/* Concatenate orders from msg into local order matrix */
							Global.Orders[msgFloor][msgElevID].Up =
								Global.Orders[msgFloor][msgElevID].Up || msgFloorState.Up 		
							Global.Orders[msgFloor][msgElevID].Down =
								Global.Orders[msgFloor][msgElevID].Down || msgFloorState.Down
							Global.Orders[msgFloor][msgElevID].Cab =
								Global.Orders[msgFloor][msgElevID].Cab || msgFloorState.Cab
						} else {
							/* Remove all up/down orders if there is a clear present */
							for elevID := 0; elevID < config.NElevs; elevID++ {
								Global.Orders[msgFloor][elevID].Up = false
								Global.Orders[msgFloor][elevID].Down = false
							}
							/* Also remove cab order if present */
							Global.Orders[msgFloor][msgElevID].Cab = false
						}
					}
				}
				/* Refresh all button lights on new packet reception */
				refreshButtonLights <- -1
			}
		case floor := <-clearOrderLocal:
			go clearTimeout(floor)

			/* The following block might be superfluous when networks are introduced*/
			/********************************************/
			/* Remove all up/down orders if there is a clear present */
			for elevID := 0; elevID < config.NElevs; elevID++ {
				Global.Orders[floor][elevID].Up = false
				Global.Orders[floor][elevID].Down = false
			}
			/* Also remove cab order if present */
			Global.Orders[floor][Global.ID].Cab = false
			/*********************************************/

			refreshButtonLights <- floor
		}
	}
}

/*
Updates every button light in accordance with the global order matrix
on refresh call
*/
func LightSetter(refresh <-chan int) {
	for {
		floor := <-refresh
		fmt.Println("refreshing button lights on floor", floor)
		// Really bad code quality, but this works for now :)
		if floor != -1 {
			for elevID := 0; elevID < config.NElevs; elevID++ {
				for button := BT_HallUp; button <= BT_Cab; button++ {
					lightValue := false
					switch button {
					case BT_HallUp:
						lightValue = Global.Orders[floor][elevID].Up
					case BT_HallDown:
						lightValue = Global.Orders[floor][elevID].Down
					case BT_Cab:
						lightValue = Global.Orders[floor][elevID].Cab &&
							elevID == Global.ID
					}
					elevio.SetButtonLamp(button, floor, lightValue)
				}
			}
		} else {
			for floor := 0; floor < config.MFloors; floor++ {
				for elevID := 0; elevID < config.NElevs; elevID++ {
					for button := BT_HallUp; button <= BT_Cab; button++ {
						lightValue := false
						switch button {
						case BT_HallUp:
							lightValue = Global.Orders[floor][elevID].Up
						case BT_HallDown:
							lightValue = Global.Orders[floor][elevID].Down
						case BT_Cab:
							lightValue = Global.Orders[floor][elevID].Cab &&
								elevID == Global.ID
						}
						elevio.SetButtonLamp(button, floor, lightValue)
					}
				}
			}
		}
	}
}
