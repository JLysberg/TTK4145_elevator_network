package cost_estimator

import (
	"fmt"
	"time"
	"../monitor"

	"../../pkg/elevio"
)

const _Queue_UpdateRate = 20 * time.Millisecond

var orderQueue [monitor.NumFloors]bool

func UpdateQueue(receiver <-chan int) {
	fmt.Println("test")

	for {
		time.Sleep(_Queue_UpdateRate)
		for _, element := monitor.OrderMatrix {
			if (element.Up or element.Down or element.Cab) {
				
			}
		}
	}
}
