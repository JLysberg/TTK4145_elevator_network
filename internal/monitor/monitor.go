package monitor

import (
	"fmt"
	"math"
	"time"

	/* Setup desc. in main */
	/*"github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
	*/
		"../common/config"
		. "../common/types"
		"../../pkg/elevio"
	)

var getGlobalCopy = make(chan GlobalInfo)

func Global() GlobalInfo {
	return <-getGlobalCopy
}

var Local LocalInfo

/*	CostEstimator is a goroutine which continuously assigns orders from
	the global order matrix to any node, taking multiple factors into account.
	All active orders are always assigned to the elevator with the least cost.
	The responsibility of CostEstimator is to guarantee that Local.Queue
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
func CostEstimator(updateQueue chan<- int) {
	for {
		estBegin := time.Now()
		/*	Always assign cab orders to local node */
		for floor, floorStates := range Global().Orders {
			if floorStates[Global().ID].Cab && !floorStates[Global().ID].Clear {
				Local.Queue[floor].Cab = true
				updateQueue <- floor
			}
		}
		/*	Cost calculation for non-cab orders */
		for floor, floorStates := range Global().Orders {
			for elevID, floorState := range floorStates {
				if floorState.Clear {
					if elevID == Global().ID {
						Local.Queue[floor].Up = false
						Local.Queue[floor].Down = false
						Local.Queue[floor].Cab = false
					}
				} else if floorState.Up || floorState.Down {
					bestCost := int(math.Inf(1))
					bestID := 0
					cost := 0
					for nodeID, node := range Global().Nodes {
						/*	Ignore all offline nodes */
						if !Local.OnlineList[nodeID] {
							continue
						}

						/*	Calculate distance cost */
						floorDiff := int(math.Abs(float64(node.Floor - floor)))
						if floorDiff != 0 {
							cost += floorDiff + 1
						}

						/*	Calculate state cost */
						switch node.Dir {
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
					if bestID == Global().ID && Local.Queue[floor] != floorState {
						Local.Queue[floor] = floorState
						updateQueue <- floor
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
	to guarantee that Global.Orders is always up to date with the rest of the network */
func OrderServer(id int, buttonPress <-chan ButtonEvent, newPackets <-chan GlobalInfo,
	lightRefresh chan<- int, clearOrder <-chan int) {
	Global := GlobalInfo{}
	Global.ID = id


	for {
		//Make a copy of global
		globalCopy := createGlobalCopy(Global)
		select {
		//send copy when there is a receiver available for this channel
		case getGlobalCopy <- globalCopy:

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

		case msg := <-newPackets:
			/*	Only update local Global.Orders if it differs from msg.Orders */

			//print both registered local and global orders
			/*
				fmt.Println("Got a network order")
				for i, _ := range Global.Orders {
					fmt.Println("Elev 0, Local:", Global.Orders[i][0],
						"Network:", msg.Orders[i][0])
				}
				for i, _ := range Global.Orders {
					fmt.Println("Elev 1, Local:", Global.Orders[i][1],
						"Network:", msg.Orders[i][1])
				}
				fmt.Println()
			*/
			//
			if msg.ID != Global.ID {
				fmt.Println("Got a network order")
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
							//Should probably be done in the next case?
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
			
			Global.Orders[clearFloor][Global.ID].Clear = true
			timeout := time.NewTimer(config.ClearTimeout)
			<-timeout.C
			Global.Orders[clearFloor][Global.ID].Clear = false

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

func createGlobalCopy(global GlobalInfo) GlobalInfo {
	cpy := global
	cpy.Orders = global.Orders
	for i, v := range global.Orders {
		cpy.Orders[i] = global.Orders[i]
		for j, k := range v {
			cpy.Orders[i][j] = k
		}
	}
	return cpy
}

/*	LightServer updates every button light in accordance with the global order
	matrix on refresh call. A refresh call of -1 updates all buttons, and
	any specific floor call restricts the iteration to said floor. */
func LightServer(lightRefresh <-chan int) {
	for {
		select {
		case callingFloor := <-lightRefresh:
			for floor, floorStates := range Global().Orders {
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
								elevID == Global().ID
						}
						//fmt.Println("Set the lights for " , button, " in floor ", floor , " to ", lightValue)

						elevio.SetButtonLamp(button, floor, lightValue)
					}
				}
			}
		}
	}
}
