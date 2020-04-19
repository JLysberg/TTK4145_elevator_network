package node

import (
	"fmt"
	"time"

	// "fmt"

	// "github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	// . "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	// "github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	// "github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
	
	"../common/config"
	. "../common/types"
	"../monitor"
	"../../pkg/elevio")

func Printer() {
	for {
		time.Sleep(1 * time.Second)

		for _, floorStates := range monitor.Global().Orders {
			for _, floorState := range floorStates {
				fmt.Println("*", floorState)
			}
		}
		// fmt.Println("#", monitor.Queue())
		fmt.Println()
	}
}

/*	Local gives a call to ElevatorServer to return a copy of local */
func Local() LocalInfo {
	return <-getLocalCopy
}

var doorTimeout = time.NewTimer(1 * time.Hour)

/*	ElevatorServer handles all elevator logic and communications between local
	routines in current node. */
func ElevatorServer(ch NodeChannels) {
	/*	Declare local variables */
	local := LocalInfo{
		State:   ES_Idle,
		Dir:     MD_Stop,
		LastDir: MD_Down,
	}
	var queueCopy []FloorState
	/*	Initialize */
	elevio.SetMotorDirection(MD_Down)
	floor := <-ch.FloorSensor
	elevio.SetMotorDirection(MD_Stop)
	local.Floor = floor
	elevio.SetFloorIndicator(floor)
	ch.UpdateLocal <- local

	for {
		select {
		case queueCopy = <-ch.UpdateQueue:
			switch local.State {
			case ES_Stop, ES_Idle:
				if queueCopy[local.Floor].Up ||
					queueCopy[local.Floor].Down ||
					queueCopy[local.Floor].Cab {
					floorStop(local.Floor, ch.SetClearBit)
					local.State = ES_Stop
					local.Dir = MD_Stop
				} else {
					go setDirection(ch.DoorOpen, queueCopy)
				}
			case ES_Run:
				go setDirection(ch.DoorOpen, queueCopy)
			}

		case arrivedFloor := <-ch.FloorSensor:
			go elevio.SetFloorIndicator(arrivedFloor)
			local.Floor = arrivedFloor
			if stopCriteria(arrivedFloor, local, queueCopy) {
				floorStop(arrivedFloor, ch.SetClearBit)
				local.State = ES_Stop
				local.Dir = MD_Stop
				go setDirection(ch.DoorOpen, queueCopy)
			}

		case <-doorTimeout.C:
			elevio.SetDoorOpenLamp(false)
			ch.DoorOpen <- true
			if !orderAvailable(queueCopy) {
				local.State = ES_Idle
			} else {
				local.State = ES_Run
			}

		case switchEnabled := <-ch.ObstructionSwitch:
			if local.State == ES_Stop {
				if switchEnabled {
					doorTimeout.Reset(1 * time.Hour)
				} else {
					doorTimeout.Reset(config.DoorTimeout)
				}
			}

		case dir := <-setLocalDir:
			local.Dir = dir
			if dir != MD_Stop {
				local.LastDir = dir
				local.State = ES_Run
			}

		case getLocalCopy <- local:
		}
		ch.UpdateLocal <- local
	}
}
