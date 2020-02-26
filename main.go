package main

import (
	"./internal/fsm"
	"./internal/cost_estimator"
)

func main() {
	cost_update := make(chan int)

	go cost_estimator.UpdateQueue(cost_update)
	go monitor.

	fsm.Run();
}
