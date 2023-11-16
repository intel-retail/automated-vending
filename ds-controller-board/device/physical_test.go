//go:build all || physical
// +build all physical

// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"fmt"
	"os"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultConfig = Config{
	VirtualControllerBoard: false,
	PID:                    validPID,
	VID:                    validVID,
	DisplayTimeout:         10000,
	LockTimeout:            30000,
}

var lc logger.LoggingClient
var target ControllerBoard
var asyncChan chan *models.AsyncValues

func TestMain(m *testing.M) {
	var err error
	asyncChan = make(chan *models.AsyncValues, 16) // need at least buffer so writes to channel don't block

	lc = logger.NewClient("TestInitialize", false, "./unit-test.log", "DEBUG")
	target, err = NewControllerBoard(lc, asyncChan, &defaultConfig)
	if err != nil {
		fmt.Println("Failed to create ControllerBoard: " + err.Error())
		os.Exit(-1)
	}

	os.Exit(m.Run())
}

func TestFindControllerBoard(t *testing.T) {
	testCases := []struct {
		Name          string
		VID           string
		PID           string
		ExpectedError error
	}{
		{"Success", validVID, validPID, nil},
		{"Wrong VID", badVID, validPID, fmt.Errorf("no USB port found matching VID=%s & PID=%s", badVID, validPID)},
		{"Wrong PID", validVID, badPID, fmt.Errorf("no USB port found matching VID=%s & PID=%s", validVID, badPID)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actual, err := FindControllerBoard(testCase.VID, testCase.PID)
			if testCase.ExpectedError != nil {
				require.Equal(t, testCase.ExpectedError, err)
				return // test is complete
			} else {
				require.NoError(t, err)
			}

			// The actual port name can vary so can only validate that we have a non-empty port name
			require.True(t, len(actual) > 0, "Port returned is empty")
		})
	}
}

func TestGetStatus(t *testing.T) {
	expected := ""
	actual := target.GetStatus()
	assert.Equal(t, expected, actual)
}

func TestPhysicalRead(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	expectedDeviceName := deviceName
	// Since temperature & humidity are not static, have to look just for the labels for them
	expectedStatusContains := []string{`"lock1_status":1,`, `"lock2_status":1,`, `"door_closed":false,`, `"temperature":`, `"humidity":`}

	// Send a command so there is something to read
	_ = target.Write(Command.GetStatus)
	go target.Read()
	actual := <-asyncChan
	assert.NotNil(actual)
	assert.Equal(expectedDeviceName, actual.DeviceName)
	actualStatus, err := actual.CommandValues[0].StringValue()
	require.NoError(err)
	for _, expectedStatus := range expectedStatusContains {
		assert.Contains(actualStatus, expectedStatus)
	}
}

func TestPhysicalWrite(t *testing.T) {
	notExpected := new(ControllerBoardPhysical)
	require.NotEqual(t, notExpected, target)

	err := target.Write(Command.Lock1)
	require.NoError(t, err)
}
