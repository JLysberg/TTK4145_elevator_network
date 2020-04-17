package main

import (
	"flag"
	"strconv"

	//"time"
	//"fmt"
	//"os"

	// "github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	// . "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	// "github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	// "github.com/JLysberg/TTK4145_elevator_network/internal/node"
	// "github.com/JLysberg/TTK4145_elevator_network/internal/sync"
	// "github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
	// "github.com/JLysberg/TTK4145_elevator_network/pkg/network/peers"
	// "github.com/JLysberg/TTK4145_elevator_network/pkg/network/bcast"
	
	"./internal/common/config"
	. "./internal/common/types"
	"./internal/node"


	"./internal/monitor"
	"./internal/sync"
	"./pkg/elevio"
	"./pkg/network/peers"
	"./pkg/network/bcast")

func main() {

	var (
		id   string
		port string
	)

	flag.StringVar(&id, "id", "0", "id of this elevator")
	flag.StringVar(&port, "port", "15657", "init port")
	flag.Parse()
	ID, _ := strconv.Atoi(id)

	elevio.Init("localhost:"+port, config.MFloors)

	// 10.100.23.149
	// 10.100.23.174
	//test_network_1: .223 and .247

	//go run main.go --port=15658 -id=1
	//Simulator: qwe(UP) - sdf (DOWN) - zxc (CAB)

	syncCh := sync.NetworkChannels{
		MsgTransmitter: make(chan GlobalInfo),
		MsgReceiver:    make(chan GlobalInfo),
		PeerUpdate:     make(chan peers.PeerUpdate),
		PeerTxEnable:   make(chan bool),
		UpdateOrders:   make(chan GlobalInfo),
	}

	ch := NodeChannels{
		ButtonPress:       make(chan ButtonEvent),
		UpdateQueue:       make(chan []FloorState),
		FloorSensor:       make(chan int),
		ObstructionSwitch: make(chan bool),
		LightRefresh:      make(chan GlobalInfo),
		ClearOrder:        make(chan int),
		DoorOpen:          make(chan bool),
	}

	// go node.Printer()

	go monitor.CostEstimator(ch.UpdateQueue)
	go monitor.OrderServer(ID, ch.ButtonPress, syncCh.UpdateOrders,
		ch.LightRefresh, ch.ClearOrder)
	go sync.SyncMessages(syncCh, ID)
	go monitor.LightServer(ch.LightRefresh)

	go elevio.PollButtons(ch.ButtonPress)
	go elevio.PollFloorSensor(ch.FloorSensor)
	go elevio.PollObstructionSwitch(ch.ObstructionSwitch)

	go bcast.Transmitter(30025, syncCh.MsgTransmitter)
	go bcast.Receiver(30025, syncCh.MsgReceiver)
	go peers.Transmitter(30125, id, syncCh.PeerTxEnable)
	go peers.Receiver(30125, syncCh.PeerUpdate)

	go node.ElevatorServer(ch)
	select {}
}
