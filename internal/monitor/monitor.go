package monitor

import (
	"fmt"
	"math"
	"time"
	
	"../common/config"
	. "../common/types"
	"../../pkg/elevio"
)

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
func CostEstimator(updateQueue chan<- []FloorState, clearQueue <-chan int, onlineElevators <-chan []bool) {
	var (
		queue = make([]FloorState, config.MFloors)
		onlineList []bool
	)
	for {
		estBegin := time.Now()
		/*	Create copy of queue and pass on if there is a receiver available */
		queueCopy := createQueueCopy(queue)
		select {

		case copyOnlineList := <-onlineElevators:
			onlineList = copyOnlineList
		case getQueueCopy <- queueCopy:
		case clearQueueFloor := <-clearQueue:
			queue[clearQueueFloor].Up = false
			queue[clearQueueFloor].Down = false
			queue[clearQueueFloor].Cab = false
			updateQueue <- createQueueCopy(queue)
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
						(queue[floor].Up || queue[floor].Down || queue[floor].Cab) {
						queue[floor].Up = false
						queue[floor].Down = false
						queue[floor].Cab = false
						updateQueue <- createQueueCopy(queue)
					}
				} else if floorState.Up || floorState.Down {
					bestCost := 100
					bestID := 0
					for nodeID, node := range globalCopy.Nodes {
						cost := 0
						/*	Ignore all offline and malfunctioned nodes */
						if !onlineList[nodeID] || node.State == ES_Error {
						 	continue
						}

						/*	Calculate distance cost */
						floorDiff := int(math.Abs(float64(node.Floor - floor)))
						if floorDiff != 0 {
							cost += floorDiff + 1
						}

						/*	Calculate state cost */
						switch node.State {
						case ES_Run, ES_Stop:
							switch node.LastDir {
								case MD_Down:
									if node.Floor >= floor {
										break
									} else {
										cost += 5
									}
								case MD_Up:
									if node.Floor <= floor {
										break
									} else {
										cost += 5
									}
								default:
									fmt.Println("ERROR: unhandled node.LastDir case")
								}
						case ES_Idle:
							cost++
						default:
							fmt.Println("ERROR: unhandled node.State case")
						}

						if cost < bestCost {
							bestCost = cost
							bestID = nodeID
						}
					}
					/*	Assign order to local node if optimal */
					if bestCost == 100 || (bestID == globalCopy.ID && queue[floor] != floorState) {
						
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
	to guarantee that global is always up to date with the rest of the network */
func OrderServer(id int, buttonPress <-chan ButtonEvent, orderUpdates <-chan GlobalInfo,
	lightRefresh chan<- GlobalInfo, setClearBit <-chan int, clearQueue chan<- int,
	updateLocal <-chan LocalInfo) {
	global := GlobalInfo{
		ID:     id,
		Nodes:  make([]LocalInfo, config.NElevs),
		Orders: make([][]FloorState, config.MFloors),
	}
	for i := range global.Orders {
		global.Orders[i] = make([]FloorState, config.NElevs)
	}
	/*	Declare a ticker which governs the interval in which global is checked
		for handled clear bits */
	remClearTicker := time.NewTicker(1000 * time.Millisecond)

	for {
		select {
		/*	Request copy of local from ElevatorServer and update entry in global */
		case localCopy := <-updateLocal:
			global.Nodes[global.ID] = localCopy
		/*	Create copy of global and pass on if there is a receiver available */
		case getGlobalCopy <- createGlobalCopy(global):
		case pressedButton := <-buttonPress:
			switch pressedButton.Button {
			case BT_HallUp:
				global.Orders[pressedButton.Floor][global.ID].Up = true
			case BT_HallDown:
				global.Orders[pressedButton.Floor][global.ID].Down = true
			case BT_Cab:
				global.Orders[pressedButton.Floor][global.ID].Cab = true
			}
			lightRefresh <- createGlobalCopy(global)

		case msg := <-orderUpdates:
			/*	Only update local global.Nodes if it differs from msg.Orders */
			if msg.Nodes[msg.ID] !=  global.Nodes[msg.ID] && msg.ID != global.ID {
				global.Nodes[msg.ID] = msg.Nodes[msg.ID]
			}
			/*	Only update local global.Orders if it differs from msg.Orders */
			if !equalOrderMatrix(msg.Orders, global.Orders) && msg.ID != global.ID {
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
							global.Orders[msgFloor][msgElevID].Clear =
								global.Orders[msgFloor][msgElevID].Clear || msgFloorState.Clear
							/*	Remove orders on msgFloor */
							params := remOrdersParams {
								Floor: msgFloor,
								ID:    msgElevID,
							}
							go func() {
								remOrders <- params
							}()
						}
					}
				}
				lightRefresh <- createGlobalCopy(global)
			}

		case clearBitFloor := <-setClearBit:
			/*	Set clear bit in global */
			global.Orders[clearBitFloor][global.ID].Clear = true
			/*	Remove orders on clearBitFloor */
			params := remOrdersParams {
				Floor: clearBitFloor,
				ID:    global.ID,
			}
			go func() {
				remOrders <- params
				clearQueue <- clearBitFloor
			}()
			
		case params := <-remOrders:
			/*	Remove all up/down orders on specified floor */
			for elevID := 0; elevID < config.NElevs; elevID++ {
				global.Orders[params.Floor][elevID].Up = false
				global.Orders[params.Floor][elevID].Down = false
				
			}
			/*	Also remove cab order on specified floor, only for specified ID */
			global.Orders[params.Floor][params.ID].Cab = false

			lightRefresh <- createGlobalCopy(global)
			
		case <-remClearTicker.C:
			/*	Check global.Orders for clear bits. Remove clear bit if it has
				been handled by all elevators on the network */
			for floor, floorStates := range global.Orders {
				for elevID1, floorState1 := range floorStates {
					if floorState1.Clear {
						clearHandled := true
						for elevID2, floorState2 := range floorStates {
							if floorState2.Up || floorState2.Down ||
								(floorState2.Cab && elevID1 == elevID2) {
								clearHandled = false
							}
						}
						if clearHandled {
							global.Orders[floor][elevID1].Clear = false
						}
					}
				}
			}
		}
	}
}

/*	LightServer updates every button light in accordance with the global order
	matrix on refresh call. */
func LightServer(lightRefresh <-chan GlobalInfo) {
	for {
		select {
		/*	Request copy of global from monitor */
		case globalCopy := <-lightRefresh:
			for floor, floorStates := range globalCopy.Orders {
				for button := BT_HallUp; button <= BT_HallDown; button++ {
					lightValue := false
					for elevID := 0; elevID < config.NElevs; elevID++ {
						switch button {
						case BT_HallUp:
							lightValue = lightValue || floorStates[elevID].Up
						case BT_HallDown:
							lightValue = lightValue || floorStates[elevID].Down
						}
					}
					elevio.SetButtonLamp(button, floor, lightValue)
				}
				elevio.SetButtonLamp(BT_Cab, floor, globalCopy.Orders[floor][globalCopy.ID].Cab)
			}
		}
	}
}
