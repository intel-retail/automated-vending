// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"fmt"
	"time"
)

// ServiceConfig a struct that wraps CustomConfig which holds the values for driver configuration
type ServiceConfig struct {
	DriverConfig Config
}

// Config is the global device configuration, which is populated by values in
// the "Driver" section of res/configuration.toml
type Config struct {
	DeviceName             string
	VirtualControllerBoard bool
	PID                    string
	VID                    string
	DisplayTimeout         string
	LockTimeout            string
}

// UpdateFromRaw updates the service's full configuration from raw data received from
// the Service Provider.
func (c *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false
	}

	*c = *configuration

	return true
}

func (c *ServiceConfig) Validate() (time.Duration, time.Duration, error) {

	displayTimeout, err := time.ParseDuration(c.DriverConfig.DisplayTimeout)
	if err != nil {
		return 0, 0, fmt.Errorf("display timeout configuration is not a proper time duration: %v", err)
	}

	lockTimeout, err := time.ParseDuration(c.DriverConfig.LockTimeout)
	if err != nil {
		return 0, 0, fmt.Errorf("lock timeout configuration is not a proper time duration: %v", err)
	}

	return displayTimeout, lockTimeout, nil
}
