package cost_estimator

import (
	//"fmt"
	"time"

	//. "../common/types"
	"../common/config"
)

const _Queue_UpdateRate = 20 * time.Millisecond

var OrderQueue [config.MFloors]bool

//TODO: Improve function to calculate queue with multiple elevators
/*func UpdateQueue( newOrder <-chan bool ) {
	for {
		time.Sleep(_Queue_UpdateRate)
		for floor, floorState := range monitor.OrderMatrix {
			if floorState.Up || floorState.Down || floorState.Cab {
				OrderQueue[floor] = true
				newOrder <- true
			}
			if floorState.Clear {

			}
		}
	}
}*/
