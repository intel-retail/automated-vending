//go:build all || !physical
// +build all !physical

// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package driver

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"ds-controller-board/device"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	edgexcommon "github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/stretchr/testify/assert"
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

func TestInitialize(t *testing.T) {
	target := CreateControllerBoardDriver(t, true, false, "")
	err := target.Initialize(lc, make(chan *models.AsyncValues), make(chan<- []models.DiscoveredDevice))
	assert.NoError(t, err)
}

func TestHandleReadCommands(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	expectedValue := "STATUS,L1,0,L2,0,D,0,T,78.58,H,19.54"
	expectedType := edgexcommon.ValueTypeString

	target := CreateControllerBoardDriver(t, true, true, expectedValue)

	request := models.CommandRequest{
		DeviceResourceName: "L1",
		Attributes:         nil,
		Type:               "",
	}

	actual, err := target.HandleReadCommands("ControllerBoard", nil, []models.CommandRequest{request})
	require.NoError(err)
	require.NotNil(actual)
	require.True(len(actual) > 0, "No results returned")
	assert.Equal(expectedType, actual[0].Type)

	actualValue, err := actual[0].StringValue()
	require.NoError(err)
	assert.Equal(expectedValue, actualValue)
}

func TestHandleWriteCommands(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	testCases := []struct {
		Name          string
		Resource      string
		CommandValue  interface{}
		ExpectedError error
	}{
		{Name: "HandleWriteCommands - lock1 with 1", Resource: lock1, CommandValue: true, ExpectedError: nil},
		{Name: "HandleWriteCommands - lock2 with 1", Resource: lock2, CommandValue: true, ExpectedError: nil},
		{Name: "HandleWriteCommands - lock1 with 0", Resource: lock1, CommandValue: false, ExpectedError: nil},
		{Name: "HandleWriteCommands - lock2 with 0", Resource: lock2, CommandValue: false, ExpectedError: nil},
		{Name: "HandleWriteCommands - getStatus", Resource: getStatus, CommandValue: nil, ExpectedError: nil},
		{Name: "HandleWriteCommands - displayRow0", Resource: displayRow0, CommandValue: "Row 0", ExpectedError: nil},
		{Name: "HandleWriteCommands - displayRow1", Resource: displayRow1, CommandValue: "Row 1", ExpectedError: nil},
		{Name: "HandleWriteCommands - displayRow2", Resource: displayRow2, CommandValue: "Row 2", ExpectedError: nil},
		{Name: "HandleWriteCommands - displayRow3", Resource: displayRow3, CommandValue: "Row 3", ExpectedError: nil},
		{Name: "HandleWriteCommands - displayReset", Resource: displayReset, CommandValue: nil, ExpectedError: nil},
		{Name: "HandleWriteCommands - setHumidity", Resource: setHumidity, CommandValue: "86", ExpectedError: nil},
		{Name: "HandleWriteCommands - setTemperature", Resource: setTemperature, CommandValue: "102", ExpectedError: nil},
		{Name: "HandleWriteCommands - setDoorClosed", Resource: setDoorClosed, CommandValue: "Yes", ExpectedError: nil},
		{Name: "HandleWriteCommands - Unknown Command", Resource: "unknown", ExpectedError: fmt.Errorf("unknown command received: 'unknown'")},
	}

	target := CreateControllerBoardDriver(t, true, true, "")

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var params []*models.CommandValue
			var commandValue *models.CommandValue
			var err error

			if testCase.CommandValue != nil {
				if reflect.TypeOf(testCase.CommandValue).Kind() == reflect.String {
					commandValue, err = models.NewCommandValueWithOrigin(
						testCase.Resource,
						edgexcommon.ValueTypeString,
						testCase.CommandValue.(string),
						0,
					)
					require.NoError(err)

				} else {
					value, ok := testCase.CommandValue.(bool)
					require.True(ok)
					commandValue, err = models.NewCommandValueWithOrigin(
						testCase.Resource,
						edgexcommon.ValueTypeBool,
						value,
						0,
					)
					require.NoError(err)
				}
			} else {
				commandValue = &models.CommandValue{
					DeviceResourceName: testCase.Resource,
				}
			}

			params = append(params, commandValue)

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
		config: &device.Config{
			VirtualControllerBoard: virtual,
		},
	}

	if initialize {
		err = target.Initialize(lc, make(chan *models.AsyncValues), make(chan<- []models.DiscoveredDevice))
		require.NoError(err)

		virtual, ok := target.controllerBoard.(*device.ControllerBoardVirtual)
		require.True(ok)
		virtual.DevStatus = expectedStatus
	}

	return target
}
