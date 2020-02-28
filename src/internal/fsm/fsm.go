package fsm

import (
	"internal/cost_estimator"
	"internal/monitor"

	"pkg/elevio"
)

type _elevatorState int

const (
	eS_Idle _elevatorState = 0
	eS_Run                 = 1
	eS_Stop                = 2
)

func Run() {
	elevio.Init("localhost:15657", monitor.MFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	//drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	//go elevio.PollStopButton(drv_stop)

	//var flag_obstruction = false
	var state = eS_Idle
	for {
		/* Generic event handler for events applicable to every state */
		select {
		case buttonEvent := <-drv_buttons:
			go monitor.AddLocalOrder(buttonEvent)
		/*case f := <-drv_obstr:
			flag_obstruction = f*/
		}

		/* State machine */
		switch state {
		case eS_Idle:
			for _, element := range cost_estimator.OrderQueue {
				if element {
					state = eS_Run
					//TODO: Implement algorithm to calculate next motor direction
				}
			}
		case eS_Run:
			select {
			case floor := <-drv_floors:
				elevio.SetFloorIndicator(floor)
				if cost_estimator.OrderQueue[floor-1] {
					state = eS_Stop
				}
			}
			// TODO: Switch to eS_Stop on obstruction event
		case eS_Stop:
			// TODO: Open door (timer) and monitor obstruction
			// TODO: Clear order in local orderQueue and global OrderMatrix
			// TODO: Switch to eS_Idle
		}
	}
}
