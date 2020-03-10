package main

import (
	// "./pkg/elevio"

	// "./internal/fsm"
	// . "./internal/common/types"
	// /*"./internal/cost_estimator"
	// "./internal/monitor"*/
	"pkg/elevio"

	"internal/common/config"
	. "internal/common/types"
	"internal/fsm"
)

func main() {
	elevio.Init("localhost:15657", config.MFloors)

	ch := fsm.StateMachineChannels{
		ButtonPress:       make(chan ButtonEvent),
		NewOrder:          make(chan bool),
		FloorSensor:       make(chan int),
		ObstructionSwitch: make(chan bool),
	}

	//go cost_estimator.UpdateQueue()

	//go monitor.PollOrders(ch.ButtonPress)

	go elevio.PollButtons(ch.ButtonPress)
	go elevio.PollFloorSensor(ch.FloorSensor)
	go elevio.PollObstructionSwitch(ch.ObstructionSwitch)

	fsm.Run(ch)
}
