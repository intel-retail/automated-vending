// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ControllerBoardStatusAppSettings is a data structure that holds the
// validated application settings (loaded from configuration.toml).
type ControllerBoardStatusAppSettings struct {
	AverageTemperatureMeasurementDuration     time.Duration
	DeviceName                                string
	MaxTemperatureThreshold                   float64
	MinTemperatureThreshold                   float64
	MQTTEndpoint                              string
	NotificationCategory                      string
	NotificationEmailAddresses                []string
	NotificationHost                          string
	NotificationLabels                        []string
	NotificationReceiver                      string
	NotificationSender                        string
	NotificationSeverity                      string
	NotificationSlug                          string
	NotificationSlugPrefix                    string
	NotificationSubscriptionMaxRESTRetries    int
	NotificationSubscriptionRESTRetryInterval time.Duration
	NotificationThrottleDuration              time.Duration
	RESTCommandTimeout                        time.Duration
	SubscriptionHost                          string
	VendingEndpoint                           string
}

// GetGenericError helps keep the ProcessApplicationSettings slim by enabling
// re-use of a common error format that should be returned to the user when
// the user fails to specify a configuration item.
func GetGenericError(confKey string, confItemType reflect.Type) error {
	return fmt.Errorf("The \"%v\" application setting has not been set. Please set this to an acceptable %v value in configuration.toml", confKey, confItemType)
}

// GetGenericParseError helps keep the ProcessApplicationSettings slim by enabling
// re-use of a common error format that should be returned to the user when
// the user fails to specify a parseable configuration item.
func GetGenericParseError(confKey string, confValue interface{}, confItemType reflect.Type) error {
	return fmt.Errorf("The \"%v\" application setting been set, but failed to parse the value \"%v\". Please set this to an acceptable %v value in configuration.toml", confKey, confValue, confItemType)
}

// ProcessApplicationSettings returns a configuration containing all of the
// EdgeX settings. This function should be called every time a function
// in the EdgeX pipeline is processed.
func ProcessApplicationSettings(applicationSettings map[string]string) (config ControllerBoardStatusAppSettings, err error) {
	// Iterate over all of the fields of the config
	reflectedValueOfConfig := reflect.ValueOf(config)
	reflectedFieldsOfConfig := reflectedValueOfConfig.Type()
	numberOfFieldsInStruct := reflectedValueOfConfig.NumField()
	configValues := make([]interface{}, numberOfFieldsInStruct)
	configFields := make([]string, numberOfFieldsInStruct)
	configTypes := make([]reflect.Type, numberOfFieldsInStruct)

	configAsMap := make(map[string]interface{})

	// For each field in the ControllerBoardStatusAppSettings struct,
	// attempt to locate that corresponding field in the applicationSettings
	// map
	for i := 0; i < reflectedValueOfConfig.NumField(); i++ {
		configValues[i] = reflectedValueOfConfig.Field(i).Interface()
		configFields[i] = reflectedFieldsOfConfig.Field(i).Name
		configTypes[i] = reflectedValueOfConfig.Field(i).Type()

		// Fetch the current field from the application settings map
		applicationSettingsValue, ok := applicationSettings[configFields[i]]
		if !ok {
			return config, GetGenericError(configFields[i], configTypes[i])
		}

		// Parse the field according to the application settings map
		if applicationSettingsValue != "" {
			switch configValues[i].(type) {
			case float64:
				configAsMap[configFields[i]], err = strconv.ParseFloat(applicationSettingsValue, 64)
				if err != nil {
					return config, GetGenericParseError(configFields[i], applicationSettingsValue, configTypes[i])
				}
			case string:
				configAsMap[configFields[i]] = applicationSettingsValue
			case []string:
				configAsMap[configFields[i]] = strings.Split(applicationSettingsValue, ",")
			case time.Duration:
				configAsMap[configFields[i]], err = time.ParseDuration(applicationSettingsValue)
				if err != nil {
					return config, GetGenericParseError(configFields[i], applicationSettingsValue, configTypes[i])
				}
			case int:
				configAsMap[configFields[i]], err = strconv.Atoi(applicationSettingsValue)
				if err != nil {
					return config, GetGenericParseError(configFields[i], applicationSettingsValue, configTypes[i])
				}
			default:
				return config, GetGenericParseError(configFields[i], applicationSettingsValue, configTypes[i])
			}
		}
	}

	// Encode into JSON, then unmarshal into the
	// ControllerBoardStatusAppSettings struct
	configJSON, err := json.Marshal(configAsMap)
	if err != nil {
		return config, fmt.Errorf("Failed to convert the config map into a JSON byte slice: %v", err.Error())
	}

	// Unmarshal into the config struct
	err = json.Unmarshal(configJSON, &config)
	if err != nil {
		return config, fmt.Errorf("Failed to unmarshal the config map JSON into the config struct: %v", err.Error())
	}

	return config, nil
}
