//go:build all || !physical
// +build all !physical

// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package driver

import (
	"ds-controller-board/device"
	"ds-controller-board/device/mocks"
	"fmt"
	"os"
	"reflect"
	"testing"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	edgexcommon "github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var lc logger.LoggingClient

func TestMain(m *testing.M) {
	lc = logger.NewMockClient()
	os.Exit(m.Run())
}

func TestNewControllerBoardDeviceDriver(t *testing.T) {
	expected := &ControllerBoardDriver{}
	actual := NewControllerBoardDeviceDriver()
	assert.Equal(t, expected, actual)
}

func TestDisconnectDevice(t *testing.T) {
	target := CreateControllerBoardDriver(t, true, false, "")
	actual := target.Stop(true)
	assert.Nil(t, actual)
}

func TestHandleWriteCommands(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)
	var cmdType interface{}

	mocksControllerBoard := &mocks.ControllerBoard{}
	mocksControllerBoard.On("Write", mock.Anything).Return(fmt.Errorf("failed"))
	cmdType = "notbool"

	var emptyControllerBoard device.ControllerBoard

	testCases := []struct {
		Name            string
		Resource        string
		CommandValue    interface{}
		ExpectedError   error
		controllerBoard device.ControllerBoard
		paramValue      interface{}
	}{
		{Name: "HandleWriteCommands - lock1 with 1", Resource: lock1, CommandValue: true, ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - lock2 with 1", Resource: lock2, CommandValue: true, ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - lock1 with 0", Resource: lock1, CommandValue: false, ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - lock2 with 0", Resource: lock2, CommandValue: false, ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - getStatus", Resource: getStatus, CommandValue: nil, ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - displayRow0", Resource: displayRow0, CommandValue: "Row 0", ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - displayRow1", Resource: displayRow1, CommandValue: "Row 1", ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - displayRow2", Resource: displayRow2, CommandValue: "Row 2", ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - displayRow3", Resource: displayRow3, CommandValue: "Row 3", ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - displayReset", Resource: displayReset, CommandValue: nil, ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - setHumidity", Resource: setHumidity, CommandValue: "86", ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - setTemperature", Resource: setTemperature, CommandValue: "102", ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - setDoorClosed", Resource: setDoorClosed, CommandValue: "Yes", ExpectedError: nil, controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands - Unknown Command", Resource: "unknown", ExpectedError: fmt.Errorf("unknown command received: 'unknown'"), controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands -error lock1", Resource: lock1, CommandValue: nil, ExpectedError: fmt.Errorf("unknown Command Type: '%v'", cmdType), paramValue: "notbool", controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands -error lock2", Resource: lock2, CommandValue: nil, ExpectedError: fmt.Errorf("unknown Command Type: '%v'", cmdType), paramValue: "notbool", controllerBoard: emptyControllerBoard},
		{Name: "HandleWriteCommands -error mock", Resource: lock1, CommandValue: nil, ExpectedError: fmt.Errorf("unknown Command Type: '%v'", cmdType), paramValue: "notbool", controllerBoard: mocksControllerBoard},
		{Name: "HandleWriteCommands -error mock", Resource: lock2, CommandValue: nil, ExpectedError: fmt.Errorf("unknown Command Type: '%v'", cmdType), paramValue: "notbool", controllerBoard: mocksControllerBoard},
		{Name: "HandleWriteCommands -error mock", Resource: getStatus, CommandValue: nil, ExpectedError: fmt.Errorf("failed"), paramValue: "notbool", controllerBoard: mocksControllerBoard},
	}

	target := CreateControllerBoardDriver(t, true, true, "")

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var params []*dsModels.CommandValue
			var commandValue *dsModels.CommandValue
			var err error

			if testCase.CommandValue != nil {
				if reflect.TypeOf(testCase.CommandValue).Kind() == reflect.String {
					commandValue, err = dsModels.NewCommandValueWithOrigin(
						testCase.Resource,
						edgexcommon.ValueTypeString,
						testCase.CommandValue.(string),
						0,
					)
					require.NoError(err)

				} else {
					value, ok := testCase.CommandValue.(bool)
					require.True(ok)
					commandValue, err = dsModels.NewCommandValueWithOrigin(
						testCase.Resource,
						edgexcommon.ValueTypeBool,
						value,
						0,
					)
					require.NoError(err)
				}
			} else {
				commandValue = &dsModels.CommandValue{
					DeviceResourceName: testCase.Resource,
					Value:              testCase.paramValue,
				}
			}

			params = append(params, commandValue)
			if !reflect.DeepEqual(testCase.controllerBoard, emptyControllerBoard) {
				target.controllerBoard = testCase.controllerBoard
			}
			actualError := target.HandleWriteCommands("ControllerBoard", nil, nil, params)

			if testCase.ExpectedError != nil {
				require.Error(actualError)
				assert.Equal(testCase.ExpectedError, actualError)
			} else {
				require.NoError(actualError)
			}
		})
	}
}

func CreateControllerBoardDriver(t *testing.T, virtual bool, initialize bool, expectedStatus string) *ControllerBoardDriver {
	var err error
	// use community-recommended shorthand (known name clash)
	require := require.New(t)

	target := &ControllerBoardDriver{
		lc: lc,
		config: &device.ServiceConfig{
			DriverConfig: device.Config{
				VirtualControllerBoard: virtual,
			},
		},
	}

	if initialize {
		err = target.Initialize(lc, make(chan *dsModels.AsyncValues), make(chan<- []dsModels.DiscoveredDevice))
		require.NoError(err)

		virtual, ok := target.controllerBoard.(*device.ControllerBoardVirtual)
		require.True(ok)
		virtual.DevStatus = expectedStatus
	}

	return target
}

func TestControllerBoardDriver_Initialize(t *testing.T) {

	mocklc := logger.NewMockClient()

	tests := []struct {
		name      string
		isVirtual bool
		config    *device.ServiceConfig
		lc        logger.LoggingClient
		wantErr   bool
	}{
		{
			name: "valid virtual case",
			config: &device.ServiceConfig{
				DriverConfig: device.Config{
					VirtualControllerBoard: true,
				},
			},
			lc:      mocklc,
			wantErr: false,
		},
		{
			name:    "Invalid nil config",
			config:  nil,
			lc:      mocklc,
			wantErr: true,
		},
		{
			name: "Invalid non-virtual case",
			config: &device.ServiceConfig{
				DriverConfig: device.Config{
					VirtualControllerBoard: false,
				},
			},
			lc:      mocklc,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drv := &ControllerBoardDriver{
				lc:     tt.lc,
				config: tt.config,
			}
			if err := drv.Initialize(tt.lc, make(chan *dsModels.AsyncValues), make(chan<- []dsModels.DiscoveredDevice)); (err != nil) != tt.wantErr {
				t.Errorf("ControllerBoardDriver.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControllerBoardDriver_HandleReadCommands(t *testing.T) {
	mocksControllerBoard := &mocks.ControllerBoard{}
	mocksControllerBoard.On("Write", mock.Anything).Return(fmt.Errorf("failed"))

	require := require.New(t)
	tests := []struct {
		name            string
		deviceName      string
		protocols       map[string]models.ProtocolProperties
		reqs            dsModels.CommandRequest
		lc              logger.LoggingClient
		config          *device.ServiceConfig
		controllerBoard device.ControllerBoard
		wantErr         bool
		expectedType    string
		expectedValue   string
	}{
		{
			name:       "valid case",
			deviceName: "test-device",
			protocols:  nil,
			controllerBoard: &device.ControllerBoardVirtual{
				AsyncCh:       make(chan *dsModels.AsyncValues),
				DevStatus:     "STATUS,L1,0,L2,0,D,0,T,78.58,H,19.54",
				LoggingClient: logger.NewMockClient(),
				L1:            1,
				L2:            1,
				DoorClosed:    1,
				Temperature:   78.00,
				Humidity:      10,
				DeviceName:    "test-device",
			},
			reqs: dsModels.CommandRequest{
				DeviceResourceName: "L1",
				Attributes:         nil,
				Type:               "",
			},
			wantErr: false,
			lc:      logger.NewMockClient(),
			config: &device.ServiceConfig{
				DriverConfig: device.Config{
					VirtualControllerBoard: true,
				},
			},
			expectedValue: "STATUS,L1,0,L2,0,D,0,T,78.58,H,19.54",
			expectedType:  edgexcommon.ValueTypeString,
		},
		{
			name:            "mock test",
			deviceName:      "test-device",
			protocols:       nil,
			controllerBoard: mocksControllerBoard,
			reqs: dsModels.CommandRequest{
				DeviceResourceName: "L1",
				Attributes:         nil,
				Type:               "",
			},
			wantErr: true,
			lc:      logger.NewMockClient(),
			config: &device.ServiceConfig{
				DriverConfig: device.Config{
					VirtualControllerBoard: true,
				},
			},
			expectedValue: "STATUS,L1,0,L2,0,D,0,T,78.58,H,19.54",
			expectedType:  edgexcommon.ValueTypeString,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drv := CreateControllerBoardDriver(t, true, true, tt.expectedValue)
			drv.controllerBoard = tt.controllerBoard
			actual, err := drv.HandleReadCommands(tt.deviceName, tt.protocols, []dsModels.CommandRequest{tt.reqs})
			if tt.wantErr {
				require.Error(err)
				require.Nil(actual)
			} else {
				require.NoError(err)
				require.NotNil(actual)
				require.True(len(actual) > 0, "No results returned")
				assert.Equal(t, tt.expectedType, actual[0].Type)

				actualValue, err := actual[0].StringValue()
				require.NoError(err)
				assert.Equal(t, tt.expectedValue, actualValue)
			}
		})
	}
}
