package main

import (
	"flag"
	"strconv"
	"./pkg/elevio"

	"./internal/fsm"
	. "./internal/common/types"
	/*"./internal/cost_estimator"
	"./internal/monitor"*/
)

func main() {
	var (
		id    string
		ID    int
	)

	flag.StringVar(&id, "id", "0", "id of this elevator")
	//flag.IntVar(&ID, "id", 0, "id of this elevator")
	flag.Parse()
	ID, _ := strconv.Atoi(id)

	ch := fsm.StateMachineChannels{
		ButtonPress: make(chan ButtonEvent),
		NewOrder: make(chan bool),
		FloorSensor: make(chan int),
		ObstructionSwitch: make(chan bool),
	}

	//go cost_estimator.UpdateQueue()

	//go monitor.PollOrders(ch.ButtonPress)

	go elevio.PollButtons(ch.ButtonPress)
	go elevio.PollFloorSensor(ch.FloorSensor)
	go elevio.PollObstructionSwitch(ch.ObstructionSwitch)

	fsm.Run(ch)
}
