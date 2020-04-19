package sync

import (
	"fmt"
	"os"
	"strconv"
	"time"
	
	"../../pkg/network/localip"
	. "../common/types"
	"../common/config"
	"../monitor"
)

func SyncMessages(ch NetworkChannels, id int) {
	onlineList := make([]bool, config.NElevs)
	bcastTicker := time.NewTicker(500 * time.Millisecond)

	for {
		select {

		case getMsg := <-ch.MsgReceiver:
			ch.UpdateOrders <- getMsg

		case <-bcastTicker.C:
			sendMsg := monitor.Global()
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
