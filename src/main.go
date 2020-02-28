package main

import (
	"internal/cost_estimator"
	"internal/fsm"
	"internal/monitor"
)

func main() {
	//cost_update := make(chan int)

	go cost_estimator.UpdateQueue( /*cost_update*/ )
	go monitor.UpdateButtonLights()

	fsm.Run()
}
