package fsm

import (
	"fmt"
	"time"

	"../common/config"
	. "../common/types"

	/*"../cost_estimator"
	"../monitor"*/

	"../../pkg/elevio"
)

type StateMachineChannels struct {
	ButtonPress       chan ButtonEvent
	FloorSensor       chan int
	ObstructionSwitch chan bool
	NewOrder          chan bool
}

var Node = NodeInfo{
	ID:    "hello",
	State: ES_Init,
	Dir:   MD_Stop,
	Floor: 2,
	Queue: make([]bool, config.MFloors),
}

var Global = GlobalInfo{
	LocalID: "hello",
	Nodes:   make(map[string]NodeInfo),
	Orders:  make([][]FloorState, config.NElevs),
}

func dummyQueue(ch chan<- bool) {
	time.Sleep(5000 * time.Millisecond)
	Node.Queue[2] = true
	ch <- true
}

func ordersInFront() bool {
	if Node.Dir == MD_Stop {
		return false
	}

	for floor, order := range Node.Queue {
		if order {
			diff := Node.Floor - floor
			switch Node.Dir {
			case MD_Up:
				ret := diff <= 0
			case MD_Down:
				ret := diff <= 0
			}
		}
	}
	return false
}

func calculateDirection() MotorDirection {
	var Dir MotorDirection
	for floor, order := range Node.Queue {
		if order {
			switch Node.Dir {
			case MD_Up:
				switch diff := Node.Floor - floor; {
				case diff > 0:

				case diff < 0:
				case diff == 0:
					Dir = MD_Stop
					break Loop
				}
			case MD_Down:
				switch diff := Node.Floor - floor; {
				case diff > 0:
				case diff < 0:
				case diff == 0:
					Dir = MD_Stop
					break Loop
				}
			case MD_Stop:
				switch diff := Node.Floor - floor; {
				case diff > 0:
				case diff < 0:
				case diff == 0:
					Dir = MD_Stop
					break Loop
				}
			}
		}
	}
	return Dir
}

func Run(ch StateMachineChannels) {
	elevio.Init("localhost:15657", config.MFloors)

	//var state = ES_Init

	for i := range Global.Orders {
		Global.Orders[i] = make([]FloorState, config.MFloors)
	}
	go dummyQueue(ch.NewOrder)

	elevio.SetMotorDirection(MD_Down)
	<-ch.FloorSensor
	elevio.SetMotorDirection(MD_Stop)

	for {
		select {
		case <-ch.NewOrder:
			//TODO: Add order to OrderMatrix
			fmt.Println("Hello")

		/* OLD State machine */
		default:
			/*switch state {
			case eS_Init:
				if f.floorIndicator {
					elevio.SetMotorDirection(elevio.MD_Stop)
					state = eS_Idle
				}
			case eS_Idle:
				for _, element := range cost_estimator.OrderQueue {
					if element {
						state = eS_Run
						elevio.SetMotorDirection(elevio.MD_Up)
						//TODO: Implement algorithm to calculate next motor direction
					}
				}
			case eS_Run:
				if f.floorIndicator {
					if cost_estimator.OrderQueue[monitor.ElevStates[monitor.Elev_id].Floor] {
						state = eS_Stop
						elevio.SetMotorDirection(elevio.MD_Stop)
					}
				}
				// TODO: Handle obstruction between floors
			case eS_Stop:
				// TODO: Open door (timer) and monitor obstruction
				go monitor.RemoveGlobalOrder()
				// TODO: Clear order in local orderQueue and global OrderMatrix
				// TODO: Switch to eS_Idle
			}*/
		}
	}
}
