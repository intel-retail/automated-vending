// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package driver

import (
	"fmt"

	common "ds-card-reader/common"
	device "ds-card-reader/device"
	sdk "github.com/edgexfoundry/device-sdk-go"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// CardReaderDriver represents the EdgeX driver that interfaces with the
// underlying device
type CardReaderDriver struct {
	LoggingClient logger.LoggingClient
	CardReader    device.CardReader
	Config        common.CardReaderConfig
}

// Initialize initializes the card reader device within EdgeX. This is the
// main entrypoint of this application
func (drv *CardReaderDriver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues) error {
	// propagate the logging client to the driver so it can use it too
	drv.LoggingClient = lc

	// parse the device's configuration into a proper struct
	err := utilities.MarshalSettings(sdk.DriverConfigs(), &drv.Config, false)
	if err != nil {
		return fmt.Errorf("failed to process card reader settings, check configuration.toml: %w", err)
	}

	// initialize the card reader device so that it can be controlled by our
	// EdgeX driver, and so that it can store configuration values
	drv.CardReader, err = device.InitializeCardReader(
		lc,
		asyncCh,
		drv.Config.DeviceSearchPath,
		drv.Config.DeviceName,
		drv.Config.VID,
		drv.Config.PID,
		drv.Config.SimulateDevice,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize card reader: %w", err)
	}

	return nil
}

// HandleReadCommands is responsible for handling read commands from EdgeX
func (drv *CardReaderDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (result []*dsModels.CommandValue, err error) {
	deviceResourceName := reqs[0].DeviceResourceName

	switch deviceResourceName {
	// the "card reader status" auto event is intended to be a frequent health
	// check that ensures we have a lock on the underlying device
	case common.CommandCardReaderStatus:
		drv.LoggingClient.Debug(fmt.Sprintf("read command: %v, verifying lock on device", common.CommandCardReaderStatus))

		err := drv.CardReader.Status()
		if err != nil {
			errMsg := fmt.Sprintf("read command: %v, failed to verify lock on device: %v", common.CommandCardReaderStatus, err.Error())
			drv.LoggingClient.Error(errMsg)
			return result, fmt.Errorf(errMsg)
		}

		drv.LoggingClient.Debug(fmt.Sprintf("read command: %v, device ok", common.CommandCardReaderStatus))
		return result, nil
	}

	errMsg := fmt.Sprintf("read command \"%v\" is not handled by this device service", deviceResourceName)
	drv.LoggingClient.Error(errMsg)
	return result, fmt.Errorf(errMsg)
}

// HandleWriteCommands implements a standard EdgeX device service function to
// handle incoming EdgeX write commands that come from other services
// connected to EdgeX. Write commands are intended to be used by the virtual
// device service only, but works in either physical or virtual conditions.
func (drv *CardReaderDriver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest, params []*dsModels.CommandValue) error {
	if len(params) == 0 {
		return fmt.Errorf("no params were passed into the write command handler for device %v", deviceName)
	}

	commandName := params[0].DeviceResourceName

	switch commandName {
	case common.CommandCardReaderEvent:
		{
			// parse the card number from the event
			cardNumber, err := params[0].StringValue()
			if err != nil {
				errMsg := fmt.Sprintf("write command \"%v\" received non-string value: %v", commandName, err.Error())
				drv.LoggingClient.Debug(errMsg)
				return fmt.Errorf(errMsg)
			}

			drv.CardReader.Write(common.CommandCardReaderEvent, cardNumber)

			return nil
		}
	}
	errMsg := fmt.Sprintf("write command \"%v\" is not handled by this device service", commandName)
	drv.LoggingClient.Error(errMsg)
	return fmt.Errorf(errMsg)
}

// AddDevice responds to when a device is added.
func (drv *CardReaderDriver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// Nothing to do
	return nil
}

// UpdateDevice responds to when a device is updated.
func (drv *CardReaderDriver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// Nothing to do
	return nil
}

// RemoveDevice responds to when a device is removed.
func (drv *CardReaderDriver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	// Nothing to do
	return nil
}

// Stop allows EdgeX to emulate stopping the device
func (drv *CardReaderDriver) Stop(force bool) error {
	return nil
}

// DisconnectDevice allows EdgeX to emulate disconnection
func (drv *CardReaderDriver) DisconnectDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	return nil
}
