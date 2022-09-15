// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFromRaw(t *testing.T) {
	expectedDisplayTimeout, err := time.ParseDuration("10s")
	require.NoError(t, err)
	expectedLockTimeout, err := time.ParseDuration("30s")
	require.NoError(t, err)
	expectedConfig := &ServiceConfig{
		AppCustom: CustomConfig{
			DriverConfig: Config{
				VirtualControllerBoard: true,
				PID:                    "8037",
				VID:                    "2341",
				DisplayTimeout:         expectedDisplayTimeout,
				LockTimeout:            expectedLockTimeout,
			},
		},
	}
	testCases := []struct {
		Name      string
		rawConfig interface{}
		isValid   bool
	}{
		{
			Name:      "valid",
			isValid:   true,
			rawConfig: expectedConfig,
		},
		{
			Name:      "not valid",
			isValid:   false,
			rawConfig: expectedConfig.AppCustom,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actualConfig := ServiceConfig{}

			ok := actualConfig.UpdateFromRaw(testCase.rawConfig)

			assert.Equal(t, testCase.isValid, ok)
			if testCase.isValid {
				assert.Equal(t, expectedConfig, &actualConfig)
			}
		})
	}

}
