package node

import (
	"time"

	/* Setup desc. in main*/
	"github.com/JLysberg/TTK4145_elevator_network/internal/common/config"
	. "github.com/JLysberg/TTK4145_elevator_network/internal/common/types"
	"github.com/JLysberg/TTK4145_elevator_network/internal/monitor"
	"github.com/JLysberg/TTK4145_elevator_network/pkg/elevio"
)

// func Printer() {
// 	for {
// 		time.Sleep(1 * time.Millisecond)

// 		fmt.Println("State:", monitor.Local.State)
// 	}
// }

var doorTimeout = time.NewTimer(1 * time.Hour)

/*	NodeServer handles all logic and communications between local
	routines. */
func NodeServer(ch NodeChannels) {
	initialize(ch.FloorSensor)

	for {
		select {
		case floor := <-ch.NewOrder:
			switch monitor.Local.State {
			case ES_Stop, ES_Idle:
				if floor == monitor.Local.Floor {
					floorStop(floor, ch.ClearOrder)
				} else {
					go setDirection(ch.DoorTimeout)
				}
			case ES_Run:
				go setDirection(ch.DoorTimeout)
			}

		case floor := <-ch.FloorSensor:
			go elevio.SetFloorIndicator(floor)
			monitor.Local.Floor = floor
			if stopCriteria(floor) {
				floorStop(floor, ch.ClearOrder)
				go setDirection(ch.DoorTimeout)
			}

		case <-doorTimeout.C:
			elevio.SetDoorOpenLamp(false)
			ch.DoorTimeout <- true
			if !orderAvailable() {
				monitor.Local.State = ES_Idle
			} else {
				monitor.Local.State = ES_Run
			}

		case enabled := <-ch.ObstructionSwitch:
			if monitor.Local.State == ES_Stop {
				if enabled {
					doorTimeout.Reset(1 * time.Hour)
				} else {
					doorTimeout.Reset(config.DoorTimeout)
				}
			}
		}
	}
}
