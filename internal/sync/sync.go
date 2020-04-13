	package sync
	
	import (
		"fmt"
		"time"

		"../common/config"
		. "../common/types"
		"../monitor"
		"../../pkg/elevio"
		"../../pkg/network/peers"
		"../../pkg/network/bcast"
		"../../pkg/network/localip"
	)

	type NetworkChannels struct{
		MsgTransmitter		chan GlobalInfo
		MsgReceiver			chan GlobalInfo
		PeerUpdate			chan peers.PeerUpdare
		PeerTxEnable		chan bool
	}


	func SyncMessages(ch NetworkChannels, id int){
		var(
			sendMsg		GlobalInfo 
			nodes		monitor.GlobalInfo.Nodes  //could be directly inserted into the send case?
			orders		monitor.GlobalInfo.Orders
			//nodes 		[config.NEleevs]
			//orders		[config.MFloors][config.NElevs]
		)

		timeout := make(chan bool)
		go func() { time.Sleep(1 * time.Second); timeout <- true }()

		bcastTicker := time.NewTicker(500 * time.Millisecond)

		select {
		case msg := <- ch.GlobalInfoRx
			//update ElevLastSent?
			//update onlineList?

		case <- bcastTicker.C:
			sendMsg.ID = id
			sendMsg.Nodes = nodes
			sendMsg.Orders = orders
			ch.GlobalInfoTx <- sendMsg

		case p := <- ch.UpdatePeers:
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