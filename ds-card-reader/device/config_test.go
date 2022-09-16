// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestUpdateFromRaw(t *testing.T) {

	expectedConfig := &ServiceConfig{
		DriverConfig: Config{
			DeviceName:       "CardReader001",
			PID:              8037,
			VID:              2341,
			DeviceSearchPath: "/dev/input/event*",
			SimulateDevice:   true,
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
			rawConfig: expectedConfig.DriverConfig,
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
