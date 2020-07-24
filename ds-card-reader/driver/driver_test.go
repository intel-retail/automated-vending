// +build all physical !physical

// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

// notes on why physical and !physical build tags are present:

// this file requires !physical tag so it can be run when *no* tags are passed
// into the test tool (i.e. running simulated mode tests)

// this file also requires the physical tag so that it can be run when the
// "physical" tag is passed into the test tool

package driver

import (
	"bufio"
	"ds-card-reader/common"
	"ds-card-reader/device"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

var once sync.Once

const (
	invalid                     = "invalid"
	logFile                     = "driver_test.log"
	cardReaderDeviceServiceName = "ds-card-reader"
	expectedCardNumber          = "0003292356"
)

func clearLogs() error {
	return ioutil.WriteFile(logFile, []byte{}, 0644)
}

func doesLogFileContainString(input string) (bool, error) {
	// attempt to open the log file
	file, err := os.Open(logFile)
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

// getDefaultCardReaderConfig returns a CardReaderConfig that contains the
// same values as the current default values in configuration.toml
//
// WARNING: If changing the default values in configuration.toml, please
// update this function
func getDefaultCardReaderConfig() *common.CardReaderConfig {
	return &common.CardReaderConfig{
		DeviceName:       cardReaderDeviceServiceName,
		DeviceSearchPath: "/dev/input/event*",
		VID:              0xffff,
		PID:              0x0035,
		SimulateDevice:   true,
	}
}

// TestInitialize validates that the device service interacts with the driver
// as expected. Due to the way that the EdgeX device service relies on a
// the singleton function "sdk.NewService()", we have to put most tests in
// a specific order and parallelizing them might cause the SDK to yield
// bad results
func TestInitialize(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	driver := CardReaderDriver{}
	lc := logger.NewClient(cardReaderDeviceServiceName, false, logFile, "DEBUG")

	// create an empty logging client/card reader device to compare against
	var emptyLogger logger.LoggingClient
	var emptyCardReaderDevice device.CardReader

	driver.Config = getDefaultCardReaderConfig()

	err := driver.Initialize(
		lc,
		make(chan *dsModels.AsyncValues, 16),
		make(chan []dsModels.DiscoveredDevice, 16),
	)

	require.NoError(t, err)
	assert.NotEqual(emptyLogger, driver.LoggingClient)
	assert.Equal(getDefaultCardReaderConfig(), driver.Config)
	assert.NotEqual(emptyCardReaderDevice, driver.CardReader)
}

// TestStop validates that the driver Stop function is implemented without
// throwing any errors
func TestStop(t *testing.T) {
	driver := CardReaderDriver{
		Config: &common.CardReaderConfig{DeviceName: cardReaderDeviceServiceName},
	}

	err := driver.Stop(false)
	assert.NoError(t, err)
}

// TestDisconnectDevice validates that the driver DisconnectDevice function is
// implemented without throwing any errors
func TestDisconnectDevice(t *testing.T) {
	driver := CardReaderDriver{
		Config: &common.CardReaderConfig{DeviceName: cardReaderDeviceServiceName},
	}

	err := driver.DisconnectDevice(
		driver.Config.DeviceName,
		map[string]models.ProtocolProperties{},
	)

	assert.NoError(t, err)
}

// TestHandleReadCommands validates that the HandleReadCommands function
// properly handles incoming EdgeX read commands
func TestHandleReadCommands(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	lc := logger.NewClient(cardReaderDeviceServiceName, false, logFile, "DEBUG")

	tests := []struct {
		Name               string
		DeviceResourceName string
		ExpectedLogLines   []string
		ExpectedResult     []*dsModels.CommandValue
		ExpectedError      error
		driver             *CardReaderDriver
	}{
		{
			Name:               "HandleReadCommands successful status test",
			DeviceResourceName: common.CommandCardReaderStatus,
			ExpectedLogLines: []string{
				fmt.Sprintf("read command: %v, verifying lock on device", common.CommandCardReaderStatus),
				fmt.Sprintf("read command: %v, device ok", common.CommandCardReaderStatus),
			},
			ExpectedError: nil,
			driver: &CardReaderDriver{
				LoggingClient: lc,
				CardReader: &device.CardReaderVirtual{
					AsyncCh:       make(chan *dsModels.AsyncValues, 16),
					LoggingClient: lc,
				},
				Config: &common.CardReaderConfig{DeviceName: cardReaderDeviceServiceName},
			},
		},
		{
			Name:               "HandleReadCommands unsuccessful status test",
			DeviceResourceName: common.CommandCardReaderStatus,
			ExpectedLogLines: []string{
				fmt.Sprintf("read command: %v, verifying lock on device", common.CommandCardReaderStatus),
				fmt.Sprintf("read command: %v, failed to verify lock on device: status check failed", common.CommandCardReaderStatus),
			},
			ExpectedError: fmt.Errorf("read command: %v, failed to verify lock on device: status check failed", common.CommandCardReaderStatus),
			driver: &CardReaderDriver{
				LoggingClient: lc,
				CardReader: &device.CardReaderVirtual{
					LoggingClient:       lc,
					MockFailStatusCheck: true,
				},
				Config: &common.CardReaderConfig{DeviceName: cardReaderDeviceServiceName},
			},
		},
		{
			Name:               "HandleReadCommands unhandled device resource name",
			DeviceResourceName: invalid,
			ExpectedLogLines:   []string{},
			ExpectedError:      fmt.Errorf("read command \"%v\" is not handled by this device service", invalid),
			driver: &CardReaderDriver{
				LoggingClient: lc,
				CardReader: &device.CardReaderVirtual{
					LoggingClient: lc,
				},
				Config: &common.CardReaderConfig{DeviceName: cardReaderDeviceServiceName},
			},
		},
	}

	// run the tests to handle read commands
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// clear the log file
			err := clearLogs()
			require.NoError(err)

			result, err := test.driver.HandleReadCommands(
				test.driver.Config.DeviceName,
				map[string]models.ProtocolProperties{},
				[]dsModels.CommandRequest{
					{
						DeviceResourceName: test.DeviceResourceName,
					},
				},
			)

			// perform assertions
			require.Equal(test.ExpectedError, err)
			assert.Equal(test.ExpectedResult, result)

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

// TestHandleWriteCommands validates that the HandleWriteCommands behaves
// as expected
func TestHandleWriteCommands(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	// prepare some variables for usage in the below tests
	protocolProperties := map[string]models.ProtocolProperties{}
	lc := logger.NewClient(cardReaderDeviceServiceName, false, logFile, "DEBUG")

	successfulCommandVal, err := dsModels.NewCommandValue(common.CommandCardReaderEvent, 0, expectedCardNumber, dsModels.String)
	require.NoError(err)

	invalidCommandVal, err := dsModels.NewCommandValue(invalid, 0, expectedCardNumber, dsModels.String)
	require.NoError(err)

	nonStringCommandVal, err := dsModels.NewCommandValue(common.CommandCardReaderEvent, 0, 0.01, dsModels.ParseValueType("float64"))
	require.NoError(err)

	tests := []struct {
		name             string
		inputParams      []*dsModels.CommandValue
		inputReqs        []dsModels.CommandRequest
		expectedLogLines []string
		expectedError    error
		driver           *CardReaderDriver
	}{
		{
			"HandleWriteCommands empty input params (command values)",
			[]*dsModels.CommandValue{},
			[]dsModels.CommandRequest{{}},
			[]string{},
			fmt.Errorf("no params were passed into the write command handler for device %v", cardReaderDeviceServiceName),
			&CardReaderDriver{
				LoggingClient: lc,
				CardReader: &device.CardReaderVirtual{
					AsyncCh:       make(chan *dsModels.AsyncValues, 16),
					LoggingClient: lc,
				},
				Config: &common.CardReaderConfig{DeviceName: cardReaderDeviceServiceName},
			},
		},
		{
			"HandleWriteCommands input param with non-string type",
			[]*dsModels.CommandValue{nonStringCommandVal},
			[]dsModels.CommandRequest{{}},
			[]string{fmt.Sprintf("write command \\\"%v\\\" received non-string value: %v", common.CommandCardReaderEvent, "the data type is not string")},
			fmt.Errorf("write command \"%v\" received non-string value: %v", common.CommandCardReaderEvent, "the data type is not string"),
			&CardReaderDriver{
				LoggingClient: lc,
				CardReader: &device.CardReaderVirtual{
					AsyncCh:       make(chan *dsModels.AsyncValues, 16),
					LoggingClient: lc,
				},
				Config: &common.CardReaderConfig{DeviceName: cardReaderDeviceServiceName},
			},
		},
		{
			"HandleWriteCommands unhandled device resource name",
			[]*dsModels.CommandValue{invalidCommandVal},
			[]dsModels.CommandRequest{{DeviceResourceName: invalid}},
			[]string{fmt.Sprintf("write command \\\"%v\\\" is not handled by this device service", invalid)},
			fmt.Errorf("write command \"%v\" is not handled by this device service", invalid),
			&CardReaderDriver{
				LoggingClient: lc,
				CardReader: &device.CardReaderVirtual{
					AsyncCh:       make(chan *dsModels.AsyncValues, 16),
					LoggingClient: lc,
				},
				Config: &common.CardReaderConfig{DeviceName: cardReaderDeviceServiceName},
			},
		},
		{
			"HandleWriteCommands successful write test",
			[]*dsModels.CommandValue{successfulCommandVal},
			[]dsModels.CommandRequest{{DeviceResourceName: common.CommandCardReaderStatus}},
			[]string{},
			nil,
			&CardReaderDriver{
				LoggingClient: lc,
				CardReader: &device.CardReaderVirtual{
					AsyncCh:       make(chan *dsModels.AsyncValues, 16),
					LoggingClient: lc,
				},
				Config: &common.CardReaderConfig{DeviceName: cardReaderDeviceServiceName},
			},
		},
	}

	// run the tests to handle read commands
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// clear the log file
			err := clearLogs()
			require.NoError(err)

			// run the handle write commands function
			err = test.driver.HandleWriteCommands(
				test.driver.Config.DeviceName,
				protocolProperties,
				test.inputReqs,
				test.inputParams,
			)

			// perform assertions
			require.Equal(test.expectedError, err)

			// check if the expected log output is in the log file
			for _, str := range test.expectedLogLines {
				result, err := doesLogFileContainString(str)
				require.NoError(err)
				// this log output is confusing when it errors out, so
				// a message describing the assertion clearly has been added
				assert.True(result, fmt.Sprintf("test named \"%v\" expects log file to contain log string \"%v\"", test.name, str))
			}
		})
	}

	// clear the log file
	err = clearLogs()
	require.NoError(err)
}
