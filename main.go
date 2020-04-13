package main

import (
	"flag"
	"sync"
	"strconv"
	"time"
	"fmt"
	"os"

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
	"./pkg/elevio"
	"./pkg/network/peers"
	"./pkg/network/bcast"
	"./pkg/network/localip"

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
	monitor.Global.ID = ID

	elevio.Init("localhost:"+port, config.MFloors)

	var(
		GlobalInfoTx = make(chan GlobalInfo)
		GlobalInfoRx = make(chan GlobalInfo)
	)	
	
	ch := NodeChannels{
		ButtonPress:       make(chan ButtonEvent),
		UpdateQueue:          make(chan int),
		FloorSensor:       make(chan int),
		ObstructionSwitch: make(chan bool),
		LightRefresh:      make(chan int),
		ClearOrder:        make(chan int),
		DoorTimeout:       make(chan bool),
	}

	go node.Printer()
	go node.Initialize(ch.FloorSensor, ch.LightRefresh)
	go monitor.CostEstimator(ch.UpdateQueue)
	go monitor.OrderServer(ch.ButtonPress, GlobalInfoRx,
		ch.LightRefresh, ch.ClearOrder)
	go monitor.LightServer(ch.LightRefresh)

	go elevio.PollButtons(ch.ButtonPress)
	go elevio.PollFloorSensor(ch.FloorSensor)
	go elevio.PollObstructionSwitch(ch.ObstructionSwitch)
	
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(30125, id, peerTxEnable)
	go peers.Receiver(30125, peerUpdateCh)

	go bcast.Transmitter(30025, GlobalInfoTx)
	go bcast.Receiver(30025, GlobalInfoRx)

	var _mtx sync.Mutex
	go func() {
		for {
			_mtx.Lock()
			GlobalInfoTx <- monitor.Global			
			_mtx.Unlock()
			time.Sleep(1 * time.Second)	
		}
	}()
	go node.ElevatorServer(ch)
	select{}
}