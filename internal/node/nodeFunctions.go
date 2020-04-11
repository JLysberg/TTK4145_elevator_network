package node

import (
	"time"

	/* Setup desc. in main*/
	"github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
)

func orderInFront() bool {
	for floor, floorState := range monitor.Local.Queue {
		if floorState.Cab || floorState.Up || floorState.Down {
			diff := monitor.Local.Floor - floor
			switch monitor.Local.LastDir {
			case MD_Up:
				if diff < 0 {
					return true
				}
			case MD_Down:
				if diff > 0 {
					return true
				}
			}
		}
	}
	return false
}

func orderAvailable() bool {
	for _, floorState := range monitor.Local.Queue {
		if floorState.Cab || floorState.Up || floorState.Down {
			return true
		}
	}
	return false
}

func calculateDirection() MotorDirection {
	if orderAvailable() {
		if orderInFront() {
			return monitor.Local.LastDir
		}
		return -1 * monitor.Local.LastDir
	}
	return MD_Stop
}

var setDirectionInstance = false

func setDirection(doorOpen <-chan bool) {
	/* Ensure only one instance of this thread is running at once */
	if setDirectionInstance {
		return
	}
	setDirectionInstance = true
	/*	Minor delay to allow cost estimator to evaluate orders
		CONSIDER USING SEMAPHORES */
	time.Sleep(1 * time.Nanosecond)
	/*	Safety loop to ensure direction is never changed while door is open */
	if monitor.Local.State == ES_Stop {
	safety:
		for {
			select {
			case <-doorOpen:
				break safety
			}
		}
	}
	/*	Calculate, set and save direction to local memory */
	dir := calculateDirection()
	elevio.SetMotorDirection(dir)
	monitor.Local.Dir = dir
	if dir != MD_Stop {
		monitor.Local.LastDir = monitor.Local.Dir
		monitor.Local.State = ES_Run
	}
	setDirectionInstance = false
}

func stopCriteria(floor int) bool {
	floorState := monitor.Local.Queue[floor]
	return floorState.Up && monitor.Local.Dir == MD_Up ||
		floorState.Down && monitor.Local.Dir == MD_Down ||
		(floorState.Down || floorState.Up) && !orderInFront() ||
		floorState.Cab
}

func floorStop(floor int, clearOrder chan<- int) {
	/* Stop */
	clearOrder <- floor
	elevio.SetMotorDirection(MD_Stop)
	monitor.Local.State = ES_Stop
	monitor.Local.Dir = MD_Stop
	/* Open door */
	doorTimeout.Reset(config.DoorTimeout)
	elevio.SetDoorOpenLamp(true)
}