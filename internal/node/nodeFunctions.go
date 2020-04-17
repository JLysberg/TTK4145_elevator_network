package node

import (
	// "fmt"
	"sync"

	// "github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	// . "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	// "github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	// "github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
	
	"../common/config"
	. "../common/types"
	"../monitor"
	"../../pkg/elevio")

var getLocalCopy = make(chan LocalInfo)
var setLocalDir = make(chan MotorDirection)

func orderInFront(local LocalInfo, queue []FloorState) bool {
	for floor, floorState := range queue {
		if floorState.Cab || floorState.Up || floorState.Down {
			diff := local.Floor - floor
			switch local.LastDir {
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

func orderAvailable(queue []FloorState) bool {
	for _, floorState := range queue {
		if floorState.Cab || floorState.Up || floorState.Down {
			return true
		}
	}
	return false
}

func calculateDirection(local LocalInfo, queue []FloorState) MotorDirection {
	if orderAvailable(queue) {
		if orderInFront(local, queue) {
			return local.LastDir
		}
		return -1 * local.LastDir
	}
	return MD_Stop
}

var (
	setDirectionInstance bool
	setDirectionInstanceMx sync.Mutex
)

func setDirection(doorOpen <-chan bool, queue []FloorState) {
	/* Ensure only one instance of this thread is running at once */
	setDirectionInstanceMx.Lock()
	start := !setDirectionInstance
	setDirectionInstance = true
	setDirectionInstanceMx.Unlock()
	if start {
		go func() {
			/*	Get copy of local from ElevatorServer */
			local := Local()
			/*	Safety loop to ensure direction is never changed while door is open */
			if local.State == ES_Stop {
			safety:
				for {
					select {
					case <-doorOpen:
						/*	Update local and queue in case of change */
						local = Local()
						queue = monitor.Queue()
						break safety
					}
				}
			}
			/*	Calculate, set and save direction to local memory */
			dir := calculateDirection(local, queue)
			elevio.SetMotorDirection(dir)

			setLocalDir <- dir

			setDirectionInstanceMx.Lock()
			setDirectionInstance = false
			setDirectionInstanceMx.Unlock()
		}()
	}
}

func stopCriteria(floor int, local LocalInfo, queue []FloorState) bool {
	floorState := queue[floor]
	return floorState.Up && local.Dir == MD_Up ||
		floorState.Down && local.Dir == MD_Down ||
		(floorState.Down || floorState.Up) &&
			!orderInFront(local, queue) ||
		floorState.Cab
}

func floorStop(floor int, clearOrder chan<- int) {
	/* Stop */
	clearOrder <- floor
	elevio.SetMotorDirection(MD_Stop)
	/* Open door */
	doorTimeout.Reset(config.DoorTimeout)
	elevio.SetDoorOpenLamp(true)
}
