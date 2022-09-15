// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"time"
)

// CustomConfig holds the values for the driver configuration
type CustomConfig struct {
	DriverConfig Config
}

// ServiceConfig a struct that wraps CustomConfig which holds the values for driver configuration
type ServiceConfig struct {
	AppCustom CustomConfig
}

// Config is the global device configuration, which is populated by values in
// the "Driver" section of res/configuration.toml
type Config struct {
	VirtualControllerBoard bool
	PID                    string
	VID                    string
	DisplayTimeout         time.Duration
	LockTimeout            time.Duration
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
