package node

import (
	"fmt"
	"time"
	
	"../common/config"
	. "../common/types"
	"../../pkg/elevio"
)

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
	watchdog := time.NewTimer(10 * time.Second)
	/*	Initialize to known state */
	go elevio.SetMotorDirection(MD_Down)
	arrivedFloor := <-ch.FloorSensor
	go elevio.SetMotorDirection(MD_Stop)
	go elevio.SetFloorIndicator(arrivedFloor)
	go elevio.SetDoorOpenLamp(false)
	local.Floor = arrivedFloor
	/*	Initialize monitor */
	ch.UpdateLocal <- local

	for {
		watchdog.Reset(10 * time.Second)
		select {
		case getLocalCopy <- local:
		case queueCopy = <-ch.UpdateQueue:
			switch local.State {
			case ES_Stop, ES_Idle:
				/* Check for orders in queueCopy */
				if queueCopy[local.Floor].Up || 
				   queueCopy[local.Floor].Down ||
				   queueCopy[local.Floor].Cab {
					floorStop(local.Floor, ch.SetClearBit)
					local.State = ES_Stop
					local.Dir = MD_Stop
				} else {
					go setDirection(ch.DoorClose, queueCopy)
				}
			case ES_Run:
				go setDirection(ch.DoorClose, queueCopy)
			}

		case arrivedFloor := <-ch.FloorSensor:
			go elevio.SetFloorIndicator(arrivedFloor)
			local.Floor = arrivedFloor
			/* Stop the elevator if there is an order at arrivedFloor */
			if stopCriteria(arrivedFloor, local, queueCopy) {
				floorStop(arrivedFloor, ch.SetClearBit)
				local.State = ES_Stop
				local.Dir = MD_Stop
				go setDirection(ch.DoorClose, queueCopy)
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

		case <-doorTimeout.C:
			elevio.SetDoorOpenLamp(false)
			ch.DoorClose <- true
			if !orderAvailable(queueCopy) {
				local.State = ES_Idle
			} else {
				local.State = ES_Run
			}
		
		case <-watchdog.C:
			if local.State == ES_Run {
				fmt.Println("Error: No node activity")
				/*	Set elevator in error state */
				local.State = ES_Error
				ch.UpdateLocal <- local
				/*	Reset elevator if activity is resumed */
				arrivedFloor := <-ch.FloorSensor
				go elevio.SetFloorIndicator(arrivedFloor)
				floorStop(arrivedFloor, ch.SetClearBit)
				local.State = ES_Stop
				local.Dir = MD_Stop
				go setDirection(ch.DoorClose, queueCopy)
			}
		}
		ch.UpdateLocal <- local
	}
}
