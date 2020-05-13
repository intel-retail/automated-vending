// +build all !physical

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

	"ds-card-reader/common"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

const (
	virtualLogFile                     = "virtual_test.log"
	virtualCardReaderDeviceServiceName = "ds-card-reader"
	virtualDeviceSearchPath            = ""
	virtualDeviceName                  = "ds-card-reader"
	virtualVID                         = uint16(0x0000)
	virtualPID                         = uint16(0x0000)
	expectedCardNumberVirtual          = "0123456789"
)

func clearVirtualLogs() error {
	return ioutil.WriteFile(virtualLogFile, []byte{}, 0644)
}

func doesVirtualLogFileContainString(input string) (bool, error) {
	// attempt to open the log file
	file, err := os.Open(virtualLogFile)
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

// TestVirtualCardReader validates that the InitializeCardReader
// and the virtual functions work as expected for a virtual card reader
func TestVirtualCardReader(t *testing.T) {

	// prepare a few needed interfaces for use in tests
	lc := logger.NewClient(virtualCardReaderDeviceServiceName, false, virtualLogFile, "DEBUG")
	expectedAsyncCh := make(chan<- *dsModels.AsyncValues, 16)

	// we will expect the result of our tests to have a CardReader interface
	// that does not equal an uninitialized CardReader interface, hence
	// why it is "unexpected" - it is _not_ the expected value in our test
	var unexpectedCardReader CardReader

	// run the function
	cardReader, err := InitializeCardReader(
		lc,
		expectedAsyncCh,
		virtualDeviceSearchPath,
		virtualDeviceName,
		virtualVID,
		virtualPID,
		true,
		true,
	)

	// perform assertions
	require.NoError(t, err)
	assert.NotEqual(t, unexpectedCardReader, cardReader)
}

// TestVirtualListen validates that the Listen() function properly
// does nothing. The function has to be implemented in order to follow
// the virtual/physical abstraction interface
func TestVirtualListen(t *testing.T) {
	reader := &CardReaderVirtual{}

	reader.Listen()
}

// TestVirtualRelease validates that the Release() function properly
// returns a nil error. The function has to be implemented in order to follow
// the virtual/physical abstraction interface
func TestVirtualRelease(t *testing.T) {
	reader := &CardReaderVirtual{}

	err := reader.Release()
	assert.NoError(t, err)
}

// TestVirtualStatus validates that the Status() function properly
// returns a nil error. The function has to be implemented in order to follow
// the virtual/physical abstraction interface
func TestVirtualStatus(t *testing.T) {
	tests := []struct {
		reader        *CardReaderVirtual
		expectedError error
	}{
		{
			&CardReaderVirtual{},
			nil,
		},
		{
			&CardReaderVirtual{MockFailStatusCheck: true},
			fmt.Errorf("status check failed"),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.reader.Status(), test.expectedError)
	}
}

// TestWrite validates that the virtual device's Write function pushes
// a valid command value to the async channel of the device
func TestWrite(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	err := clearVirtualLogs()
	require.NoError(err)

	// note: it is critical to create a two-way channel and then pass in the
	// one-way component of this channel to the CardReaderVirtual below.
	// In order to receive values from the channel, the original channel
	// must be used, not the channel that was passed to the CardReaderVirtual,
	// since it only takes the one-way component
	asyncCh := make(chan *dsModels.AsyncValues, 16)

	loggingClient := logger.NewClient(virtualCardReaderDeviceServiceName, false, virtualLogFile, "DEBUG")

	reader := CardReaderVirtual{
		AsyncCh:       asyncCh,
		DeviceName:    virtualDeviceName,
		LoggingClient: loggingClient,
	}

	reader.Write(common.CommandCardReaderEvent, expectedCardNumberVirtual)

	actual := <-asyncCh
	require.NotNil(actual.DeviceName) // "actual" is a pointer, must not be nil
	assert.Equal(virtualDeviceName, actual.DeviceName)

	actualStringValue, err := actual.CommandValues[0].StringValue()
	require.NoError(err)

	assert.Equal(expectedCardNumberVirtual, actualStringValue)

	actualLogPresence, err := doesVirtualLogFileContainString(fmt.Sprintf("received event with card number %v", expectedCardNumberVirtual))
	require.NoError(err)
	assert.True(actualLogPresence)

	err = clearVirtualLogs()
	require.NoError(err)
}
