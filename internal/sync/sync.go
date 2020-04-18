package sync

import (
	"fmt"
	"os"
	"strconv"
	"time"

	// . "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	// "github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	// "github.com/JLysberg/TTK4145_elevator_network/pkg/network/localip"
	// "github.com/JLysberg/TTK4145_elevator_network/pkg/network/peers"
	
	"../../pkg/network/localip"
	"../../pkg/network/peers"
	. "../common/types"
	"../common/config"
	"../monitor"
	)

/*func OnlineList() []bool {
	return <-getQueueCopy
}*/

type NetworkChannels struct {
	MsgTransmitter chan GlobalInfo
	MsgReceiver    chan GlobalInfo
	PeerUpdate     chan peers.PeerUpdate
	PeerTxEnable   chan bool
	UpdateOrders   chan GlobalInfo
	OnlineElevators chan []bool
}

func SyncMessages(ch NetworkChannels, id int) {
	var (
//		lostID 		 int
//		newID		 int
		//ElevLastSent [config.NElevs]int
	)
	onlineList := make([]bool, config.NElevs)

	bcastTicker := time.NewTicker(500 * time.Millisecond)
	//bcastTicker := time.NewTicker(2 * time.Second)

	for {
		select {

		case getMsg := <-ch.MsgReceiver:
			ch.UpdateOrders <- getMsg

		case <-bcastTicker.C:
			// fmt.Println("Broadcasting message")
			sendMsg := monitor.Global()
			// fmt.Println("Sending:", sendMsg.ID)
			ch.MsgTransmitter <- sendMsg

		case p := <-ch.PeerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			fmt.Println(len(p.New))
			localid := strconv.Itoa(id)
			if localid == "" {
				localIP, err := localip.LocalIP()
				if err != nil {
					fmt.Println(err)
					localIP = "DISCONNECTED"
				}
				localid = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
			}

			if len(p.New) > 0 {
				newID, _ := strconv.Atoi(p.New)
				onlineList[newID] = true
			} else if len(p.Lost) > 0 {
				lostID, _ := strconv.Atoi(p.Lost[0])
				onlineList[lostID] = false
			}
			fmt.Println("onlineList: ", onlineList)
			tmpList := onlineList
			ch.OnlineElevators <- tmpList 
		}
	}
}
