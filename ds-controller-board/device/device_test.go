//go:build all || physical
// +build all physical

// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"fmt"
	"testing"

	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewControllerBoardError(t *testing.T) {
	config := &Config{
		VirtualControllerBoard: false,
	}

	testCases := []struct {
		Name                 string
		VID                  string
		PID                  string
		ExpectedErrorMessage string
	}{
		// Not testing success since it is handled in physical_test.go and will fail here due to port already in use.
		{"Wrong VID", badVID, validPID, fmt.Sprintf("no USB port found matching VID=%s & PID=%s", badVID, validPID)},
		{"Wrong PID", validVID, badPID, fmt.Sprintf("no USB port found matching VID=%s & PID=%s", validVID, badPID)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			config.VID = testCase.VID
			config.PID = testCase.PID
			_, err := NewControllerBoard(logger.NewMockClient(), make(chan *dsModels.AsyncValues), config)
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
		})
	}
}
