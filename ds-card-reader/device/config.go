// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

// ServiceConfig a struct that wraps CustomConfig which holds the values for driver configuration
type ServiceConfig struct {
	DriverConfig Config
}

// DriverConfig holds the configurable options for a automated checkout
// card reader device
type Config struct {
	DeviceName       string
	DeviceSearchPath string
	VID              uint16
	PID              uint16
	SimulateDevice   bool
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
