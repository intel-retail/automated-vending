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

const (
	deviceName = "controller-board"
)

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

func TestControllerBoardVirtual_Read(t *testing.T) {
	tests := []struct {
		name           string
		AsyncCh        chan *models.AsyncValues
		DevStatus      string
		LoggingClient  logger.LoggingClient
		L1             int
		L2             int
		DoorClosed     int
		Temperature    float64
		Humidity       int64
		DeviceName     string
		expectedStatus string
	}{
		{
			name:           "valid case",
			AsyncCh:        make(chan *models.AsyncValues, 16),
			DevStatus:      "",
			LoggingClient:  logger.NewMockClient(),
			L1:             0,
			L2:             0,
			DoorClosed:     0,
			Temperature:    0.00,
			Humidity:       0,
			DeviceName:     deviceName,
			expectedStatus: `{"lock1_status":0,"lock2_status":0,"door_closed":false,"temperature":0,"humidity":0}`, // Since temperature & humidity are not static, have to look just for the labels for them
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := &ControllerBoardVirtual{
				AsyncCh:       tt.AsyncCh,
				DevStatus:     tt.DevStatus,
				LoggingClient: tt.LoggingClient,
				L1:            tt.L1,
				L2:            tt.L2,
				DoorClosed:    tt.DoorClosed,
				Temperature:   tt.Temperature,
				Humidity:      tt.Humidity,
				DeviceName:    tt.DeviceName,
			}

			// Send a command so there is something to read
			_ = target.Write(Command.GetStatus)
			go target.Read()
			actual := <-tt.AsyncCh
			assert.NotNil(t, actual)
			actualStatus, err := actual.CommandValues[0].StringValue()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, actualStatus)
		})
	}

}
