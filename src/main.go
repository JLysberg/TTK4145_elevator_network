package main

import (
	"flag"
	"fmt"
	"strconv"

	/* LAB setup */
	// . "./internal/common/types"
	// "./internal/common/config"
	// "./internal/fsm"
	// "./internal/monitor"
	// "./pkg/elevio"

	/* GOPATH setup */
	"internal/common/config"
	. "internal/common/types"
	"internal/fsm"
	"internal/monitor"
	"pkg/elevio"
)

func main() {
	elevio.Init("localhost:15657", config.MFloors)
	var (
		id string
	)

	flag.StringVar(&id, "id", "0", "id of this elevator")
	//flag.IntVar(&ID, "id", 0, "id of this elevator")
	flag.Parse()
	ID, _ := strconv.Atoi(id)
	fmt.Println("ID is", ID)

	ch := fsm.StateMachineChannels{
		ButtonPress:          make(chan ButtonEvent),
		NewOrder:             make(chan bool),
		FloorSensor:          make(chan int),
		ObstructionSwitch:    make(chan bool),
		PacketReceiver:       make(chan []byte),
		ButtonLights_Refresh: make(chan int),
		ClearOrder:           make(chan int),
	}

	//go cost_estimator.UpdateQueue()
	go fsm.Printer()

	go monitor.CostEstimator(ch.NewOrder)
	go monitor.KingOfOrders(ch.ButtonPress, ch.PacketReceiver,
		ch.ButtonLights_Refresh, ch.ClearOrder)
	go monitor.LightSetter(ch.ButtonLights_Refresh)

	go elevio.PollButtons(ch.ButtonPress)
	go elevio.PollFloorSensor(ch.FloorSensor)
	go elevio.PollObstructionSwitch(ch.ObstructionSwitch)

	fsm.Run(ch)
}

/*
KNOWN BUGS:
	- fsm: Elevator does not care whether order at current floor is up/down
		   and stops regardsless. Most likely caused by orderInFront().
	- fsm: Elevator does not stop and handle new order if order is at
		   elevators current floor. Caused by PollFloorSensor() goroutine
		   which only sends floor on channel on change.
*/