	package sync
	
	import (
		"fmt"
		"time"

		//"../common/config"
		. "../common/types"
		"../monitor"
		//"../../pkg/elevio"
		"../../pkg/network/peers"
		//"../../pkg/network/bcast"
		//"../../pkg/network/localip"
	)

	type NetworkChannels struct{
		MsgTransmitter		chan GlobalInfo
		MsgReceiver			chan GlobalInfo
		PeerUpdate			chan peers.PeerUpdate
		PeerTxEnable		chan bool
	}


	func SyncMessages(ch NetworkChannels, id int){
		var(
			sendMsg		GlobalInfo 
			//nodes		monitor.Global.Nodes  //could be directly inserted into the send case?
			//orders		monitor.Global.Orders
			//nodes 		[config.NEleevs]
			//orders		[config.MFloors][config.NElevs]
		)

		timeout := make(chan bool)
		go func() { time.Sleep(1 * time.Second); timeout <- true }()

		bcastTicker := time.NewTicker(500 * time.Millisecond)

		select {
		//case msg := <- ch.MsgReceiver:
			//update ElevLastSent?
			//update onlineList?

		case <- bcastTicker.C:
			sendMsg.ID = id
			sendMsg.Nodes = monitor.Global.Nodes   //nodes
			sendMsg.Orders = monitor.Global.Orders //orders
			ch.MsgTransmitter <- sendMsg

		case p := <- ch.PeerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			/*
			if id == "" {
				localIP, err := localip.LocalIP()
				if err != nil {
					fmt.Println(err)
					localIP = "DISCONNECTED"
				}
				id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
			}
			*/
		}

	}
