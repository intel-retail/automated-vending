// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"fmt"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// CardReader is an abstraction that allows for operations to be performed on
// either a virtual or physical card reader, without explicitly passing around
// a virtual or physical card reader struct
type CardReader interface {
	Write(string, string)
	Status() error
	Listen()
	Release() error
}

// InitializeCardReader gets called by the EdgeX entry point, Initialize(), and
// is responsible for spawning Go routines that await & handle incoming events
//
// simulateDevice dictates whether the device is a physical or virtual card
// reader
//
// mockDevice is used for tests, and dictates whether or not the Listen
// loop goes forever. This should be false unless running unit tests
func InitializeCardReader(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues, deviceSearchPath string, deviceName string, vid uint16, pid uint16, simulateDevice bool, mockDevice bool) (cardReader CardReader, err error) {
	// check if we are configured to only simulate a physical card reader
	// device, or if we are allowed to use an actual physical card reader device
	if !simulateDevice {
		dev, err := GrabCardReader(deviceSearchPath, vid, pid)
		if err != nil {
			return cardReader, fmt.Errorf("failed to grab card reader under search path %v with VID %v and PID %v: %v", deviceSearchPath, vid, pid, err)
		}

		// initialize the physical card reader device
		cardReader = &CardReaderPhysical{
			AsyncCh:          asyncCh,
			CardNumber:       "",
			Device:           dev,
			DeviceName:       deviceName,
			LoggingClient:    lc,
			VID:              vid,
			PID:              pid,
			DeviceSearchPath: deviceSearchPath,
			StableDevice:     true,
		}

		if !mockDevice {
			// spawn a go routine to handle events from the raw device
			go cardReader.Listen()
		}
	} else {
		// initialize the virtual card reader device
		cardReader = &CardReaderVirtual{
			AsyncCh:       asyncCh,
			DeviceName:    deviceName,
			LoggingClient: lc,
		}
	}

	return cardReader, nil
}
