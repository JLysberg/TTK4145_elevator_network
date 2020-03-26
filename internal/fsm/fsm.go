package fsm

import (
	"fmt"
	"time"

	/* Setup desc. in main*/
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
)

type StateMachineChannels struct {
	ButtonPress          chan ButtonEvent
	FloorSensor          chan int
	ObstructionSwitch    chan bool
	NewOrder             chan bool
	PacketReceiver       chan []byte
	ButtonLights_Refresh chan int
	ClearOrder          chan int
}

func orderInFront() bool {
	for floor, order := range monitor.Node.Queue {
		if order {
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
	for _, order := range monitor.Node.Queue {
		if order {
			return true
		}
	}
	return false
}

func calculateDirection() MotorDirection {
	if orderAvailable() {
		if orderInFront() {
			return monitor.Node.LastDir
		} else {
			return -1 * monitor.Node.LastDir
		}
	} else {
		return MD_Stop
	}
}

func setNodeDirection() {
	/* Minor delay to allow cost estimator to evaluate orders 
	   CONSIDER USING SEMAPHORES */
	time.Sleep(1 * time.Nanosecond)
	dir := calculateDirection()
	elevio.SetMotorDirection(dir)
	monitor.Node.Dir = dir
	if dir != MD_Stop {
		monitor.Node.LastDir = monitor.Node.Dir
	}
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
	for {
		select {
		case <-ch.NewOrder:
			switch monitor.Node.State {
			case ES_Stop:
				fallthrough
			case ES_Idle:
				//TODO: jdfksjldfk
			case ES_Run:
				//TODO
			}
			setNodeDirection()
		case floor := <-ch.FloorSensor:
			elevio.SetFloorIndicator(floor)
			monitor.Node.Floor = floor
			if monitor.Node.Queue[floor] {
				ch.ClearOrder <- floor
				setNodeDirection()
			}
		}
	}
}

//TODO: Add door/obstruction timer
//TODO: Network
