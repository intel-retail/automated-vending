// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"ds-card-reader/common"
	"fmt"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	logger "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	edgexcommon "github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	evdev "github.com/gvalkov/golang-evdev"
)

// CardReaderPhysical follows some EdgeX standards for implementing a device
// that creates events & readings in an EdgeX device service.
type CardReaderPhysical struct {
	AsyncCh          chan<- *dsModels.AsyncValues
	Device           *evdev.InputDevice
	DeviceName       string
	DeviceSearchPath string
	LoggingClient    logger.LoggingClient
	PID              uint16
	VID              uint16
	CardNumber       string
	Mocked           bool // used for unit testing
	StableDevice     bool
}

// Listen is called by the EdgeX driver for this device service, it listens
// for raw input from the device itself and facilitates relaying the data
// back up to EdgeX. It is intended to be called as a Go routine.
// This function is not feasibly unit tested because it relies on a live
// feed of events from an input device. It is best to perform integration
// tests instead, by simply unplugging & plugging in the device
// and observing graceful handling
func (reader *CardReaderPhysical) Listen() {
	for {
		if reader.StableDevice {
			// warning: the below line will panic if the device is not first
			// initialized. However, this function should never be reached if
			// that is the case, because the initialization sequence is done
			// upon startup
			events, err := reader.Device.Read()
			if err != nil {
				reader.LoggingClient.Error(fmt.Sprintf("device read event error: %v", err.Error()))
				reader.StableDevice = false
				continue
			}
			reader.processDevReadEvents(events)
		} else {
			// needed because go syntax will not allow := below
			var err error

			reader.Device, err = GrabCardReader(reader.DeviceSearchPath, reader.VID, reader.PID)
			if err != nil {
				reader.LoggingClient.Error(fmt.Sprintf("failed to re-grab card reader device, retry in 3s: %v", err.Error()))
				time.Sleep(3 * time.Second)
				continue
			}

			reader.LoggingClient.Info("successfully re-grabbed device")
			reader.StableDevice = true
		}
	}
}

// Status is called frequently by the driver (as an auto-event) to check the
// "grab" state of the card reader device
func (reader *CardReaderPhysical) Status() error {
	// if the ReadValues thread is working properly, the GrabCardReader
	// function should actually fail, hence why err == nil is checked
	_, err := GrabCardReader(reader.DeviceSearchPath, reader.VID, reader.PID)
	if err == nil {
		errMsg := "failure: physical card reader is not locked"
		reader.LoggingClient.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	return nil
}

// Write grants the physical card reader device the ability to respond to
// EdgeX commands to create a card reader "badge-in" event, and push the event
// through the EdgeX framework. Both the physical and virtual card reader
// have this code, and as a result, they both can respond to REST API calls
// to "badge-in", but only the physical card reader is listening for input
// events from a real device
func (reader *CardReaderPhysical) Write(commandName string, cardNumber string) {
	// assemble the values that will be propagated throughout the
	// device service
	commandvalue, err := dsModels.NewCommandValueWithOrigin(
		commandName,
		edgexcommon.ValueTypeString,
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
	// Note: this will loop indefinitely in unit testing, so we have to respect
	// a mocked environment
	if !reader.Mocked {
		reader.AsyncCh <- asyncValues
	}

	// reset the card number after pushing to the async channel
	reader.CardNumber = ""

	reader.LoggingClient.Info(fmt.Sprintf("received event with card number %v", cardNumber))
}

// processDevReadEvents handles an incoming evdev device event and pushes it
// into the EdgeX framework
func (reader *CardReaderPhysical) processDevReadEvents(events []evdev.InputEvent) {
	for i := range events {
		// process the input event from the evdev device
		keyValue, err := GetKeyValueFromEvent(&events[i])
		if err != nil {
			continue
		}

		// check if the key that was pressed is the enter key
		if keyValue != evdev.KEY_ENTER {
			reader.CardNumber = reader.CardNumber + fmt.Sprintf("%v", keyValue)
			continue
		}

		reader.Write(common.CommandCardNumber, reader.CardNumber)
	}
}

// GetKeyValueFromEvent is responsible for handling key stroke events
// from the physical card reader and ensuring that only a digit/keystroke we
// care about is returned (note that it does not necessarily return the
// corresponding evdev key code value)
//
// for reference, please visit:
// https://github.com/gvalkov/golang-evdev/blob/master/ecodes.go
func GetKeyValueFromEvent(ev *evdev.InputEvent) (result int, err error) {
	// check if the event type is a key stroke
	if ev.Type == evdev.EV_KEY && ev.Value == int32(evdev.KeyDown) {
		switch ev.Code {
		case evdev.KEY_1:
			return 1, nil
		case evdev.KEY_2:
			return 2, nil
		case evdev.KEY_3:
			return 3, nil
		case evdev.KEY_4:
			return 4, nil
		case evdev.KEY_5:
			return 5, nil
		case evdev.KEY_6:
			return 6, nil
		case evdev.KEY_7:
			return 7, nil
		case evdev.KEY_8:
			return 8, nil
		case evdev.KEY_9:
			return 9, nil
		case evdev.KEY_0:
			return 0, nil
		case evdev.KEY_ENTER:
			return int(evdev.KEY_ENTER), nil
		default:
			return result, fmt.Errorf("received an undesired or unexpected key code: %v", ev.Code)
		}
	}

	return result, fmt.Errorf("event created by the physical card reader was not a key press event")
}

// findInputDevice returns the path of an input device under a given
// (bash globstar) path, with a corresponding vendor ID and product ID
//
// the Linux command "lsusb" can help identify VID and PID values
func findInputDevice(path string, VID uint16, PID uint16) (string, error) {
	// get a list of input devices by scanning
	devices, _ := evdev.ListInputDevices(path)

	// see if the card reader shows up in the list of devices in evdev
	for _, dev := range devices {
		if dev.Product == PID && dev.Vendor == VID {
			// return the file path of the found input device
			return dev.Fn, nil
		}
	}

	return "", fmt.Errorf("failed to find card reader under path \"%v\" with VID %v and PID %v", path, VID, PID)
}

// GrabCardReader attempts to grab the physical card reader device.
// Any time the device hold is lost, it will simply type badge numbers to the
// focused window (it basically behaves like a regular keyboard).
// Depending on circumstances, it may be necessary to run dev.Release() after
//
// The searchPath parameter is a bash globstar expression that will be passed
// to evdev so that it knows where to look for the input device that represents
// the card reader.
func GrabCardReader(searchPath string, VID uint16, PID uint16) (dev *evdev.InputDevice, err error) {
	// attempt to find the reader in the given search path
	deviceFilePath, err := findInputDevice(searchPath, VID, PID)
	if err != nil {
		return dev, fmt.Errorf("unable to find the card reader: %w", err)
	}

	// attempt to open the card reader device, should succeed if it exists
	dev, err = evdev.Open(deviceFilePath)
	if err != nil {
		return dev, fmt.Errorf("unable to open the card reader: %w", err)
	}

	// attempting to grab device should fail in normal conditions, because it
	// should already be locked from another go routine (ReadValues function)
	err = dev.Grab()
	if err != nil {
		return dev, fmt.Errorf("failed to grab the card reader device: %w", err)
	}

	return dev, nil
}

// Release attempts to release the grab on the device
func (reader *CardReaderPhysical) Release() error {
	return reader.Device.Release()
}
