package watchdog

import (
	"fmt"
	"time"
)

func UpdateWatchdog(newPackets chan packetReceiver){
	select{
		case packet := <- newPackets
				var msg GlobalInfo
				err := json.Unmarshal(packet, &msg)
				if err != nil {
					fmt.Println("Error with unmarshaling message in Watchdog:", err)
				}
				
				ElevLastSent[msg.LocalID] = time.Now()
			}
		}
	}

	for i := 0; i < NElevs; i++{
		if (time.Now() - ElevLastSent[i]) < 3 * time.Second(){
			OnlineList[i] = 1
		}
		else{
			OnlineList[i] = 0
		}
	}
}

func AmIOffline(id int){
	Solitude := true
	
	for i := 0; i < NElevs; i++{
		if OnlineList[i] && i != id{
			Solitude = false
		}
	}
	return Solitude
}