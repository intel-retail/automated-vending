//go:build all || physical || !physical
// +build all physical !physical

// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

// notes on why physical and !physical build tags are present:

// this file requires !physical tag so it can be run when *no* tags are passed
// into the test tool (i.e. running simulated mode tests)

// this file also requires the physical tag so that it can be run when the
// "physical" tag is passed into the test tool

package driver

import (
	"ds-card-reader/common"
	"ds-card-reader/device"
	"fmt"
	"testing"

	sdkMocks "github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces/mocks"
	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	edgexcommon "github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	require "github.com/stretchr/testify/require"
)

const (
	invalid                     = "invalid"
	cardReaderDeviceServiceName = "ds-card-reader"
	expectedCardNumber          = "0003292356"
)

// getDefaultDriverConfig returns a DriverConfig that contains the
// same values as the current default values in configuration.toml
//
// WARNING: If changing the default values in configuration.toml, please
// update this function
func getDefaultDriverConfig() *device.ServiceConfig {
	return &device.ServiceConfig{
		DriverConfig: device.Config{
			DeviceName:       cardReaderDeviceServiceName,
			DeviceSearchPath: "/dev/input/event*",
			VID:              0xffff,
			PID:              0x0035,
			SimulateDevice:   true,
		},
	}
}

// TestStop validates that the driver Stop function is implemented without
// throwing any errors
func TestStop(t *testing.T) {
	driver := CardReaderDriver{
		Config: &device.ServiceConfig{
			DriverConfig: device.Config{
				DeviceName: cardReaderDeviceServiceName,
			},
		},
	}

	err := driver.Stop(false)
	assert.NoError(t, err)
}

// TestDisconnectDevice validates that the driver DisconnectDevice function is
// implemented without throwing any errors
func TestDisconnectDevice(t *testing.T) {
	driver := CardReaderDriver{
		Config: &device.ServiceConfig{
			DriverConfig: device.Config{
				DeviceName: cardReaderDeviceServiceName,
			},
		},
	}

	err := driver.DisconnectDevice(
		driver.Config.DriverConfig.DeviceName,
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

	lc := logger.NewMockClient()

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
				Config: &device.ServiceConfig{
					DriverConfig: device.Config{
						DeviceName: cardReaderDeviceServiceName,
					},
				},
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
				Config: &device.ServiceConfig{
					DriverConfig: device.Config{
						DeviceName: cardReaderDeviceServiceName,
					},
				},
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
				Config: &device.ServiceConfig{
					DriverConfig: device.Config{
						DeviceName: cardReaderDeviceServiceName,
					},
				},
			},
		},
	}

	// run the tests to handle read commands
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			result, err := test.driver.HandleReadCommands(
				test.driver.Config.DriverConfig.DeviceName,
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

		})
	}
}

// TestHandleWriteCommands validates that the HandleWriteCommands behaves
// as expected
func TestHandleWriteCommands(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	require := require.New(t)

	// prepare some variables for usage in the below tests
	protocolProperties := map[string]models.ProtocolProperties{}
	lc := logger.NewMockClient()

	successfulCommandVal, err := dsModels.NewCommandValueWithOrigin(common.CommandCardNumber, edgexcommon.ValueTypeString, expectedCardNumber, 0)
	require.NoError(err)

	invalidCommandVal, err := dsModels.NewCommandValueWithOrigin(invalid, edgexcommon.ValueTypeString, expectedCardNumber, 0)
	require.NoError(err)

	nonStringCommandVal, err := dsModels.NewCommandValueWithOrigin(common.CommandCardNumber, edgexcommon.ValueTypeFloat64, 0.01, 0)
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
				Config: &device.ServiceConfig{
					DriverConfig: device.Config{
						DeviceName: cardReaderDeviceServiceName,
					},
				},
			},
		},
		{
			"HandleWriteCommands input param with non-string type",
			[]*dsModels.CommandValue{nonStringCommandVal},
			[]dsModels.CommandRequest{{}},
			[]string{fmt.Sprintf("write command \\\"%v\\\" received non-string value: %v", common.CommandCardNumber, "cannot convert CommandValue of Float64 to String")},
			fmt.Errorf("write command \"%v\" received non-string value: %v", common.CommandCardNumber, "cannot convert CommandValue of Float64 to String"),
			&CardReaderDriver{
				LoggingClient: lc,
				CardReader: &device.CardReaderVirtual{
					AsyncCh:       make(chan *dsModels.AsyncValues, 16),
					LoggingClient: lc,
				},
				Config: &device.ServiceConfig{
					DriverConfig: device.Config{
						DeviceName: cardReaderDeviceServiceName,
					},
				},
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
				Config: &device.ServiceConfig{
					DriverConfig: device.Config{
						DeviceName: cardReaderDeviceServiceName,
					},
				},
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
				Config: &device.ServiceConfig{
					DriverConfig: device.Config{
						DeviceName: cardReaderDeviceServiceName,
					},
				},
			},
		},
	}

	// run the tests to handle read commands
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// run the handle write commands function
			err = test.driver.HandleWriteCommands(
				test.driver.Config.DriverConfig.DeviceName,
				protocolProperties,
				test.inputReqs,
				test.inputParams,
			)

			// perform assertions
			require.Equal(test.expectedError, err)
		})
	}
}

func TestCardReaderDriver_Initialize(t *testing.T) {
	type args struct {
		lc       logger.LoggingClient
		asyncCh  chan<- *dsModels.AsyncValues
		deviceCh chan<- []dsModels.DiscoveredDevice
	}
	tests := []struct {
		name    string
		drv     *CardReaderDriver
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			drv: &CardReaderDriver{
				LoggingClient: logger.NewMockClient(),
				CardReader:    nil,
				Config:        getDefaultDriverConfig(),
			},
			args: args{
				lc:       logger.NewMockClient(),
				asyncCh:  make(chan *dsModels.AsyncValues, 16),
				deviceCh: make(chan []dsModels.DiscoveredDevice, 16)},
			wantErr: false,
		},
		{
			name: "nil configuration",
			drv: &CardReaderDriver{
				LoggingClient: logger.NewMockClient(),
				CardReader:    nil,
				Config:        nil,
			},
			args: args{
				lc:       logger.NewMockClient(),
				asyncCh:  make(chan *dsModels.AsyncValues, 16),
				deviceCh: make(chan []dsModels.DiscoveredDevice, 16)},
			wantErr: true,
		},
		{
			name: "invalid configuration info",
			drv: &CardReaderDriver{
				LoggingClient: logger.NewMockClient(),
				CardReader:    nil,
				Config: &device.ServiceConfig{
					DriverConfig: device.Config{
						DeviceName:       "invalid",
						DeviceSearchPath: "/dev/input/event*",
						VID:              0,
						PID:              0,
						SimulateDevice:   false,
					},
				},
			},
			args: args{
				lc:       logger.NewMockClient(),
				asyncCh:  make(chan *dsModels.AsyncValues, 16),
				deviceCh: make(chan []dsModels.DiscoveredDevice, 16)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSDK := &sdkMocks.DeviceServiceSDK{}
			mockSDK.On("LoggingClient").Return(tt.args.lc)
			mockSDK.On("AsyncValuesChannel").Return(nil)
			mockSDK.On("LoadCustomConfig", mock.Anything, mock.Anything).Return(nil)
			if err := tt.drv.Initialize(mockSDK); (err != nil) != tt.wantErr {
				t.Errorf("CardReaderDriver.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
