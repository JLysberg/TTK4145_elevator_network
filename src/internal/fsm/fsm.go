package fsm

import (
	//"fmt"
	"time"
	"math/rand"

	/* LAB setup */
	. "../common/types"
	"../monitor"
	"../../pkg/elevio"
	"../common/config"

	/* GOPATH setup */
	// . "internal/common/types"
	// "internal/monitor"
	// "pkg/elevio"
)

type StateMachineChannels struct {
	ButtonPress       chan ButtonEvent
	FloorSensor       chan int
	ObstructionSwitch chan bool
	NewOrder          chan bool
	PacketReceiver    chan []byte
}

func dummyQueue(ch chan<- bool) {
	var r int
	for {
		r = rand.Intn(config.MFloors)
		time.Sleep(5000 * time.Millisecond)
		monitor.Node.Queue[r] = true
		elevio.SetButtonLamp(BT_Cab, r, true)
		ch <- true
	}
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
	elevio.SetMotorDirection(dir)
	monitor.Node.Dir = dir
	if dir != MD_Stop{
		monitor.Node.LastDir = monitor.Node.Dir
	}
}

func Run(ch StateMachineChannels) {
	//var state = ES_Init

	//go dummyQueue(ch.NewOrder)

	setNodeDirection(MD_Down)
	monitor.Node.Floor = <-ch.FloorSensor
	setNodeDirection(MD_Stop)

	for {
		select {
		case <-ch.NewOrder:
			switch monitor.Node.State {
			case ES_Stop:
				fallthrough
			case ES_Idle:
				//TODO: jdfksjldfk
			case ES_Run:
				//TODO
			}
			setNodeDirection(calculateDirection())
		case floor := <-ch.FloorSensor:
			elevio.SetFloorIndicator(floor)
			monitor.Node.Floor = floor
			if monitor.Node.Queue[floor] {
				monitor.Node.Queue[floor] = false
				elevio.SetButtonLamp(BT_Cab, floor, false)
				setNodeDirection(calculateDirection())
			}
		}
	}
}
