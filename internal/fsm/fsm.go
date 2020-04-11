package fsm

import (
	"fmt"
	"time"

	/* Setup desc. in main*/
	/*
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
	*/

	. "../common/types"
	"../monitor"
	"../../pkg/elevio"

)

type StateMachineChannels struct {
	ButtonPress          chan ButtonEvent
	FloorSensor          chan int
	ObstructionSwitch    chan bool
	NewOrder             chan int
	PacketReceiver       chan []byte
	ButtonLights_Refresh chan int
	ClearOrder           chan int
	DoorTimeout			 chan bool
}

func orderInFront() bool {
	for floor, floorState := range monitor.Node.Queue {
		if floorState.Cab || floorState.Up || floorState.Down {
			diff := monitor.Node.Floor - floor
			switch monitor.Node.LastDir {
			case MD_Up:
				if diff < 0 {
					return true
				}
			case MD_Down:
				if diff > 0 {
					return true
				}
			}
		}
	}
	return false
}

func orderAvailable() bool {
	for _, floorState := range monitor.Node.Queue {
		if floorState.Cab || floorState.Up || floorState.Down {
			return true
		}
	}
	return false
}

func calculateDirection() MotorDirection {
	if orderAvailable() {
		if orderInFront() {
			return monitor.Node.LastDir
		}
		return -1 * monitor.Node.LastDir
	}
	return MD_Stop
}

var runningInstance = false
func setNodeDirection(doorOpen <-chan bool) {
	/* Minor delay to allow cost estimator to evaluate orders
	   CONSIDER USING SEMAPHORES */
	if runningInstance {
		return
	}
	runningInstance = true
	time.Sleep(1 * time.Nanosecond)
	
	/* Safety loop to ensure direction is never changed while door is open */
	if monitor.Node.State == ES_Stop {
		safety:
		for {
			select {
			case <-doorOpen:
				break safety
			}
		}
	}
	
	dir := calculateDirection()
	elevio.SetMotorDirection(dir)
	monitor.Node.Dir = dir
	if dir != MD_Stop {
		monitor.Node.LastDir = monitor.Node.Dir
		monitor.Node.State = ES_Run
	}
	runningInstance = false
}

func stopCriteria(floor int) bool {
	floorState := monitor.Node.Queue[floor]
	return floorState.Up && monitor.Node.Dir == MD_Up ||
		floorState.Down && monitor.Node.Dir == MD_Down ||
		(floorState.Down || floorState.Up) && !orderInFront() ||
		floorState.Cab
}

func Printer() {
	for {
		time.Sleep(1 * time.Second)

		for _, floorStates := range monitor.Global.Orders {
			for _, floorState := range floorStates {
		 		fmt.Println("*", floorState)
		 	}
		}
		fmt.Println("#", monitor.Node.Queue)
		fmt.Println()

		fmt.Println("State:", monitor.Node.State)
	}
}

func elev_Init(floorSensor <-chan int) {
	elevio.SetMotorDirection(MD_Down)
	monitor.Node.LastDir = MD_Down
	floor := <-floorSensor
	monitor.Node.Floor = floor
	elevio.SetFloorIndicator(floor)
	elevio.SetMotorDirection(MD_Stop)
}

func Run(ch StateMachineChannels) {
	elev_Init(ch.FloorSensor)
	doorTimeout := time.NewTimer(3 * time.Second)
	doorTimeout.Stop()

	/* NEEDS REFACTORING */
	for {
		select {
		case floor := <-ch.NewOrder:
			switch monitor.Node.State {
			case ES_Stop, ES_Idle:
				if floor == monitor.Node.Floor {
					ch.ClearOrder <- floor
					monitor.Node.State = ES_Stop
					doorTimeout.Reset(3 * time.Second)
					elevio.SetDoorOpenLamp(true)
				} else {
					go setNodeDirection(ch.DoorTimeout)
				}
			case ES_Run:
				go setNodeDirection(ch.DoorTimeout)
			}
		case floor := <-ch.FloorSensor:
			go elevio.SetFloorIndicator(floor)
			monitor.Node.Floor = floor
			if stopCriteria(floor) {
				ch.ClearOrder <- floor
				elevio.SetMotorDirection(MD_Stop)
				monitor.Node.Dir = MD_Stop
				monitor.Node.State = ES_Stop
				doorTimeout.Reset(3 * time.Second)
				elevio.SetDoorOpenLamp(true)
				/* Run thread to set direction after door has timed out */
				go setNodeDirection(ch.DoorTimeout)
			}
		case <-doorTimeout.C:
			elevio.SetDoorOpenLamp(false)
			ch.DoorTimeout <- true
			if !orderAvailable() {
				monitor.Node.State = ES_Idle
			} else {
				monitor.Node.State = ES_Run
			}
		case on := <-ch.ObstructionSwitch:
			if monitor.Node.State == ES_Stop {
				if on {
					doorTimeout.Reset(1000 * time.Second)
				} else {
					doorTimeout.Reset(3 * time.Second)
				}
			}
		}
	}
}
