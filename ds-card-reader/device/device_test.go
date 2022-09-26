// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"reflect"
	"testing"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	require "github.com/stretchr/testify/require"
)

const (
	physicalDeviceSearchPath = "/dev/input/event*"
	physicalDeviceName       = "ds-card-reader"
	physicalVID              = uint16(0xFFFF)
	physicalPID              = uint16(0x0035)
)

func TestInitializeCardReader(t *testing.T) {
	require := require.New(t)
	var emptyCardReader CardReader
	tests := []struct {
		name             string
		lc               logger.LoggingClient
		asyncCh          chan<- *dsModels.AsyncValues
		deviceSearchPath string
		deviceName       string
		vid              uint16
		pid              uint16
		simulateDevice   bool
		mockDevice       bool
		wantErr          bool
	}{
		{
			name:             "valid case simulatorMode",
			lc:               logger.NewMockClient(),
			asyncCh:          make(chan<- *dsModels.AsyncValues, 16),
			deviceSearchPath: physicalDeviceSearchPath,
			deviceName:       physicalDeviceName,
			vid:              physicalVID,
			pid:              physicalPID,
			simulateDevice:   true,
			mockDevice:       true,
			wantErr:          false,
		},
		{
			name:             "not simulatorMode",
			lc:               logger.NewMockClient(),
			asyncCh:          make(chan<- *dsModels.AsyncValues, 16),
			deviceSearchPath: physicalDeviceSearchPath,
			deviceName:       physicalDeviceName,
			vid:              physicalVID,
			pid:              physicalPID,
			simulateDevice:   false,
			mockDevice:       true,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCardReader, err := InitializeCardReader(tt.lc, tt.asyncCh, tt.deviceSearchPath, tt.deviceName, tt.vid, tt.pid, tt.simulateDevice, tt.mockDevice)
			require.Equal((err != nil), tt.wantErr)

			IsAEmptyCardReader := reflect.DeepEqual(gotCardReader, emptyCardReader)
			require.Equal(IsAEmptyCardReader, tt.wantErr)
			if !IsAEmptyCardReader {
				err = gotCardReader.Release()
				require.NoError(err)
			}
		})
	}
}
