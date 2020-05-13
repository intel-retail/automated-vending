// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// These constants are used to simplify code repetition when working with the
// config map/struct
const (
	AverageTemperatureMeasurementDuration     = "AverageTemperatureMeasurementDuration"
	DeviceName                                = "DeviceName"
	MaxTemperatureThreshold                   = "MaxTemperatureThreshold"
	MinTemperatureThreshold                   = "MinTemperatureThreshold"
	MQTTEndpoint                              = "MQTTEndpoint"
	NotificationCategory                      = "NotificationCategory"
	NotificationEmailAddresses                = "NotificationEmailAddresses"
	NotificationHost                          = "NotificationHost"
	NotificationLabels                        = "NotificationLabels"
	NotificationReceiver                      = "NotificationReceiver"
	NotificationSender                        = "NotificationSender"
	NotificationSeverity                      = "NotificationSeverity"
	NotificationSlug                          = "NotificationSlug"
	NotificationSlugPrefix                    = "NotificationSlugPrefix"
	NotificationSubscriptionMaxRESTRetries    = "NotificationSubscriptionMaxRESTRetries"
	NotificationSubscriptionRESTRetryInterval = "NotificationSubscriptionRESTRetryInterval"
	NotificationThrottleDuration              = "NotificationThrottleDuration"
	RESTCommandTimeout                        = "RESTCommandTimeout"
	SubscriptionHost                          = "SubscriptionHost"
	VendingEndpoint                           = "VendingEndpoint"
)

// GetCommonSuccessConfig is used in test cases to quickly build out
// an example of a successful ControllerBoardStatusAppSettings configuration
func GetCommonSuccessConfig() ControllerBoardStatusAppSettings {
	return ControllerBoardStatusAppSettings{
		AverageTemperatureMeasurementDuration:     -15 * time.Second,
		DeviceName:                                "ds-controller-board",
		MaxTemperatureThreshold:                   83.0,
		MinTemperatureThreshold:                   10.0,
		MQTTEndpoint:                              "http://localhost:48082/api/v1/device/name/Inference-MQTT-device/command/vendingDoorStatus",
		NotificationCategory:                      "HW_HEALTH",
		NotificationEmailAddresses:                []string{"test@site.com", "test@site.com"},
		NotificationHost:                          "http://localhost:48060/api/v1/notification",
		NotificationLabels:                        []string{"HW_HEALTH"},
		NotificationReceiver:                      "System Administrator",
		NotificationSender:                        "Automated Checkout Maintenance Notification",
		NotificationSeverity:                      "CRITICAL",
		NotificationSlug:                          "sys-admin",
		NotificationSlugPrefix:                    "maintenance-notification",
		NotificationSubscriptionMaxRESTRetries:    10,
		NotificationSubscriptionRESTRetryInterval: 10 * time.Second,
		NotificationThrottleDuration:              1 * time.Minute,
		RESTCommandTimeout:                        15 * time.Second,
		SubscriptionHost:                          "http://localhost:48060/api/v1/subscription",
		VendingEndpoint:                           "http://localhost:48099/boardStatus",
	}
}

type testTableGetGenericErrorStruct struct {
	TestCaseName      string
	Output            string
	InputConfKey      string
	InputConfItemType reflect.Type
}

func prepGetGenericErrorTest() ([]testTableGetGenericErrorStruct, error) {
	testTimeDuration, err := time.ParseDuration("1s")
	if err != nil {
		return []testTableGetGenericErrorStruct{}, fmt.Errorf("Failed to set up time duration test value: %v", err.Error())
	}
	testTableGetGenericError := []testTableGetGenericErrorStruct{
		{
			TestCaseName:      "Time duration string input",
			Output:            fmt.Sprintf("The \"%v\" application setting has not been set. Please set this to an acceptable %v value in configuration.toml", "AverageTemperatureMeasurementDuration", "time.Duration"),
			InputConfKey:      "AverageTemperatureMeasurementDuration",
			InputConfItemType: reflect.TypeOf(testTimeDuration),
		},
		{
			TestCaseName:      "String slice input",
			Output:            fmt.Sprintf("The \"%v\" application setting has not been set. Please set this to an acceptable %v value in configuration.toml", "NotificationEmailAddresses", "[]string"),
			InputConfKey:      "NotificationEmailAddresses",
			InputConfItemType: reflect.TypeOf([]string{}),
		},
		{
			TestCaseName:      "Integer input",
			Output:            fmt.Sprintf("The \"%v\" application setting has not been set. Please set this to an acceptable %v value in configuration.toml", "NotificationSubscriptionMaxRESTRetries", "int"),
			InputConfKey:      "NotificationSubscriptionMaxRESTRetries",
			InputConfItemType: reflect.TypeOf(10),
		},
	}
	return testTableGetGenericError, nil
}

// TestGetGenericError validates that the GetGenericError function
// returns an error message with a particular format
func TestGetGenericError(t *testing.T) {
	testTable := []testTableGetGenericErrorStruct{}
	err := fmt.Errorf("")
	t.Run("GetGenericError set up test", func(t *testing.T) {
		testTable, err = prepGetGenericErrorTest()
		require.NoError(t, err)
	})
	// Test the function
	for _, testCase := range testTable {
		ct := testCase // pinning "current test" solves concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			output := GetGenericError(ct.InputConfKey, ct.InputConfItemType)
			assert.EqualError(t, output, ct.Output)
		})
	}
}

type testTableGetGenericParseErrorStruct struct {
	TestCaseName      string
	Output            string
	InputConfKey      string
	InputConfValue    string
	InputConfItemType reflect.Type
}

func prepGetGenericParseErrorTest() ([]testTableGetGenericParseErrorStruct, error) {
	testTimeDuration, err := time.ParseDuration("1s")
	if err != nil {
		return []testTableGetGenericParseErrorStruct{}, fmt.Errorf("Failed to set up time duration test value: %v", err.Error())
	}
	testTableGetGenericParseError := []testTableGetGenericParseErrorStruct{
		{
			TestCaseName:      "Time duration input",
			Output:            fmt.Sprintf("The \"%v\" application setting been set, but failed to parse the value \"%v\". Please set this to an acceptable %v value in configuration.toml", "AverageTemperatureMeasurementDuration", "-15s", "time.Duration"),
			InputConfKey:      "AverageTemperatureMeasurementDuration",
			InputConfValue:    "-15s",
			InputConfItemType: reflect.TypeOf(testTimeDuration),
		},
		{
			TestCaseName:      "String slice input",
			Output:            fmt.Sprintf("The \"%v\" application setting been set, but failed to parse the value \"%v\". Please set this to an acceptable %v value in configuration.toml", "NotificationEmailAddresses", "test@site.com,test2@site.com", "[]string"),
			InputConfKey:      "NotificationEmailAddresses",
			InputConfValue:    "test@site.com,test2@site.com",
			InputConfItemType: reflect.TypeOf([]string{}),
		},
		{
			TestCaseName:      "Integer input",
			Output:            fmt.Sprintf("The \"%v\" application setting been set, but failed to parse the value \"%v\". Please set this to an acceptable %v value in configuration.toml", "NotificationSubscriptionMaxRESTRetries", "10", "int"),
			InputConfKey:      "NotificationSubscriptionMaxRESTRetries",
			InputConfValue:    "10",
			InputConfItemType: reflect.TypeOf(10),
		},
	}
	return testTableGetGenericParseError, nil
}

// TestGetGenericParseError validates that the GetGenericParseError function
// returns an error message with a particular format
func TestGetGenericParseError(t *testing.T) {
	testTable := []testTableGetGenericParseErrorStruct{}
	err := fmt.Errorf("")
	t.Run("GetGenericParseError set up test", func(t *testing.T) {
		testTable, err = prepGetGenericParseErrorTest()
		require.NoError(t, err)
	})
	// Test the function
	for _, testCase := range testTable {
		ct := testCase // pinning "current test" solves concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			output := GetGenericParseError(ct.InputConfKey, ct.InputConfValue, ct.InputConfItemType)
			assert.EqualError(t, output, ct.Output)
		})
	}
}

type testTableProcessApplicationSettingsStruct struct {
	TestCaseName string
	Error        error
	Input        map[string]string
	Output       ControllerBoardStatusAppSettings
}

func prepProcessApplicationSettingsTest() []testTableProcessApplicationSettingsStruct {
	return []testTableProcessApplicationSettingsStruct{
		{
			TestCaseName: "Successful Test Case",
			Error:        nil,
			Input: map[string]string{
				AverageTemperatureMeasurementDuration:     "-15s",
				DeviceName:                                "ds-controller-board",
				MaxTemperatureThreshold:                   "83",
				MinTemperatureThreshold:                   "10",
				MQTTEndpoint:                              "http://localhost:48082/api/v1/device/name/Inference-MQTT-device/command/vendingDoorStatus",
				NotificationCategory:                      "HW_HEALTH",
				NotificationEmailAddresses:                "test@site.com,test@site.com",
				NotificationHost:                          "http://localhost:48060/api/v1/notification",
				NotificationLabels:                        "HW_HEALTH",
				NotificationReceiver:                      "System Administrator",
				NotificationSender:                        "Automated Checkout Maintenance Notification",
				NotificationSeverity:                      "CRITICAL",
				NotificationSlug:                          "sys-admin",
				NotificationSlugPrefix:                    "maintenance-notification",
				NotificationSubscriptionMaxRESTRetries:    "10",
				NotificationSubscriptionRESTRetryInterval: "10s",
				NotificationThrottleDuration:              "1m",
				RESTCommandTimeout:                        "15s",
				SubscriptionHost:                          "http://localhost:48060/api/v1/subscription",
				VendingEndpoint:                           "http://localhost:48099/boardStatus",
			},
			Output: ControllerBoardStatusAppSettings{
				AverageTemperatureMeasurementDuration:     -15 * time.Second,
				DeviceName:                                "ds-controller-board",
				MaxTemperatureThreshold:                   83.0,
				MinTemperatureThreshold:                   10.0,
				MQTTEndpoint:                              "http://localhost:48082/api/v1/device/name/Inference-MQTT-device/command/vendingDoorStatus",
				NotificationCategory:                      "HW_HEALTH",
				NotificationEmailAddresses:                []string{"test@site.com", "test@site.com"},
				NotificationHost:                          "http://localhost:48060/api/v1/notification",
				NotificationLabels:                        []string{"HW_HEALTH"},
				NotificationReceiver:                      "System Administrator",
				NotificationSender:                        "Automated Checkout Maintenance Notification",
				NotificationSeverity:                      "CRITICAL",
				NotificationSlug:                          "sys-admin",
				NotificationSlugPrefix:                    "maintenance-notification",
				NotificationSubscriptionMaxRESTRetries:    10,
				NotificationSubscriptionRESTRetryInterval: 10 * time.Second,
				NotificationThrottleDuration:              1 * time.Minute,
				RESTCommandTimeout:                        15 * time.Second,
				SubscriptionHost:                          "http://localhost:48060/api/v1/subscription",
				VendingEndpoint:                           "http://localhost:48099/boardStatus",
			},
		},
		{
			TestCaseName: "Item Not Specified Case",
			Error:        GetGenericError("AverageTemperatureMeasurementDuration", reflect.TypeOf(1*time.Second)),
			Input: map[string]string{
				// AverageTemperatureMeasurementDuration:     "-15s",
				DeviceName:                                "ds-controller-board",
				MaxTemperatureThreshold:                   "83",
				MinTemperatureThreshold:                   "10",
				MQTTEndpoint:                              "http://localhost:48082/api/v1/device/name/Inference-MQTT-device/command/vendingDoorStatus",
				NotificationCategory:                      "HW_HEALTH",
				NotificationEmailAddresses:                "test@site.com,test@site.com",
				NotificationHost:                          "http://localhost:48060/api/v1/notification",
				NotificationLabels:                        "HW_HEALTH",
				NotificationReceiver:                      "System Administrator",
				NotificationSender:                        "Automated Checkout Maintenance Notification",
				NotificationSeverity:                      "CRITICAL",
				NotificationSlug:                          "sys-admin",
				NotificationSlugPrefix:                    "maintenance-notification",
				NotificationSubscriptionMaxRESTRetries:    "10",
				NotificationSubscriptionRESTRetryInterval: "10s",
				NotificationThrottleDuration:              "1m",
				RESTCommandTimeout:                        "15s",
				SubscriptionHost:                          "http://localhost:48060/api/v1/subscription",
				VendingEndpoint:                           "http://localhost:48099/boardStatus",
			},
			Output: ControllerBoardStatusAppSettings{},
		},
		{
			TestCaseName: "Invalid Time Duration Parse Case",
			Error:        GetGenericParseError("AverageTemperatureMeasurementDuration", "invalid time.ParseDuration string", reflect.TypeOf(1*time.Second)),
			Input: map[string]string{
				AverageTemperatureMeasurementDuration:     "invalid time.ParseDuration string",
				DeviceName:                                "ds-controller-board",
				MaxTemperatureThreshold:                   "83",
				MinTemperatureThreshold:                   "10",
				MQTTEndpoint:                              "http://localhost:48082/api/v1/device/name/Inference-MQTT-device/command/vendingDoorStatus",
				NotificationCategory:                      "HW_HEALTH",
				NotificationEmailAddresses:                "test@site.com,test@site.com",
				NotificationHost:                          "http://localhost:48060/api/v1/notification",
				NotificationLabels:                        "HW_HEALTH",
				NotificationReceiver:                      "System Administrator",
				NotificationSender:                        "Automated Checkout Maintenance Notification",
				NotificationSeverity:                      "CRITICAL",
				NotificationSlug:                          "sys-admin",
				NotificationSlugPrefix:                    "maintenance-notification",
				NotificationSubscriptionMaxRESTRetries:    "10",
				NotificationSubscriptionRESTRetryInterval: "10s",
				NotificationThrottleDuration:              "1m",
				RESTCommandTimeout:                        "15s",
				SubscriptionHost:                          "http://localhost:48060/api/v1/subscription",
				VendingEndpoint:                           "http://localhost:48099/boardStatus",
			},
			Output: ControllerBoardStatusAppSettings{},
		},
		{
			TestCaseName: "Invalid float64 Parse Case",
			Error:        GetGenericParseError("MaxTemperatureThreshold", "invalid float64 value", reflect.TypeOf(83.0)),
			Input: map[string]string{
				AverageTemperatureMeasurementDuration:     "-15s",
				DeviceName:                                "ds-controller-board",
				MaxTemperatureThreshold:                   "invalid float64 value",
				MinTemperatureThreshold:                   "10",
				MQTTEndpoint:                              "http://localhost:48082/api/v1/device/name/Inference-MQTT-device/command/vendingDoorStatus",
				NotificationCategory:                      "HW_HEALTH",
				NotificationEmailAddresses:                "test@site.com,test@site.com",
				NotificationHost:                          "http://localhost:48060/api/v1/notification",
				NotificationLabels:                        "HW_HEALTH",
				NotificationReceiver:                      "System Administrator",
				NotificationSender:                        "Automated Checkout Maintenance Notification",
				NotificationSeverity:                      "CRITICAL",
				NotificationSlug:                          "sys-admin",
				NotificationSlugPrefix:                    "maintenance-notification",
				NotificationSubscriptionMaxRESTRetries:    "10",
				NotificationSubscriptionRESTRetryInterval: "10s",
				NotificationThrottleDuration:              "1m",
				RESTCommandTimeout:                        "15s",
				SubscriptionHost:                          "http://localhost:48060/api/v1/subscription",
				VendingEndpoint:                           "http://localhost:48099/boardStatus",
			},
			Output: ControllerBoardStatusAppSettings{},
		},
		{
			TestCaseName: "Invalid int Parse Case",
			Error:        GetGenericParseError("NotificationSubscriptionMaxRESTRetries", "invalid int value", reflect.TypeOf(10)),
			Input: map[string]string{
				AverageTemperatureMeasurementDuration:     "-15s",
				DeviceName:                                "ds-controller-board",
				MaxTemperatureThreshold:                   "83",
				MinTemperatureThreshold:                   "10",
				MQTTEndpoint:                              "http://localhost:48082/api/v1/device/name/Inference-MQTT-device/command/vendingDoorStatus",
				NotificationCategory:                      "HW_HEALTH",
				NotificationEmailAddresses:                "test@site.com,test@site.com",
				NotificationHost:                          "http://localhost:48060/api/v1/notification",
				NotificationLabels:                        "HW_HEALTH",
				NotificationReceiver:                      "System Administrator",
				NotificationSender:                        "Automated Checkout Maintenance Notification",
				NotificationSeverity:                      "CRITICAL",
				NotificationSlug:                          "sys-admin",
				NotificationSlugPrefix:                    "maintenance-notification",
				NotificationSubscriptionMaxRESTRetries:    "invalid int value",
				NotificationSubscriptionRESTRetryInterval: "10s",
				NotificationThrottleDuration:              "1m",
				RESTCommandTimeout:                        "15s",
				SubscriptionHost:                          "http://localhost:48060/api/v1/subscription",
				VendingEndpoint:                           "http://localhost:48099/boardStatus",
			},
			Output: ControllerBoardStatusAppSettings{},
		},
	}
}

// TestProcessApplicationSettings validates that the ProcessApplicationSettings
// function converts a map from the EdgeX SDK into a
// ControllerBoardStatusAppSettings struct with properly parsed fields
func TestProcessApplicationSettings(t *testing.T) {
	testTable := prepProcessApplicationSettingsTest()
	for _, testCase := range testTable {
		ct := testCase // pinning "current test" solves concurrency issues
		t.Run("ProcessApplicationSettings test table", func(t *testing.T) {
			output, err := ProcessApplicationSettings(testCase.Input)
			assert.Equal(t, err, ct.Error)
			assert.Equal(t, output, ct.Output)
		})
	}
}
