package main

import (
	"flag"
	"strconv"

	/*
		1. Set GOPATH environment variable, e.g. to %USERPROFILE%/go. Can NOT
		   equal to %GOROOT%.
		2. Pull repository with "go get github.com/JLysberg/TTK4145_elevator_network"
		3. The following import paths should be compatible
	*/
	"github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	"github.com/JLysberg/TTK4145_elevator_network/internal/node"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
)

func main() {
	elevio.Init("localhost:15657", config.MFloors)
	var (
		id string
	)

	flag.StringVar(&id, "id", "0", "id of this elevator")
	// flag.IntVar(&ID, "id", 0, "id of this elevator")
	flag.Parse()
	ID, _ := strconv.Atoi(id)
	ID = 0
	monitor.Global.ID = ID

	ch := NodeChannels{
		ButtonPress:       make(chan ButtonEvent),
		UpdateQueue:          make(chan int),
		FloorSensor:       make(chan int),
		ObstructionSwitch: make(chan bool),
		PacketReceiver:    make(chan []byte),
		LightRefresh:      make(chan int),
		ClearOrder:        make(chan int),
		DoorTimeout:       make(chan bool),
	}

	go node.Initialize(ch.FloorSensor, ch.LightRefresh)
	// go node.Printer()

	go monitor.CostEstimator(ch.UpdateQueue)
	go monitor.OrderServer(ch.ButtonPress, ch.PacketReceiver,
		ch.LightRefresh, ch.ClearOrder)
	go monitor.LightServer(ch.LightRefresh)

	go elevio.PollButtons(ch.ButtonPress)
	go elevio.PollFloorSensor(ch.FloorSensor)
	go elevio.PollObstructionSwitch(ch.ObstructionSwitch)

	node.ElevatorServer(ch)
}

/*
TODO:
Jostein:
	- monitor: Split cost estimator into sereral threads to improve performance.
		Current run time with one elevator and all orders present is about ~2s,
		which is unacceptable and will introduce problems later.
	- fsm/monitor: Semaphore integration between order clearance in monitor and
		setDirection in fsm
	- monitor: watchdog lookup table in cost estimator

	- network
	- watchdog: lookup table integration with network
*/
