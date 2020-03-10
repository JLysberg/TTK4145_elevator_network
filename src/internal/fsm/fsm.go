package fsm

import (
	"fmt"
	"time"

	// "../common/config"
	// . "../common/types"
	// /*"../cost_estimator"
	// "../monitor"*/
	//"../../pkg/elevio"
	"internal/common/config"
	. "internal/common/types"

	"pkg/elevio"
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
	time.Sleep(4000 * time.Millisecond)
	Node.Queue[2] = true
	ch <- true
}

func orderInFront() MotorDirection {
	fmt.Println("1")
	for floor, order := range Node.Queue {
		fmt.Println("2")
		if order {
			fmt.Println("3")
			diff := Node.Floor - floor
			switch Node.Dir {
			case MD_Up:
				if diff < 0 {
					return MD_Up
				}
			case MD_Down:
				if diff > 0 {
					return MD_Down
				}
			case MD_Stop:
				fmt.Println("5, n.floor:", Node.Floor, ", floor:", floor, "diff:", diff)
				switch {
				case diff < 0:
					fmt.Println("6")
					return MD_Up
				case diff > 0:
					fmt.Println("7")
					return MD_Down
				}
			}
		}
	}
	fmt.Println("4")
	return MD_Stop
}

func orderAvailable() bool {
	for _, order := range Node.Queue {
		if order {
			return true
		}
	}
	return false
}

func calculateDirection() MotorDirection {
	if orderAvailable() {
		return orderInFront()
	}
	return MD_Stop
}

func Run(ch StateMachineChannels) {
	//var state = ES_Init

	for i := range Global.Orders {
		Global.Orders[i] = make([]FloorState, config.MFloors)
	}
	go dummyQueue(ch.NewOrder)

	elevio.SetMotorDirection(MD_Down)
	Node.Floor = <-ch.FloorSensor
	elevio.SetMotorDirection(MD_Stop)

	for {
		select {
		case <-ch.NewOrder:
			//TODO: Add order to OrderMatrix
			dir := calculateDirection()
			elevio.SetMotorDirection(dir)
			Node.Dir = dir
			fmt.Println("Hello, dir: ", dir)
		case floor := <-ch.FloorSensor:
			fmt.Println("Arrived at floor:", floor)
			Node.Floor = floor
			if Node.Queue[floor] {
				Node.Queue[floor] = false
				dir := calculateDirection()
				elevio.SetMotorDirection(dir)
				Node.Dir = dir
			}

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
