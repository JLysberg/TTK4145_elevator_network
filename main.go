package main

import (
	"flag"
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
	"github.com/JLysberg/TTK4145_elevator_network/internal/fsm"
	"github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
	*/
	//"./internal/network_handler"

	"./internal/common/config"
	. "./internal/common/types"
	"./internal/fsm"

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
	// flag.IntVar(&ID, "id", 0, "id of this elevator")
	flag.StringVar(&port, "port", "15657", "init port")
	flag.Parse()
	ID, _ := strconv.Atoi(id)
	monitor.Global.ID = ID

	elevio.Init("localhost:"+port, config.MFloors)

	// 10.100.23.149
	// 10.100.23.203

	//go run main.go -id=1 -port=15658

	ch := fsm.StateMachineChannels{
		ButtonPress:          make(chan ButtonEvent),
		NewOrder:             make(chan bool),
		FloorSensor:          make(chan int),
		ObstructionSwitch:    make(chan bool),
		//PacketReceiver:   	  make(chan GlobalInfo),
		//PacketSender:     	  make(chan GlobalInfo),
		ButtonLights_Refresh: make(chan int),
		ClearOrder:           make(chan int),
	}

	var(
		GlobalInfoTx = make(chan GlobalInfo)
		GlobalInfoRx = make(chan GlobalInfo)
	)
	//go fsm.Printer()

	go monitor.CostEstimator(ch.NewOrder)
	go monitor.KingOfOrders(ch.ButtonPress, GlobalInfoRx,
		ch.ButtonLights_Refresh, ch.ClearOrder)
	go monitor.LightSetter(ch.ButtonLights_Refresh)

	go elevio.PollButtons(ch.ButtonPress)
	go elevio.PollFloorSensor(ch.FloorSensor)
	go elevio.PollObstructionSwitch(ch.ObstructionSwitch)

	go fsm.Run(ch)
   
	// We define some custom struct to send over the network.
	// Note that all members we want to transmit must be public. Any private members
	//  will be received as zero-values.
	/*
	type HelloMsg struct {
		Message string
		Iter    int
	}
	*/

	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	/*
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	*/

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(30125, id, peerTxEnable)
	go peers.Receiver(30125, peerUpdateCh)

	//15647
	// We make channels for sending and receiving our custom data types
	//helloTx := make(chan HelloMsg)
	//helloRx := make(chan HelloMsg)


	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	//go bcast.Transmitter(16569, helloTx)
	//go bcast.Receiver(16569, helloRx)

	go bcast.Transmitter(30025, GlobalInfoTx)
	go bcast.Receiver(30025, GlobalInfoRx)
	//16569
	// The example message. We just send one of these every second.
	
	//monitor.Global.TestString = "Hello from "

	//sendGlobal := GlobalInfo{"Hello from ", monitor.Global.ID, monitor.Global.Nodes, monitor.Global.Orders}
		//sendGlobal := Global
		

			//helloMsg.Iter++
			//helloTx <- helloMsg
		
	go func() {
		for {
			//fmt.Printf(monitor.Global.TestString, monitor.Global.ID)
			
			GlobalInfoTx <- monitor.Global			
			time.Sleep(500 * time.Millisecond)	
		}
	}()


	fmt.Println("Started")
	go func() {
		for {
			select {
			case p := <-peerUpdateCh:
				fmt.Printf("Peer update:\n")
				fmt.Printf("  Peers:    %q\n", p.Peers)
				fmt.Printf("  New:      %q\n", p.New)
				fmt.Printf("  Lost:     %q\n", p.Lost)

			//case a := <-helloRx:
				//fmt.Printf("Received: %#v\n", a)
			
			case m := <- GlobalInfoRx:
				
				fmt.Printf("Received GlobalInfo: %#v\n", m.ID)
			}
		}
	}()

	select {}
}


/*
KNOWN BUGS:
Jostein:
	- fsm: Elevator does not stop and handle new order if order is at
		elevator's current floor. Caused by PollFloorSensor() goroutine
		which only sends floor on channel on change.

TODO:
Jostein:
	- monitor: Split cost estimator into sereral threads to improve performance.
		Current run time with one elevator and all orders present is about ~2s,
		which is unacceptable and will introduce problems later.
	- fsm: Implement obstruction timer
	- fsm/monitor: Semaphore integration between order clearance in monitor and
		setDirection in fsm
	- monitor: watchdog lookup table in cost estimator

	- network
	- watchdog: lookup table integration with network
*/



/*

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	globalInfoTx := make(chan GlobalInfo)
	globalInfoRx := make(chan GlobalInfo)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, globalInfoTx)
	go bcast.Receiver(16569, globalInfoRx)

	// The example message. We just send one of these every second.
	go func() {
		for {
			globalInfoTx <- GlobalInfo
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-globalInfoRx:
			ch.PacketReceiver <- a
			fmt.Printf("Received: %#v\n", a)
		}
	}

*/



/*

    // We make a channel for receiving updates on the id's of the peers that are
    //  alive on the network
    peerUpdateCh := make(chan peers.PeerUpdate)
    
    // We can disable/enable the transmitter after it has been started.
    // This could be used to signal that we are somehow "unavailable".
    peerTxEnable := make(chan bool)
    go peers.Transmitter(15648, id, peerTxEnable)
    go peers.Receiver(15648, peerUpdateCh)
    go network_handler.SendMsg(ch.PacketSender)
	go network_handler.ReceiveMsg(ch.PacketReceiver)
	


	
*/