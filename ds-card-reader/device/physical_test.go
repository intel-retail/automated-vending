// +build all physical

// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	evdev "github.com/gvalkov/golang-evdev"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

const (
	physicalLogFile                     = "physical_test.log"
	physicalCardReaderDeviceServiceName = "ds-card-reader"
	physicalDeviceSearchPath            = "/dev/input/event*"
	physicalInvalidDeviceSearchPath     = "/dev/input/invalid-device*"
	physicalDeviceName                  = "ds-card-reader"
	physicalVID                         = uint16(0xFFFF)
	physicalPID                         = uint16(0x0035)
	expectedCardNumberPhysical          = "0123456789"
)

func clearLogs() error {
	return ioutil.WriteFile(physicalLogFile, []byte{}, 0644)
}

func doesLogFileContainString(input string) (bool, error) {
	// attempt to open the log file
	file, err := os.Open(physicalLogFile)
	if err != nil {
		return false, err
	}

	// defer closing the file open buffer
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// iterate over every line in the log file
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, input) {
			return true, nil
		}
	}

	// check for residual errors in the scanner
	err = scanner.Err()
	if err != nil {
		return false, err
	}

	return false, nil
}

// TestInitializeCardReader validates that the InitializeCardReader
// and the physical functions work as expected for a physical card reader
func TestInitializeCardReader(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	// create a few essential variables for facilitating tests
	lc := logger.NewClient(physicalCardReaderDeviceServiceName, false, physicalLogFile, "DEBUG")
	expectedAsyncCh := make(chan<- *dsModels.AsyncValues, 16)
	var notExpectedCardReader CardReader

	// run the function
	cardReader, err := InitializeCardReader(
		lc,
		expectedAsyncCh,
		physicalDeviceSearchPath,
		physicalDeviceName,
		physicalVID,
		physicalPID,
		false,
		true,
	)

	// perform assertions
	require.NoError(err)
	assert.NotEqual(notExpectedCardReader, cardReader)

	// release the device so that other testing routines can use it
	err = cardReader.Release()
	require.NoError(err)
}

// TestGrabCardReader validates that the GrabCardReader function handles
// acquiring (or failing to acquire) a lock on the input event device
func TestGrabCardReader(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	// build tests
	tests := []struct {
		Name           string
		VID            uint16
		PID            uint16
		searchPath     string
		expectedErrMsg string
	}{
		{
			"failure to find card reader in search path",
			physicalVID,
			physicalPID,
			physicalInvalidDeviceSearchPath,
			"unable to find the card reader:",
		},
		{
			"failure to find card by VID/PID",
			uint16(0xABCD),
			uint16(0xABCD),
			physicalDeviceSearchPath,
			"unable to find the card reader:",
		},
		{
			"successfully grabbed card reader",
			physicalVID,
			physicalPID,
			physicalDeviceSearchPath,
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			dev, err := GrabCardReader(test.searchPath, test.VID, test.PID)
			if err == nil {
				assert.Empty(test.expectedErrMsg)

				assert.Equal(test.VID, dev.Vendor)
				assert.Equal(test.PID, dev.Product)

				require.NoError(t, dev.Release())
			} else {
				assert.Contains(err.Error(), test.expectedErrMsg)
			}
		})
	}
}

// TestGetKeyValueFromEvent validates that the GetKeyValueFromEvent function
// parses the data from an evdev input event and returns a key value according
// to what is required for this service to operate
func TestGetKeyValueFromEvent(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	// construct the test cases
	tests := []struct {
		Name          string
		InputEvent    *evdev.InputEvent
		ExpectedValue int
		ExpectedError error
	}{
		{
			Name: "key down 1",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_1,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: 1,
			ExpectedError: nil,
		},
		{
			Name: "key down 2",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_2,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: 2,
			ExpectedError: nil,
		},
		{
			Name: "key down 3",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_3,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: 3,
			ExpectedError: nil,
		},
		{
			Name: "key down 4",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_4,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: 4,
			ExpectedError: nil,
		},
		{
			Name: "key down 5",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_5,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: 5,
			ExpectedError: nil,
		},
		{
			Name: "key down 6",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_6,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: 6,
			ExpectedError: nil,
		},
		{
			Name: "key down 7",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_7,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: 7,
			ExpectedError: nil,
		},
		{
			Name: "key down 8",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_8,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: 8,
			ExpectedError: nil,
		},
		{
			Name: "key down 9",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_9,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: 9,
			ExpectedError: nil,
		},
		{
			Name: "key down 0",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_0,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: 0,
			ExpectedError: nil,
		},
		{
			Name: "key down ENTER",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_ENTER,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedValue: int(evdev.KEY_ENTER),
			ExpectedError: nil,
		},
		{
			Name: "error key up 1",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_1,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyUp),
			},
			ExpectedError: fmt.Errorf("event created by the physical card reader was not a key press event"),
		},
		{
			Name: "error key down on an evdev event type that is not intended to be handled (EV_LED)",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_A,
				Type:  evdev.EV_LED,
				Value: int32(evdev.KeyDown),
			},
			ExpectedError: fmt.Errorf("event created by the physical card reader was not a key press event"),
		},
		{
			Name: "error key down on a key that is not intended to be handled (A)",
			InputEvent: &evdev.InputEvent{
				Code:  evdev.KEY_A,
				Type:  evdev.EV_KEY,
				Value: int32(evdev.KeyDown),
			},
			ExpectedError: fmt.Errorf("received an undesired or unexpected key code: %v", evdev.KEY_A),
		},
	}

	// iterate over test cases and run each test
	for _, test := range tests {
		// run individual tests
		t.Run(test.Name, func(t *testing.T) {
			// run the function in question
			actualExpectedValue, actualExpectedError := GetKeyValueFromEvent(test.InputEvent)
			// perform assertions
			require.Equal(test.ExpectedError, actualExpectedError)
			assert.Equal(test.ExpectedValue, actualExpectedValue)
		})
	}
}

// TestStatus is very similar to the GrabCardReader functionality, except
// the logic is reversed. A successful grab of the input device means that
// the device is not grabbed by our service, so an error is returned
func TestStatus(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	lc := logger.NewClient(physicalCardReaderDeviceServiceName, false, physicalLogFile, "DEBUG")

	// build tests
	tests := []struct {
		Name           string
		reader         *CardReaderPhysical
		expectedErrMsg string
	}{
		{
			"Status successfully grab physical card reader",
			&CardReaderPhysical{
				DeviceSearchPath: physicalDeviceSearchPath,
				VID:              physicalVID,
				PID:              physicalPID,
				LoggingClient:    lc,
			},
			"failure: physical card reader is not locked",
		},
		{
			"Status unsuccessfully grab card reader",
			&CardReaderPhysical{
				DeviceSearchPath: physicalInvalidDeviceSearchPath,
				VID:              physicalVID,
				PID:              physicalPID,
				LoggingClient:    lc,
			},
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := test.reader.Status()
			if err == nil {
				assert.Empty(test.expectedErrMsg)
			} else {
				assert.Contains(err.Error(), test.expectedErrMsg)
			}
		})
	}
}

// TestProcessDevReadEvents tests that the processDevReadEvents function
// properly handles input events
func TestProcessDevReadEvents(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	lc := logger.NewClient(physicalCardReaderDeviceServiceName, false, physicalLogFile, "DEBUG")

	tests := []struct {
		Name                         string
		InputEvents                  []evdev.InputEvent
		InputCardReaderPhysical      *CardReaderPhysical
		ExpectedCardReaderCardNumber string
		ExpectedLogLines             []string
	}{
		{
			Name: "successful test",
			InputEvents: []evdev.InputEvent{
				{Code: evdev.KEY_0, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_A, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_1, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_2, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_3, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_4, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_5, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_6, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_7, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_8, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_9, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
				{Code: evdev.KEY_ENTER, Type: evdev.EV_KEY, Value: int32(evdev.KeyDown)},
			},
			InputCardReaderPhysical: &CardReaderPhysical{
				DeviceName:    physicalDeviceName,
				LoggingClient: lc,
				Mocked:        true,
				CardNumber:    "",
			},
			ExpectedCardReaderCardNumber: "", // card number gets cleared
			ExpectedLogLines: []string{
				fmt.Sprintf("received event with card number %v", expectedCardNumberPhysical),
			},
		},
	}

	// run the tests
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// clear the log file
			err := clearLogs()
			require.NoError(err)

			// run the function in question
			test.InputCardReaderPhysical.processDevReadEvents(test.InputEvents)

			// perform assertions
			assert.Equal(test.ExpectedCardReaderCardNumber, test.InputCardReaderPhysical.CardNumber)

			// check if the expected log output is in the log file
			for _, str := range test.ExpectedLogLines {
				result, err := doesLogFileContainString(str)
				require.NoError(err)
				assert.True(result)
			}
		})
	}

	// clear the logs as a last step
	err := clearLogs()
	require.NoError(err)
}
