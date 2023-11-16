// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	edgexcommon "github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	"go.bug.st/serial.v1"
	"go.bug.st/serial.v1/enumerator"
)

const (
	deviceResource = "controller-board-status"
)

// ControllerBoardPhysical : A physical ControllerBoard that leverages an Arduino Micro-Controller for reading/writing sensor and input values.
type ControllerBoardPhysical struct {
	AsyncCh       chan<- *dsModels.AsyncValues
	DevStatus     string // uC -> Host : STATUS,L1,0,L2,0,D,0,T,78.58,H,19.54
	LoggingClient logger.LoggingClient
	DevSerialPort serial.Port
	TTYPort       string // typically is /dev/ttyACM0
	DeviceName    string
}

// Write is used to handle commands being written to the
// controller board. It also forwards commands to the controller board via
// its serial interface.
func (board *ControllerBoardPhysical) Write(cmd string) error {
	n, err := board.DevSerialPort.Write([]byte(cmd))
	if err != nil {
		return err
	}
	board.LoggingClient.Debug("Sent %v bytes\n", n)
	return nil
}

// Read is a continuous loop that reads the controller board's
// serial status messages and forwards them to EdgeX as a reading
func (board *ControllerBoardPhysical) Read() {
	board.LoggingClient.Debug("ControllerBoard Read Listening for events ...\n")

	var serialString string

	buff := make([]byte, 100)
	for {
		n, err := board.DevSerialPort.Read(buff)
		if err != nil {
			board.LoggingClient.Error(fmt.Sprintf("Error reading from serial port '%s'", board.TTYPort), "error", err)
			break
		}

		if n == 0 {
			board.LoggingClient.Warn(fmt.Sprintf("Read: No data read from serial port '%s'", board.TTYPort))
			break
		}

		board.LoggingClient.Info(fmt.Sprintf("Read: '%s' status read from ControllerBoard.", string(buff[:n])))

		// Must build up the complete string read until all of it is received
		serialString = serialString + string(buff[:n])
		if !strings.ContainsAny(serialString, "\r\n") {
			continue
		}

		board.LoggingClient.Info(fmt.Sprintf("Read: Processing '%s'", serialString))

		if strings.Contains(serialString, "STATUS") {
			parsedStatus, err := ParseStatus(serialString)
			if err != nil {
				board.LoggingClient.Error("unable to parse status", "error", err)
				break
			}

			parsedStatusBytes, err := json.Marshal(parsedStatus)
			if err != nil {
				board.LoggingClient.Error("unable to marshal parsed status", "error", err)
				break
			}

			board.DevStatus = string(parsedStatusBytes) //string(buff[:n])
			now := time.Now().UnixNano() / int64(time.Millisecond)

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

		serialString = ""
	}
}

// GetStatus : Returns the ControllerBoard's JSON 'DevStatus' field as a String.
func (board *ControllerBoardPhysical) GetStatus() string {
	return board.DevStatus
}

// FindControllerBoard : Finds the ControllerBoards TTY URI (e.g. /dev/ttyACM0) based off of its PID and VID values.
func FindControllerBoard(vid string, pid string) (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", err
	}

	for _, port := range ports {
		if port.IsUSB {
			if port.VID == vid && port.PID == pid {
				return port.Name, nil
			}
		}
	}

	return "", fmt.Errorf("no USB port found matching VID=%s & PID=%s", vid, pid)
}

// OpenAndConfigureSerialPort : Opens the TTY URI (e.g. /dev/ttyACM0) as a Serial connection with the appropriate configuration (e.g. baud-rate, parity, data-bits, stop-bits, etc.).
func OpenAndConfigureSerialPort(ttyPort string) (serial.Port, error) {
	port, err := serial.Open(ttyPort, &serial.Mode{})

	if err != nil {
		return nil, err
	}

	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	if err := port.SetMode(mode); err != nil {
		return nil, err
	}

	return port, nil
}
