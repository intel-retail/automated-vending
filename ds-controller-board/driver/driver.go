// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package driver

import (
	"fmt"
	"strconv"
	"time"

	"ds-controller-board/device"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	service "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

const (
	lock1          = "lock1"
	lock2          = "lock2"
	getStatus      = "getStatus"
	displayRow0    = "displayRow0"
	displayRow1    = "displayRow1"
	displayRow2    = "displayRow2"
	displayRow3    = "displayRow3"
	displayReset   = "displayReset"
	setHumidity    = "setHumidity"
	setTemperature = "setTemperature"
	setDoorClosed  = "setDoorClosed"
)

// ControllerBoardDriver follows EdgeX standards for a device struct.
type ControllerBoardDriver struct {
	lc              logger.LoggingClient
	StopChannel     chan int
	controllerBoard device.ControllerBoard
	config          *device.Config
}

// NewControllerBoardDeviceDriver allows EdgeX to initialize the
// ControllerBoardDriver instance
func NewControllerBoardDeviceDriver() dsModels.ProtocolDriver {
	return new(ControllerBoardDriver)
}

// Initialize is an EdgeX function that initializes the device
func (drv *ControllerBoardDriver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues, deviceCh chan<- []dsModels.DiscoveredDevice) (err error) {
	drv.lc = lc

	// Only setting if nil allows for unit testing with VirtualBoard enabled
	if drv.config == nil {
		drv.config = new(device.Config)
		if err = utilities.MarshalSettings(service.DriverConfigs(), drv.config, true); err != nil {
			return err
		}
	}

	drv.StopChannel = make(chan int)

	drv.controllerBoard, err = device.NewControllerBoard(lc, asyncCh, drv.config)
	if err != nil {
		return err
	}

	go drv.controllerBoard.Read()

	return nil
}

// HandleReadCommands handles AutoEvents and other read events from EdgeX.
func (drv *ControllerBoardDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	err = drv.controllerBoard.Write(device.Command.GetStatus)
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixNano() / int64(time.Millisecond)
	result := dsModels.NewStringValue(reqs[0].DeviceResourceName, now, drv.controllerBoard.GetStatus())

	return []*dsModels.CommandValue{result}, nil
}

// HandleWriteCommands handles incoming write commands from EdgeX.
func (drv *ControllerBoardDriver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest, params []*dsModels.CommandValue) error {

	deviceType := params[0].DeviceResourceName
	drv.lc.Info(fmt.Sprintf("---Received PUT Command: '%s'", deviceType))
	switch deviceType {

	case lock1:
		cmdType := params[0].NumericValue[0]
		switch cmdType {
		case 1:
			if err := drv.controllerBoard.Write(device.Command.UnLock1); err != nil {
				return err
			}
			go func() {
				time.Sleep(drv.config.LockTimeout)
				_ = drv.controllerBoard.Write(device.Command.Lock1)
			}()
		case 0:
			if err := drv.controllerBoard.Write(device.Command.Lock1); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown Command Type: '%d'", cmdType)
		}

	case lock2:
		cmdType := params[0].NumericValue[0]
		switch cmdType {
		case 1:
			if err := drv.controllerBoard.Write(device.Command.UnLock2); err != nil {
				return err
			}
			go func() {
				time.Sleep(drv.config.LockTimeout)
				_ = drv.controllerBoard.Write(device.Command.Lock2)
			}()
		case 0:
			if err := drv.controllerBoard.Write(device.Command.Lock2); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown Command Type: '%d'", cmdType)
		}

	case getStatus:
		if err := drv.controllerBoard.Write(device.Command.Status); err != nil {
			return err
		}

	case displayRow0:

		cmdType, _ := params[0].StringValue()
		message := device.Command.Message0 + cmdType + "\n"
		drv.displayText(message)

	case displayRow1:

		cmdType, _ := params[0].StringValue()
		message := device.Command.Message1 + cmdType + "\n"
		drv.displayText(message)

	case displayRow2:

		cmdType, _ := params[0].StringValue()
		message := device.Command.Message2 + cmdType + "\n"
		drv.displayText(message)

	case displayRow3:

		cmdType, _ := params[0].StringValue()
		message := device.Command.Message3 + cmdType + "\n"
		drv.displayText(message)

	case displayReset:
		drv.displayReset()

	case setHumidity:
		cmdType, _ := params[0].StringValue()
		newHumidity, _ := strconv.ParseInt(cmdType, 10, 64)
		_, ok := (drv.controllerBoard).(*device.ControllerBoardVirtual)
		if ok {
			drv.lc.Info(fmt.Sprintf("ControllerBoardVirtual: 'setHumidity' is being set to %d%% Humidity.", newHumidity))
			drv.controllerBoard.(*device.ControllerBoardVirtual).SetHumidity(newHumidity)
		} else {
			drv.lc.Error("Command 'setHumidity' is only available to Virtual ControllerBoard.")
		}

	case setTemperature:
		cmdType, _ := params[0].StringValue()
		newTemperature, _ := strconv.ParseFloat(cmdType, 64)
		_, ok := (drv.controllerBoard).(*device.ControllerBoardVirtual)
		if ok {
			drv.lc.Info(fmt.Sprintf("ControllerBoardVirtual: 'setTemperature' is being set to %.2f degrees Fahrenheit.", newTemperature))
			drv.controllerBoard.(*device.ControllerBoardVirtual).SetTemperature(newTemperature)
		} else {
			drv.lc.Error("Command 'setTemperature' is only available to Virtual ControllerBoard.")
		}

	case setDoorClosed:
		cmdType, _ := params[0].StringValue()
		newDoorClosed, _ := strconv.ParseInt(cmdType, 10, 64)
		_, ok := (drv.controllerBoard).(*device.ControllerBoardVirtual)
		if ok {
			drv.lc.Info(fmt.Sprintf("ControllerBoardVirtual: 'setDoorClosed' is being set to %t.", !(newDoorClosed == 0)))
			drv.controllerBoard.(*device.ControllerBoardVirtual).SetDoorClosed(int(newDoorClosed))
		} else {
			drv.lc.Error("Command 'setDoorClosed' is only available to Virtual ControllerBoard.")
		}

	default:
		return fmt.Errorf("unknown command received: '%s'", deviceType)
	}

	return nil
}

func (drv *ControllerBoardDriver) displayText(message string) {
	drv.lc.Info(message)
	_ = drv.controllerBoard.Write(message)
	// Stop the display reset thread and restart the timeout
	close(drv.StopChannel)
	drv.StopChannel = make(chan int)

	go func() {
		for {
			select {
			case <-time.After(drv.config.DisplayTimeout):
				drv.displayReset()
				return
			case <-drv.StopChannel:
				drv.lc.Info("Reset the display reset thread")
				return
			}
		}
	}()
}

func (drv *ControllerBoardDriver) displayReset() {
	_ = drv.controllerBoard.Write(device.Command.Message0 + "                   " + "\n")
	_ = drv.controllerBoard.Write(device.Command.Message1 + "                   " + "\n")
	_ = drv.controllerBoard.Write(device.Command.Message2 + "                   " + "\n")
	_ = drv.controllerBoard.Write(device.Command.Message3 + "                   " + "\n")
	_ = drv.controllerBoard.Write(device.Command.DefaultDisp)
}

// AddDevice responds to when a device is added.
func (drv *ControllerBoardDriver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// Nothing to do
	return nil
}

// UpdateDevice responds to when a device is updated.
func (drv *ControllerBoardDriver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// Nothing to do
	return nil
}

// RemoveDevice responds to when a device is removed.
func (drv *ControllerBoardDriver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	// Nothing to do
	return nil
}

// Stop stops a device
func (drv *ControllerBoardDriver) Stop(force bool) error {
	return nil
}
