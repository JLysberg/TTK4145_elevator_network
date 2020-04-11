package node

import (
	"time"

	/* Setup desc. in main*/
	"github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
)

// func Printer() {
// 	for {
// 		time.Sleep(1 * time.Millisecond)

// 		fmt.Println("State:", monitor.Local.State)
// 	}
// }

func Initialize(floorSensor <-chan int, lightRefresh chan<- int) {
	/*	Enter defined state */
	elevio.SetMotorDirection(MD_Down)
	floor := <-floorSensor
	elevio.SetMotorDirection(MD_Stop)
	/*	Initialize local memory */
	monitor.Local.Dir = MD_Stop
	monitor.Local.LastDir = MD_Down
	monitor.Local.State = ES_Idle
	monitor.Local.Floor = floor
	/*	Refresh all button lights */
	lightRefresh <- -1

	elevio.SetFloorIndicator(floor)
}

var doorTimeout = time.NewTimer(1 * time.Hour)

/*	ElevatorServer handles all elevator logic and communications between local
	routines in current node. */
func ElevatorServer(ch NodeChannels) {
	for {
		select {
		case orderFloor := <-ch.UpdateQueue:
			switch monitor.Local.State {
			case ES_Stop, ES_Idle:
				if orderFloor == monitor.Local.Floor {
					floorStop(orderFloor, ch.ClearOrder)
				} else {
					go setDirection(ch.DoorTimeout)
				}
			case ES_Run:
				go setDirection(ch.DoorTimeout)
			}

		case arrivedFloor := <-ch.FloorSensor:
			go elevio.SetFloorIndicator(arrivedFloor)
			monitor.Local.Floor = arrivedFloor
			if stopCriteria(arrivedFloor) {
				floorStop(arrivedFloor, ch.ClearOrder)
				go setDirection(ch.DoorTimeout)
			}

		case <-doorTimeout.C:
			elevio.SetDoorOpenLamp(false)
			ch.DoorTimeout <- true
			if !orderAvailable() {
				monitor.Local.State = ES_Idle
			} else {
				monitor.Local.State = ES_Run
			}

		case switchEnabled := <-ch.ObstructionSwitch:
			if monitor.Local.State == ES_Stop {
				if switchEnabled {
					doorTimeout.Reset(1 * time.Hour)
				} else {
					doorTimeout.Reset(config.DoorTimeout)
				}
			}
		}
	}
}
