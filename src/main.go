package main

import (
	"internal/fsm"
	. "internal/common/types"
)

func main() {
	ch := fsm.StateMachineChannels{
		buttonPress: make(chan ButtonEvent)
		newOrder: make(chan bool)
		floorSensor: make(chan int)
		obstructionSwitch: make(chan bool)
		
	}

	go cost_estimator.UpdateQueue()

	go monitor.PollOrders(ch.buttonPress)

	go elevio.PollButtons(ch.buttonPress)
	go elevio.PollFloorSensor(ch.floorSensor)
	go elevio.PollObstructionSwitch(ch.obstructionSwitch)

	fsm.Run(ch)
}
