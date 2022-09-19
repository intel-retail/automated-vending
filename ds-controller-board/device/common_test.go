//go:build all
// +build all

// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStatus(t *testing.T) {
	expected := StatusEvent{
		Lock1Status: 1,
		Lock2Status: 1,
		DoorClosed:  true,
		Temperature: 78.58,
		Humidity:    19.54,
	}

	testCases := []struct {
		Name          string
		Status        string
		Expected      *StatusEvent
		ErrorExpected bool
		ErrorContains string
	}{
		{"Success", "STATUS,L1,1,L2,1,D,1,T,78.58,H,19.54", &expected, false, ""},
		{"Lock1 error", "STATUS,L1,a,L2,1,D,1,T,78.58,H,19.54", nil, true, "unable to parse Lock1 status"},
		{"Lock2 error", "STATUS,L1,0,L2,b,D,1,T,78.58,H,19.54", nil, true, "unable to parse Lock2 status"},
		{"DoorClosed error", "STATUS,L1,0,L2,0,D,c,T,78.58,H,19.54", nil, true, "unable to parse DoorClosed status"},
		{"Temperature error", "STATUS,L1,0,L2,0,D,0,T,temp,H,19.54", nil, true, "unable to parse Temperature status"},
		{"Humidity error", "STATUS,L1,0,L2,0,D,0,T,0.0,H,hum", nil, true, "unable to parse Humidity status"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actual, err := ParseStatus(testCase.Status)
			if testCase.ErrorExpected {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ErrorContains)
				return // test complete
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.Expected, &actual)
		})
	}
}
