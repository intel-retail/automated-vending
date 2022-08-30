// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package config

import "errors"

type ServiceConfig struct {
	AppCustom AppCustomConfig
}

// AppCustomConfig is custom structured configuration that is specified in the service's
// configuration.toml file and Configuration Provider (aka Consul), if enabled.
type AppCustomConfig struct {
	InventoryEndpoint string
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

// Validate ensures your custom configuration has proper values.
func (ac *AppCustomConfig) Validate() error {

	if len(ac.InventoryEndpoint) == 0 {
		return errors.New("AppCustom.InventoryEndpoint can not be empty")
	}
	return nil
}
