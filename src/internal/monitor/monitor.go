package monitor

import (
	//"fmt"
	"time"
	/*"encoding/json"

	. "../common/types"
	"../common/config"
	"../../pkg/elevio"*/
)

const (
	_UpdateRate = 20 * time.Millisecond
)

//Make a struct with what needs to be sent; ElevStates and OrderMatrix (I called it ElevStates_OrderMatrix here)
//Make a function that Marshals and transmits the messages through the network

/*IncomingMsg := make(chan ElevStates_OrderMatrix)
go bcast.Receiver(0, IncomingMsg)

func PollOrdersNetwork(packet <-chan IncomingMsg) {
	var msg ElevStates_OrderMatrix
	err := json.Unmarshal(packet, &msg)
	if err != nil {
		fmt.Println("error with unmarshaling message:", err)
	}

	for floors := 0; floors < MFloors; i++ {
		for elevs := 0; elevs < NElevs; j++ {
			if elevs == Elev_id && msg.OrderMatrix[floors][elevs].Clear{
			
			OrderMatrix[floors][elevs].Up = false
			OrderMatrix[floors][elevs].Down = false
			OrderMatrix[floors][elevs].Cab = false
							
			}
			if elevs == Elev_id && msg.OrderMatrix[floors][elevs].Cab{
				OrderMatrix[floors][elevs].Cab = true				
			}	
				
			if elevs == Elev_id && msg.OrderMatrix[floors][elevs].Up{ 
				OrderMatrix[floors][elevs].Up = true
			}
	
			if elevs == Elev_id && msg.OrderMatrix[floors][elevs].Down) {
				OrderMatrix[floors][elevs].Down = true
			}
		}
	}
}
*/