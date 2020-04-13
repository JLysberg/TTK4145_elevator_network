package main

import (
	"flag"
	"strconv"
	//"time"
	//"fmt"
	//"os"

	/*
		1. Set GOPATH environment variable, e.g. to %USERPROFILE%/go. Can NOT
		   equal to %GOROOT%.
		2. Pull repository with "go get github.com/JLysberg/TTK4145_elevator_network"
		3. The following import paths should be compatible
	

	"github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	"github.com/JLysberg/TTK4145_elevator_network/internal/node"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
	*/
	//"./internal/network_handler"

	"./internal/common/config"
	. "./internal/common/types"
	"./internal/node"
	

	"./internal/monitor"
	"./internal/sync"
	"./pkg/elevio"
	"./pkg/network/peers"
	"./pkg/network/bcast"
	//"./pkg/network/localip"

)

func main() {	
	
	var (
		id string
		port string
	)

	flag.StringVar(&id, "id", "0", "id of this elevator")
	flag.StringVar(&port, "port", "15657", "init port")
	flag.Parse()
	ID, _ := strconv.Atoi(id)
	//ID = 0?
	monitor.Global.ID = ID

	elevio.Init("localhost:"+port, config.MFloors)

	// 10.100.23.149
	// 10.100.23.174
	//test_network_1: .223 and .247

	//go run main.go -id=1 -port=15658

	sch := sync.NetworkChannels{
		MsgTransmitter: 	make(chan GlobalInfo),
		MsgReceiver: 		make(chan GlobalInfo),
		PeerUpdate: 		make(chan peers.PeerUpdate),
		PeerTxEnable:		make(chan bool),
	}
	
	ch := NodeChannels{
		ButtonPress:       make(chan ButtonEvent),
		UpdateQueue:       make(chan int),
		FloorSensor:       make(chan int),
		ObstructionSwitch: make(chan bool),
		LightRefresh:      make(chan int),
		ClearOrder:        make(chan int),
		DoorTimeout:       make(chan bool),
	}

	//go node.Printer()
	go node.Initialize(ch.FloorSensor, ch.LightRefresh)
	go monitor.CostEstimator(ch.UpdateQueue)
	go monitor.OrderServer(ch.ButtonPress, sch.MsgReceiver,
		ch.LightRefresh, ch.ClearOrder)
	//go sync.SyncMessages(sch.MsgTransmitter, sch.MsgReceiver, sch.PeerUpdate, id)
	go sync.SyncMessages(sch, ID)
	go monitor.LightServer(ch.LightRefresh)

	go elevio.PollButtons(ch.ButtonPress)
	go elevio.PollFloorSensor(ch.FloorSensor)
	go elevio.PollObstructionSwitch(ch.ObstructionSwitch)

	go bcast.Transmitter(30025, sch.MsgTransmitter)
	go bcast.Receiver(30025, sch.MsgReceiver)
	go peers.Transmitter(30125, id, sch.PeerTxEnable)
	go peers.Receiver(30125, sch.PeerUpdate)

	go node.ElevatorServer(ch)
	select{}
}


/*
TODO:
Jostein:
	- fsm/monitor: Semaphore integration between order clearance in monitor and
		setDirection in fsm

	- network
	- watchdog: lookup table integration with network
*/

