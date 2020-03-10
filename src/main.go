package main

import (
	"flag"
	"fmt"
	"strconv"

	/* LAB setup */
	// "./pkg/elevio"
	// "./internal/fsm"
	// . "./internal/common/types"
	// /*"./internal/cost_estimator"
	// "./internal/monitor"*/

	/* GOPATH setup */
	"internal/common/config"
	. "internal/common/types"
	"internal/fsm"
	"pkg/elevio"
)

func main() {
	elevio.Init("localhost:15657", config.MFloors)
	var (
		id    string
	)

	flag.StringVar(&id, "id", "0", "id of this elevator")
	//flag.IntVar(&ID, "id", 0, "id of this elevator")
	flag.Parse()
	ID, _ := strconv.Atoi(id)
	fmt.Print("ID is", ID)

	ch := fsm.StateMachineChannels{
		ButtonPress:       make(chan ButtonEvent),
		NewOrder:          make(chan bool),
		FloorSensor:       make(chan int),
		ObstructionSwitch: make(chan bool),
		PacketReceiver: make(chan GlobalInfo),
		
	}

	//go cost_estimator.UpdateQueue()

	//go monitor.PollOrders(ch.ButtonPress)

	go elevio.PollButtons(ch.ButtonPress)
	go elevio.PollFloorSensor(ch.FloorSensor)
	go elevio.PollObstructionSwitch(ch.ObstructionSwitch)

	fsm.Run(ch)
}
