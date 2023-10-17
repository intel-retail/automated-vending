// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"fmt"

	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

// ControllerBoard is a common interface for controller boards to implement
type ControllerBoard interface {
	Read()
	Write(string) error
	GetStatus() string
}

// NewControllerBoard is used to determine the ControllerBoard type
// (i.e. Device or Virtual), and perform the necessary steps to initialize
// a new ControllerBoard
func NewControllerBoard(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues, config *Config) (ControllerBoard, error) {
	var controllerBoard ControllerBoard

	if !config.VirtualControllerBoard {

		// Find the port name (like /dev/ttyACM0) connected to Controller Board
		ttyPort, err := FindControllerBoard(config.VID, config.PID)
		if err != nil {
			return nil, fmt.Errorf("can't find controller board, check if it is connected: %s", err.Error())
		}

		devSerialPort, err := OpenAndConfigureSerialPort(ttyPort)
		if err != nil {
			return nil, fmt.Errorf("can't open or configure serial port %s: %s", ttyPort, err.Error())
		}

		lc.Info(fmt.Sprintf("Successfully opened and configured controller board on %s", ttyPort))

		controllerBoard = &ControllerBoardPhysical{
			AsyncCh:       asyncCh,
			DevStatus:     "",
			LoggingClient: lc,
			DevSerialPort: devSerialPort,
			TTYPort:       ttyPort,
			DeviceName:    config.DeviceName,
		}
	} else {
		controllerBoard = &ControllerBoardVirtual{
			AsyncCh:       asyncCh,
			DevStatus:     "",
			LoggingClient: lc,
			L1:            1,
			L2:            1,
			DoorClosed:    1,
			Temperature:   78.00,
			Humidity:      10,
			DeviceName:    config.DeviceName,
		}
	}

	return controllerBoard, nil
}
