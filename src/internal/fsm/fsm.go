package fsm

import (
	//"fmt"
	"time"

	/* LAB setup */
	// "../common/config"
	// . "../common/types"
	// /*"../cost_estimator"
	// "../monitor"*/
	//"../../pkg/elevio"

	/* GOPATH setup */
	. "internal/common/types"
	"internal/monitor"
	"pkg/elevio"
)

type StateMachineChannels struct {
	ButtonPress       chan ButtonEvent
	FloorSensor       chan int
	ObstructionSwitch chan bool
	NewOrder          chan bool
	PacketReceiver    chan GlobalInfo
}

func dummyQueue(ch chan<- bool) {
	time.Sleep(4000 * time.Millisecond)
	monitor.Node.Queue[2] = true
	ch <- true
}

func orderInFront() MotorDirection {
	for floor, order := range monitor.Node.Queue {
		if order {
			diff := monitor.Node.Floor - floor
			switch monitor.Node.Dir {
			case MD_Up:
				if diff < 0 {
					return MD_Up
				}
			case MD_Down:
				if diff > 0 {
					return MD_Down
				}
			case MD_Stop:
				switch {
				case diff < 0:
					return MD_Up
				case diff > 0:
					return MD_Down
				}
			}
		}
	}
	return MD_Stop
}

func orderAvailable() bool {
	for _, order := range monitor.Node.Queue {
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

func setNodeDirection(dir MotorDirection) {
	monitor.Node.LastDir = monitor.Node.Dir
	monitor.Node.Dir = dir
	elevio.SetMotorDirection(dir)
}

func Run(ch StateMachineChannels) {
	//var state = ES_Init

	go dummyQueue(ch.NewOrder)

	setNodeDirection(MD_Down)
	monitor.Node.Floor = <-ch.FloorSensor
	setNodeDirection(MD_Stop)

	for {
		select {
		case <-ch.NewOrder:
			setNodeDirection(calculateDirection())
		case floor := <-ch.FloorSensor:
			monitor.Node.Floor = floor
			if monitor.Node.Queue[floor] {
				monitor.Node.Queue[floor] = false
				setNodeDirection(calculateDirection())
			}
		}
	}
}
