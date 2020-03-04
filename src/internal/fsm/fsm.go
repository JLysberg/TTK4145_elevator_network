package fsm

import (
	//"fmt"

	. "internal/common/types"
	"internal/common/config"
	"internal/cost_estimator"
	"internal/monitor"

	"pkg/elevio"
)

type StateMachineChannels struct {
	buttonPress       chan ButtonEvent
	floorSensor       chan int
	obstructionSwitch chan bool
	newOrder 		  chan bool
}

func Run(ch StateMachineChannels) {
	elevio.Init("localhost:15657", config.MFloors)

	var state = eS_Init
	elevio.SetMotorDirection(MD_Down)
	for {

		select {
		case floor := <-ch.floorSensor:
			//Do something
		case <-ch.newOrder:
			//TODO: Add order to OrderMatrix

		/* State machine */
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
