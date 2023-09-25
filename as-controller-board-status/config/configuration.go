// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package config

import (
	"fmt"
	"reflect"
)

type ServiceConfig struct {
	ControllerBoardStatus ControllerBoardStatusConfig
}

// ControllerBoardStatusConfig is a data structure that holds the
// validated configuration settings (loaded from configuration.toml).
type ControllerBoardStatusConfig struct {
	AverageTemperatureMeasurementDuration             string
	DeviceName                                        string
	MaxTemperatureThreshold                           float64
	MinTemperatureThreshold                           float64
	DoorStatusCommandEndpoint                         string
	NotificationCategory                              string
	NotificationEmailAddresses                        string
	NotificationLabels                                string
	NotificationReceiver                              string
	NotificationSender                                string
	NotificationSeverity                              string
	NotificationName                                  string
	NotificationSubscriptionMaxRESTRetries            int
	NotificationSubscriptionRESTRetryIntervalDuration string
	NotificationThrottleDuration                      string
	RESTCommandTimeoutDuration                        string
	VendingEndpoint                                   string
	SubscriptionAdminState                            string
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
func (bs *ControllerBoardStatusConfig) Validate() error {
	config := reflect.ValueOf(*bs)
	configType := config.Type()

	for i := 0; i < config.NumField(); i++ {
		field := config.Field(i).Interface()
		fieldName := configType.Field(i).Name

		if _, ok := field.(string); ok && len(field.(string)) == 0 {
			return fmt.Errorf("%v is empty", fieldName)
		}

		if _, ok := field.(float64); ok && field.(float64) == 0.0 {
			return fmt.Errorf("%v is set to 0", fieldName)
		}

		if _, ok := field.(int); ok && field.(int) == 0 {
			return fmt.Errorf("%v is set to 0", fieldName)
		}
	}
	return nil
}
