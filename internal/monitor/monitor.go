package monitor

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	/* Setup desc. in main */
	"github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
)

var getQueueCopy = make(chan []FloorState)

func createQueueCopy(queue []FloorState) []FloorState {
	copy := make([]FloorState, len(queue))
	for i, k := range queue {
		copy[i] = k
	}
	return copy
}

func clearTimeout(floor int) {
	Global.Orders[floor][Global.ID].Clear = true
	timeout := time.NewTimer(config.ClearTimeout)
	<-timeout.C
	Global.Orders[floor][Global.ID].Clear = false
}

/*	Queue gives a call to CostEstimator to return a copy of queue */
func Queue() []FloorState {
	return <-getQueueCopy
}

var Global GlobalInfo

/*	CostEstimator is a goroutine which continuously assigns orders from 
	the global order matrix to any node, taking multiple factors into account.
	All active orders are always assigned to the elevator with the least cost.
	The responsibility of CostEstimator is to guarantee that queue 
	is always up to date.
	
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
func CostEstimator(updateQueue chan<- []FloorState) {
	queue := make([]FloorState, config.MFloors)
	for {
		estBegin := time.Now()
		/*	Always assign cab orders to local node */
		for floor, floorStates := range Global.Orders {
			if floorStates[Global.ID].Cab && !floorStates[Global.ID].Clear &&
				!queue[floor].Cab {
				queue[floor].Cab = true
				updateQueue <- createQueueCopy(queue)
			}
		}
		/*	Cost calculation for non-cab orders */
		for floor, floorStates := range Global.Orders {
			for elevID, floorState := range floorStates {
				if floorState.Clear {
					if elevID == Global.ID &&
					   (queue[floor].Up || queue[floor].Down || queue[floor].Cab){
						queue[floor].Up = false
						queue[floor].Down = false
						queue[floor].Cab = false
						updateQueue <- createQueueCopy(queue)
					}
				} else if floorState.Up || floorState.Down {
					bestCost := int(math.Inf(1))
					bestID := 0
					cost := 0
					for nodeID, node := range Global.Nodes {
						/*	Ignore all offline nodes */
						// if !Local.OnlineList[nodeID] {
						// 	continue
						// }

						/*	Calculate distance cost */
						floorDiff := int(math.Abs(float64(node.Floor - floor)))
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
							if floorDiff <= 0 && floorState.Up {
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
					if bestID == Global.ID && queue[floor] != floorState {
						queue[floor] = floorState
						queueCopy := createQueueCopy(queue)
						updateQueue <- queueCopy
					}
				}
			}
		}
		copy := createQueueCopy(queue)
		select {
		case getQueueCopy <- copy:
		default:
		}
		/*	Calculate runtime and sleep if runtime is less than update rate */
		estRuntime := time.Since(estBegin)
		if estRuntime < config.UpdateRate {
			time.Sleep(config.UpdateRate - estRuntime)
		}
	}
}

/*	OrderServer handles all incoming orders. This includes all new local orders 
	as well as incoming network packets. The responsibility of OrderServer is 
	to guarantee that Global.Orders is always up to date with the rest of the network */
func OrderServer(buttonPress <-chan ButtonEvent, newPackets <-chan []byte,
				  lightRefresh chan<- int, clearOrder <-chan int) {
	for {
		select {
		case pressedButton := <-buttonPress:
			switch pressedButton.Button {
			case BT_HallUp:
				Global.Orders[pressedButton.Floor][Global.ID].Up = true
			case BT_HallDown:
				Global.Orders[pressedButton.Floor][Global.ID].Down = true
			case BT_Cab:
				Global.Orders[pressedButton.Floor][Global.ID].Cab = true
			}
			lightRefresh <- pressedButton.Floor
		case receivedPackage := <-newPackets:
			var msg GlobalInfo
			err := json.Unmarshal(receivedPackage, &msg)
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
				lightRefresh <- -1
			}
		case clearFloor := <-clearOrder:
			go clearTimeout(clearFloor)
			lightRefresh <- clearFloor
			/*	The following block might be superfluous when networks are introduced*/
			/********************************************/
			/*	Remove all up/down orders if there is a clear present */
			for elevID := 0; elevID < config.NElevs; elevID++ {
				Global.Orders[clearFloor][elevID].Up = false
				Global.Orders[clearFloor][elevID].Down = false
			}
			/*	Also remove cab order if present */
			Global.Orders[clearFloor][Global.ID].Cab = false
			/*********************************************/
		}
	}
}

/*	LightServer updates every button light in accordance with the global order 
	matrix on refresh call. A refresh call of -1 updates all buttons, and 
	any specific floor call restricts the iteration to said floor. */
func LightServer(lightRefresh <-chan int) {
	for {
		select {
		case callingFloor := <-lightRefresh:
			for floor, floorStates := range Global.Orders {
				/*	Skip most of the iteration if callingFloor is specified */
				if (callingFloor != -1) && (floor != callingFloor) {
					continue
				}
				for elevID, floorState := range floorStates {
					for button := BT_HallUp; button <= BT_Cab; button++ {
						lightValue := false
						switch button {
						case BT_HallUp:
							lightValue = floorState.Up
						case BT_HallDown:
							lightValue = floorState.Down
						case BT_Cab:
							lightValue = floorState.Cab &&
								elevID == Global.ID
						}
						elevio.SetButtonLamp(button, floor, lightValue)
					}
				}
			}
		}
	}
}
