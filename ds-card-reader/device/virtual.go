// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"fmt"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	logger "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

// CardReaderVirtual allows for the emulation of a physical card reader device
type CardReaderVirtual struct {
	AsyncCh             chan<- *dsModels.AsyncValues
	DeviceName          string
	LoggingClient       logger.LoggingClient
	MockFailStatusCheck bool // mocks an error message on status()
}

// Write grants the virtual card reader device the ability to respond to
// EdgeX commands to create a card reader "badge-in" event, and push the event
// through the EdgeX framework
func (reader *CardReaderVirtual) Write(commandName string, cardNumber string) {
	// assemble the values that will be propagated throughout the
	// device service
	commandvalue, err := dsModels.NewCommandValueWithOrigin(
		commandName,
		common.ValueTypeString,
		cardNumber,
		time.Now().UnixNano()/int64(time.Millisecond),
	)
	if err != nil {
		reader.LoggingClient.Errorf("error on NewCommandValueWithOrigin for %v: %v", commandName, err)
		return
	}

	result := []*dsModels.CommandValue{
		commandvalue,
	}

	asyncValues := &dsModels.AsyncValues{
		DeviceName:    reader.DeviceName,
		CommandValues: result,
	}

	// push the async value to the async channel, causing an event
	// to be created within EdgeX, which will go to the central
	// as-vending service for processing
	reader.AsyncCh <- asyncValues

	reader.LoggingClient.Info(fmt.Sprintf("received event with card number %v", cardNumber))
}

// Status is a health check function that alays returns true for the virtual
// card reader
func (reader *CardReaderVirtual) Status() error {
	if reader.MockFailStatusCheck {
		return fmt.Errorf("status check failed")
	}
	return nil
}

// Listen is a function that is only needed for the physical card reader,
// but has to be implemented
func (reader *CardReaderVirtual) Listen() {
}

// Release is a function that is only needed for the physical card reader,
// but has to be implemented
func (reader *CardReaderVirtual) Release() error {
	return nil
}
