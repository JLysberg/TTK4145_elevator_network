package cost_estimator

import (
	"time"

	"internal/monitor"
)

const _Queue_UpdateRate = 20 * time.Millisecond

var OrderQueue [monitor.MFloors]bool

//TODO: Improve function to calculate queue with multiple elevators
func UpdateQueue( /*receiver <-chan int*/ ) {
	for {
		time.Sleep(_Queue_UpdateRate)
		for floor, floorState := range monitor.OrderMatrix {
			if floorState.Up || floorState.Down || floorState.Cab {
				OrderQueue[floor] = true
			}
		}
	}
}
