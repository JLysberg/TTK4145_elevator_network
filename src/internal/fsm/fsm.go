package fsm

import (
	//"fmt"

	"internal/cost_estimator"
	"internal/monitor"

	"pkg/elevio"
)

type _elevatorState int

const (
	eS_Init _elevatorState = 0
	eS_Idle                = 1
	eS_Run                 = 2
	eS_Stop                = 3
)

type _fsmFlags struct {
	floorIndicator bool
	obstruction    bool
}

var f _fsmFlags

func clearFlags() {
	f.floorIndicator = false
	f.obstruction = false
}

func Run() {
	elevio.Init("localhost:15657", monitor.MFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	go cost_estimator.UpdateQueue()

	go monitor.PollOrders(drv_buttons)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	var state = eS_Init
	elevio.SetMotorDirection(elevio.MD_Down)
	for {
		clearFlags()

		select {
		/* Generic event handler for events applicable to every state */
		case floor := <-drv_floors:
			f.floorIndicator = true
			elevio.SetFloorIndicator(floor)
			monitor.ElevStates[monitor.Elev_id].Floor = floor

		/* State machine */
		default:
			switch state {
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
			}
		}
	}
}
