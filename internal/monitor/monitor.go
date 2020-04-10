package monitor

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
)

func clearTimeout(floor int) {
	Global.Orders[floor][Global.ID].Clear = true
	timeout := time.NewTimer(config.ClearTimeout)
	<-timeout.C
	Global.Orders[floor][Global.ID].Clear = false
}

var Local LocalInfo
var Global GlobalInfo

/*	CostEstimator is a goroutine which continuously assigns orders from 
	the global order matrix to any node, taking multiple factors into account.
	All active orders are always assigned to the elevator with the least cost.
	Consequently ensures the queue in local node memory is always up to date.
	
	Cost = distance cost + state cost:
		Distance		Cost
		0				+0
		1				+2
		2				+3
		..
		M 				+(M + 1)
	
		State			Cost
		Will pass		+0
		Stopped			+1
		Has passed		+5
		NOTE: (Has passed includes case of passing order in opposite direction) */
func CostEstimator(newOrderLocal chan<- int) {
	for {
		time.Sleep(config.UpdateRate)
		/*	Always assign cab orders to local node */
		for floor, floorStates := range Global.Orders {
			if floorStates[Global.ID].Cab && !floorStates[Global.ID].Clear{
				Local.Queue[floor].Cab = true
				newOrderLocal <- floor
			}
		}
		/*	Cost calculation for non-cab orders */
		for floor, floorStates := range Global.Orders {
			for elevID, floorState := range floorStates {
				if floorState.Clear {
					if elevID == Global.ID {
						Local.Queue[floor].Up = false
						Local.Queue[floor].Down = false
						Local.Queue[floor].Cab = false
					}
				} else if floorState.Up || floorState.Down {
					bestCost := int(math.Inf(1))
					bestID := 0
					cost := 0
					for nodeID, node := range Global.Nodes {
						floorDiff := int(math.Abs(float64(node.Floor - floor)))

						/*	Calculate distance cost */
						if floorDiff != 0 {
							cost += floorDiff + 1
						}

						/*	Calculate state cost */
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
							cost++
						}

						if cost < bestCost {
							bestCost = cost
							bestID = nodeID
						}
					}
					/*	Assign order to local node if optimal */
					if bestID == Global.ID && Local.Queue[floor] != floorState {
						Local.Queue[floor] = floorState
						newOrderLocal <- floor
					}
				}
			}
		}
	}
}

func KingOfOrders(btnsPressedLocal <-chan ButtonEvent, newPackets <-chan []byte,
				  refreshButtonLights chan<- int, clearOrderLocal chan int) {
	refreshButtonLights <- -1
	for {
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
			/*	Only update local Global.Orders if it differs from msg.Orders */
			if msg.Orders != Global.Orders {
				for msgFloor, msgFloorStates := range msg.Orders {
					for msgElevID, msgFloorState := range msgFloorStates {
						if !msgFloorState.Clear {
							/*	Concatenate orders from msg into local order matrix */
							Global.Orders[msgFloor][msgElevID].Up =
								Global.Orders[msgFloor][msgElevID].Up || msgFloorState.Up 		
							Global.Orders[msgFloor][msgElevID].Down =
								Global.Orders[msgFloor][msgElevID].Down || msgFloorState.Down
							Global.Orders[msgFloor][msgElevID].Cab =
								Global.Orders[msgFloor][msgElevID].Cab || msgFloorState.Cab
						} else {
							/*	Remove all up/down orders if there is a clear present */
							for elevID := 0; elevID < config.NElevs; elevID++ {
								Global.Orders[msgFloor][elevID].Up = false
								Global.Orders[msgFloor][elevID].Down = false
							}
							/*	Also remove cab order if present */
							Global.Orders[msgFloor][msgElevID].Cab = false
						}
					}
				}
				refreshButtonLights <- -1
			}
		case floor := <-clearOrderLocal:
			go clearTimeout(floor)

			/*	Hack to ensure each elevator is not dependent on a full
				cost estimator run to clear order from queue */
			Local.Queue[floor].Up = false
			Local.Queue[floor].Down = false
			Local.Queue[floor].Cab = false
			/*	The following block might be superfluous when networks are introduced*/
			/********************************************/
			/*	Remove all up/down orders if there is a clear present */
			for elevID := 0; elevID < config.NElevs; elevID++ {
				Global.Orders[floor][elevID].Up = false
				Global.Orders[floor][elevID].Down = false
			}
			/*	Also remove cab order if present */
			Global.Orders[floor][Global.ID].Cab = false
			/*********************************************/

			refreshButtonLights <- floor
		}
	}
}

/*	Lightsetter updates every button light in accordance with the global order 
	matrix on refresh call */
func LightSetter(refresh <-chan int) {
	for {
		floor := <-refresh
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
