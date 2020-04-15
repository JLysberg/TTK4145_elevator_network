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
var getGlobalCopy = make(chan GlobalInfo)

func createQueueCopy(queue []FloorState) []FloorState {
	copy := make([]FloorState, len(queue))
	for i, k := range queue {
		copy[i] = k
	}
	return copy
}

func createGlobalCopy(global GlobalInfo) GlobalInfo {
	copy := global
	copy.Orders = make([][]FloorState, len(global.Orders))
	copy.Nodes = make([]LocalInfo, len(global.Nodes))
	for i, v := range global.Orders {
		copy.Orders[i] = make([]FloorState, len(v))
		for j, k := range v {
			copy.Orders[i][j] = k
		}
	}
	for i, v := range global.Nodes {
		copy.Nodes[i] = v
	}
	return copy
}

func equalOrderMatrix(m1 [][]FloorState, m2 [][]FloorState) bool {
	if len(m1) == len(m2) && len(m1[0]) == len(m2[0]) {
		for i, v := range m1 {
			for j, k := range v {
				if m2[i][j] != k {
					return false
				}
			}
		}
	} else {
		return false
	}
	return true
}

/*	Queue gives a call to CostEstimator to return a copy of queue */
func Queue() []FloorState {
	return <-getQueueCopy
}

/*	Global gives a call to CostEstimator to return a copy of global */
func Global() GlobalInfo {
	return <-getGlobalCopy
}

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
		/*	Create copy of queue and pass on if there is a receiver available */
		queueCopy := createQueueCopy(queue)
		select {
		case getQueueCopy <- queueCopy:
		default:
		}
		/*	Request a copy of Global from OrderServer */
		globalCopy := Global()
		/*	Always assign cab orders to local node */
		for floor, floorStates := range globalCopy.Orders {
			if floorStates[globalCopy.ID].Cab && !floorStates[globalCopy.ID].Clear &&
				!queue[floor].Cab {
				queue[floor].Cab = true
				updateQueue <- createQueueCopy(queue)
			}
		}
		/*	Cost calculation for non-cab orders */
		for floor, floorStates := range globalCopy.Orders {
			for elevID, floorState := range floorStates {
				if floorState.Clear {
					if elevID == globalCopy.ID &&
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
					for nodeID, node := range globalCopy.Nodes {
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
					if bestID == globalCopy.ID && queue[floor] != floorState {
						queue[floor] = floorState
						queueCopy := createQueueCopy(queue)
						updateQueue <- queueCopy
					}
				}
			}
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
	to guarantee that global.Orders is always up to date with the rest of the network */
func OrderServer(id int, buttonPress <-chan ButtonEvent, newPackets <-chan []byte,
				  lightRefresh chan<- int, clearOrder <-chan int) {
	global := GlobalInfo {
		ID: 	id,
		Nodes: 	make([]LocalInfo, config.NElevs),
		Orders: make([][]FloorState, config.MFloors),
	}
	for i := range global.Orders {
		global.Orders[i] = make([]FloorState, config.NElevs)
	}

	for {
		/*	Create copy of global and pass on if there is a receiver available */
		globalCopy := createGlobalCopy(global)
		select {
		case getGlobalCopy <- globalCopy:
		case pressedButton := <-buttonPress:
			switch pressedButton.Button {
			case BT_HallUp:
				global.Orders[pressedButton.Floor][global.ID].Up = true
			case BT_HallDown:
				global.Orders[pressedButton.Floor][global.ID].Down = true
			case BT_Cab:
				global.Orders[pressedButton.Floor][global.ID].Cab = true
			}
			lightRefresh <- pressedButton.Floor
		case receivedPackage := <-newPackets:
			var msg GlobalInfo
			err := json.Unmarshal(receivedPackage, &msg)
			if err != nil {
				fmt.Println("Error with unmarshaling message in Monitor:", err)
			}
			/*	Only update local global.Orders if it differs from msg.Orders */
			if !equalOrderMatrix(msg.Orders, global.Orders) {
				for msgFloor, msgFloorStates := range msg.Orders {
					for msgElevID, msgFloorState := range msgFloorStates {
						if !msgFloorState.Clear {
							/*	Concatenate orders from msg into local order matrix */
							global.Orders[msgFloor][msgElevID].Up =
								global.Orders[msgFloor][msgElevID].Up || msgFloorState.Up 		
							global.Orders[msgFloor][msgElevID].Down =
								global.Orders[msgFloor][msgElevID].Down || msgFloorState.Down
							global.Orders[msgFloor][msgElevID].Cab =
								global.Orders[msgFloor][msgElevID].Cab || msgFloorState.Cab
						} else {
							/*	Remove all up/down orders if there is a clear present */
							for elevID := 0; elevID < config.NElevs; elevID++ {
								global.Orders[msgFloor][elevID].Up = false
								global.Orders[msgFloor][elevID].Down = false
							}
							/*	Also remove cab order if present */
							global.Orders[msgFloor][msgElevID].Cab = false
						}
					}
				}
				lightRefresh <- -1
			}
		case clearFloor := <-clearOrder:
			/*	Set clear value in global which is removed after 1 second 
				OBS: Techically, this solution may result in read/write conflicts with global*/
			go func() {
				global.Orders[clearFloor][global.ID].Clear = true
				timeout := time.NewTimer(config.ClearTimeout)
				<-timeout.C
				global.Orders[clearFloor][global.ID].Clear = false
			}()
			lightRefresh <- clearFloor
			/*	The following block might be superfluous when networks are introduced*/
			/********************************************/
			/*	Remove all up/down orders if there is a clear present */
			for elevID := 0; elevID < config.NElevs; elevID++ {
				global.Orders[clearFloor][elevID].Up = false
				global.Orders[clearFloor][elevID].Down = false
			}
			/*	Also remove cab order if present */
			global.Orders[clearFloor][global.ID].Cab = false
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
			/*	Request a copy of global from OrderServer */
			globalCopy := Global()
			for floor, floorStates := range globalCopy.Orders {
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
								elevID == globalCopy.ID
						}
						elevio.SetButtonLamp(button, floor, lightValue)
					}
				}
			}
		}
	}
}
