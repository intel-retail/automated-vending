//go:build all || !physical
// +build all !physical

// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"testing"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVirtualRead(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	expectedDeviceName := deviceName
	// Since temperature & humidity are not static, have to look just for the labels for them
	expectedStatus := `{"lock1_status":0,"lock2_status":0,"door_closed":false,"temperature":0,"humidity":0}`

	asyncChan := make(chan *models.AsyncValues, 16)

	target := ControllerBoardVirtual{
		LoggingClient: logger.NewMockClient(),
		AsyncCh:       asyncChan,
	}

	// Send a command so there is something to read
	_ = target.Write(Command.GetStatus)
	go target.Read()
	actual := <-asyncChan
	assert.NotNil(actual)
	assert.Equal(expectedDeviceName, actual.DeviceName)
	actualStatus, err := actual.CommandValues[0].StringValue()
	require.NoError(err)
	assert.Equal(expectedStatus, actualStatus)
}

func TestVirtualWrite(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)
	require := require.New(t)

	expected := "STATUS,L1,1,L2,1,D,0,T,0.00,H,0"

	target := ControllerBoardVirtual{
		LoggingClient: logger.NewMockClient(),
	}

	err := target.Write(Command.Lock1)
	require.NoError(err)

	err = target.Write(Command.Lock2)
	require.NoError(err)

	actual := target.getRawStatus()
	assert.Equal(expected, actual)

	expected = "STATUS,L1,0,L2,1,D,0,T,0.00,H,0"
	err = target.Write(Command.UnLock1)
	require.NoError(err)

	actual = target.getRawStatus()
	assert.Equal(expected, actual)

	expected = "STATUS,L1,0,L2,0,D,0,T,0.00,H,0"
	err = target.Write(Command.UnLock2)
	require.NoError(err)

	actual = target.getRawStatus()
	assert.Equal(expected, actual)
}
