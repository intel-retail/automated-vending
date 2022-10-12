// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateFromRaw(t *testing.T) {

	expectedConfig := &ServiceConfig{
		DriverConfig: Config{
			VirtualControllerBoard: true,
			PID:                    "8037",
			VID:                    "2341",
			DisplayTimeout:         "30s",
			LockTimeout:            "10s",
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

func TestServiceConfig_Validate(t *testing.T) {
	type fields struct {
		DriverConfig Config
	}
	tests := []struct {
		name    string
		fields  fields
		want    time.Duration
		want1   time.Duration
		wantErr bool
	}{
		{
			name: "successful vaidation",
			fields: fields{
				DriverConfig: Config{
					DisplayTimeout: "30s",
					LockTimeout:    "10s",
				},
			},
			want:    30 * time.Second,
			want1:   10 * time.Second,
			wantErr: false,
		},
		{
			name: "unsuccessful vaidation",
			fields: fields{
				DriverConfig: Config{
					DisplayTimeout: "30s",
					LockTimeout:    "10es",
				},
			},
			want:    0,
			want1:   0,
			wantErr: true,
		},
		{
			name: "invalid displaytimeout vaidation",
			fields: fields{
				DriverConfig: Config{
					DisplayTimeout: "test",
					LockTimeout:    "10es",
				},
			},
			want:    0,
			want1:   0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ServiceConfig{
				DriverConfig: tt.fields.DriverConfig,
			}
			got, got1, err := c.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ServiceConfig.Validate() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ServiceConfig.Validate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
