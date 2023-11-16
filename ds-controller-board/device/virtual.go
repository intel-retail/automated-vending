// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"encoding/json"
	"fmt"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	edgexcommon "github.com/edgexfoundry/go-mod-core-contracts/v3/common"
)

// ControllerBoardVirtual is a virtualized controller board that locally handles
// its 'sensor' and 'input' devices in a mocked fashion
type ControllerBoardVirtual struct {
	AsyncCh       chan<- *dsModels.AsyncValues
	DevStatus     string // uC -> Host : STATUS,L1,0,L2,0,D,0,T,78.58,H,19.54
	LoggingClient logger.LoggingClient
	L1            int
	L2            int
	DoorClosed    int
	Temperature   float64
	Humidity      int64
	DeviceName    string
}

// Read : A continuous loop that reads ControllerBoard Status and forwards it to the EdgeX stack as a Reading.
func (board *ControllerBoardVirtual) Read() {
	board.LoggingClient.Debug("Virtual ControllerBoard Read Listening for events ...\n")

	for {
		time.Sleep(3 * time.Second)
		now := time.Now().UnixNano() / int64(time.Millisecond)

		parsedStatus, err := ParseStatus(board.getRawStatus())
		if err != nil {
			board.LoggingClient.Error("unable to parse status", "error", err)
			break
		}

		parsedStatusBytes, err := json.Marshal(parsedStatus)
		if err != nil {
			board.LoggingClient.Error("unable to marshal parsed status", "error", err)
			break
		}

		board.DevStatus = string(parsedStatusBytes)
		commandvalue, err := dsModels.NewCommandValueWithOrigin(
			deviceResource,
			edgexcommon.ValueTypeString,
			board.DevStatus,
			now,
		)
		if err != nil {
			board.LoggingClient.Errorf("error on NewCommandValueWithOrigin for %v: %v", deviceResource, err)
			return
		}

		asyncValues := &dsModels.AsyncValues{
			DeviceName:    board.DeviceName,
			CommandValues: []*dsModels.CommandValue{commandvalue},
		}
		board.AsyncCh <- asyncValues
	}
}

// Write : Used to handle Commands being written to the ControllerBoard.
func (board *ControllerBoardVirtual) Write(cmd string) error {
	board.LoggingClient.Debugf("Write: '%s' command issued.\n", cmd)
	switch cmd {
	case Command.Lock1:
		board.L1 = 1
		board.LoggingClient.Info("Locked Lock1")
	case Command.Lock2:
		board.L2 = 1
		board.LoggingClient.Info("Locked Lock2")
	case Command.UnLock1:
		board.L1 = 0
		board.LoggingClient.Info("Unlocked Lock1")
	case Command.UnLock2:
		board.L2 = 0
		board.LoggingClient.Info("Unlocked Lock2")
	}

	return nil
}

// GetStatus : Returns the ControllerBoard's JSON 'DevStatus' field as a String.
func (board *ControllerBoardVirtual) GetStatus() string {
	return board.DevStatus
}

func (board *ControllerBoardVirtual) getRawStatus() string {
	return fmt.Sprintf("STATUS,L1,%d,L2,%d,D,%d,T,%.2f,H,%d", board.L1, board.L2, board.DoorClosed, board.Temperature, board.Humidity)
}

// SetHumidity allows the controller board to emulate readings of an arbitrary
// humidity value
func (board *ControllerBoardVirtual) SetHumidity(humidity int64) {
	board.Humidity = humidity
}

// SetTemperature allows the controller board to emulate readings of
// an arbitrary temperature value
func (board *ControllerBoardVirtual) SetTemperature(temperature float64) {
	board.Temperature = temperature
}

// SetDoorClosed allows the controller board to emulate the condition that the
// door is closed
func (board *ControllerBoardVirtual) SetDoorClosed(closed int) {
	board.DoorClosed = closed
}
