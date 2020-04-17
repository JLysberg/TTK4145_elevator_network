package watchdog

import (
	"fmt"
	"time"
	"encoding/json"

	// . "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	// "github.com/JLysberg/TTK4145_elevator_network/internal/common/config"

	"../monitor"
	. "../common/types"
	"../common/config"
)


//Take use of the functions in network/peers! 
func UpdateElevLastSent(newPackets chan packetReceiver){
	select{
		case packet := <- newPackets
				types.NodeInfo.ElevLastSent[msg.LocalID] = time.Now()
			}
		}
	}
}

func UpdateOnlineList(newPackets chan packetReceiver){
	for {
		for i := 0; i < config.NElevs; i++{
			if (time.Now() - types.NodeInfo.ElevLastSent[i]) < 3 * time.Second(){
				types.NodeInfo.OnlineList[i] = 1
				if(types.NodeInfo.State - )
					types.NodeInfo,OnlineList[i] = 

					monitor.Node.State
				}
			}
			else{
				types.NodeInfo.OnlineList[i] = 0
			}
		}
	}
}




/*
func AmIOffline(id int){
	Solitude := true

	for i := 0; i < NElevs; i++{
		if OnlineList[i] && i != id{
			Solitude = false
		}
	}
	return Solitude
}*/