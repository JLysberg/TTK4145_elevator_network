package sync

import (
	"fmt"
	"os"
	"strconv"
	"time"

	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/network/localip"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/network/peers"
	/*
		"../../pkg/network/localip"
		"../../pkg/network/peers"
		. "../common/types"
		"../monitor"
	*/)

type NetworkChannels struct {
	MsgTransmitter chan GlobalInfo
	MsgReceiver    chan GlobalInfo
	PeerUpdate     chan peers.PeerUpdate
	PeerTxEnable   chan bool
	//
	//UpdateClear  chan int
	UpdateOrders chan GlobalInfo
}

func SyncMessages(ch NetworkChannels, id int) {
	var (
	//sendMsg GlobalInfo
	//nodes 		[config.NElevs]
	//orders		[config.MFloors][config.NElevs]
	)

	bcastTicker := time.NewTicker(500 * time.Millisecond)
	//bcastTicker := time.NewTicker(2 * time.Second)

	for {
		select {

		case getMsg := <-ch.MsgReceiver:
			ch.UpdateOrders <- getMsg

		//update ElevLastSent?
		//update onlineList?

		case <-bcastTicker.C:
			fmt.Println("Broadcasting message")
			/*
				sendMsg.Nodes = monitor.Global().Nodes  - nodes
				sendMsg.Orders = monitor.Global().Orders -orders
			*/
			sendMsg := monitor.Global()
			fmt.Println("Sending:", sendMsg.ID)
			ch.MsgTransmitter <- sendMsg

		case p := <-ch.PeerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			localid := strconv.Itoa(id)
			if localid == "" {
				localIP, err := localip.LocalIP()
				if err != nil {
					fmt.Println(err)
					localIP = "DISCONNECTED"
				}
				localid = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
			}

		}
	}

}
