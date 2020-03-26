package watchdog

import (
	"fmt"
	"time"
	"encoding/json"

	/* Setup desc. in main */
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
)

func UpdateWatchdog(newPackets chan packetReceiver){
	select{
		case packet := <- newPackets
				var msg types.GlobalInfo
				err := json.Unmarshal(packet, &msg)
				if err != nil {
					fmt.Println("Error with unmarshaling message in Watchdog:", err)
				}

				types.NodeInfo.ElevLastSent[msg.LocalID] = time.Now()
			}
		}
	}

	for i := 0; i < config.NElevs; i++{
		if (time.Now() - types.NodeInfo.ElevLastSent[i]) < 3 * time.Second(){
			types.NodeInfo.OnlineList[i] = 1
		}
		else{
			types.NodeInfo.OnlineList[i] = 0
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