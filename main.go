package main

import (
	"flag"
	"strconv"
	
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

	syncCh := NetworkChannels{
		MsgTransmitter: make(chan GlobalInfo),
		MsgReceiver:    make(chan GlobalInfo),
		PeerUpdate:     make(chan peers.PeerUpdate),
		PeerTxEnable:   make(chan bool),
		UpdateOrders:   make(chan GlobalInfo),
		OnlineElevators: make(chan []bool),
	}

	ch := NodeChannels{
		ButtonPress:       make(chan ButtonEvent),
		UpdateQueue:       make(chan []FloorState),
		FloorSensor:       make(chan int),
		ObstructionSwitch: make(chan bool),
		LightRefresh:      make(chan GlobalInfo),
		SetClearBit:       make(chan int),
		ClearQueue:        make(chan int),
		DoorClose:          make(chan bool),
		UpdateLocal:       make(chan LocalInfo),
	}

	go monitor.CostEstimator(ch.UpdateQueue, ch.ClearQueue, syncCh.OnlineElevators)
	go monitor.OrderServer(ID, syncCh.UpdateOrders, ch)
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
