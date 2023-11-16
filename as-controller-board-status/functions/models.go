// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"as-controller-board-status/config"
	"fmt"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
)

// ControllerBoardStatusAppSettings is a data structure that holds the
// validated application settings (loaded from configuration.toml).
type ControllerBoardStatusAppSettings struct {
	AverageTemperatureMeasurementDuration     time.Duration
	DeviceName                                string
	MaxTemperatureThreshold                   float64
	MinTemperatureThreshold                   float64
	NotificationCategory                      string
	NotificationEmailAddresses                []string
	NotificationLabels                        []string
	NotificationReceiver                      string
	NotificationSender                        string
	NotificationSeverity                      string
	NotificationName                          string
	NotificationSubscriptionMaxRESTRetries    int
	NotificationSubscriptionRESTRetryInterval time.Duration
	NotificationThrottleDuration              time.Duration
	RESTCommandTimeout                        time.Duration
	VendingEndpoint                           string
	SubscriptionAdminState                    string
}

// ControllerBoardStatus is used to hold the data that will be passed to
// the as-vending application service, and it is marshaled into JSON when
// someone hits the GetStatus API endpoint.
type ControllerBoardStatus struct {
	Lock1                int     `json:"lock1_status"`
	Lock2                int     `json:"lock2_status"`
	DoorClosed           bool    `json:"door_closed"` // true means the door is closed and false means the door is open
	Temperature          float64 `json:"temperature"`
	Humidity             float64 `json:"humidity"`
	MinTemperatureStatus bool    `json:"minTemperatureStatus"`
	MaxTemperatureStatus bool    `json:"maxTemperatureStatus"`
}

// TempMeasurement is a simple data structure that is meant to plug temperature
// measurements and their associated timestamps into the AvgTemp function.
type TempMeasurement struct {
	Timestamp   time.Time // used to store the time that this measurement came in
	Measurement float64   // used to store the actual temperature measurement
}

// CheckBoardStatus is the primary data state holder for this application service.
// It is similar to ControllerBoardStatus, but different in that it does not
// get passed around outside of this application service. It is used to assist
// with the delivery of ControllerBoardStatus to the as-vending service.
type CheckBoardStatus struct {
	MinTemperatureThreshold                   float64
	MaxTemperatureThreshold                   float64
	MinHumidityThreshold                      float64
	MaxHumidityThreshold                      float64
	DoorClosed                                bool              // true means the door is closed and false means the door is open
	Measurements                              []TempMeasurement // used to store temperature readings over time.
	LastNotified                              time.Time         // used to store last time a notification was sent out so we don't spam the maintenance person
	Configuration                             *config.ControllerBoardStatusConfig
	SubscriptionClient                        interfaces.SubscriptionClient
	NotificationClient                        interfaces.NotificationClient
	CommandClient                             interfaces.CommandClient
	ControllerBoardStatus                     *ControllerBoardStatus
	averageTemperatureMeasurement             time.Duration
	notificationSubscriptionRESTRetryInterval time.Duration
	notificationThrottle                      time.Duration
	restCommandTimeout                        time.Duration
	notificationEmailAddresses                []string
	notificationLabels                        []string
}

func (checkBoardStatus *CheckBoardStatus) ParseStringConfigurations() error {
	var err error
	checkBoardStatus.notificationEmailAddresses = strings.Split(checkBoardStatus.Configuration.NotificationEmailAddresses, ",")
	checkBoardStatus.notificationLabels = strings.Split(checkBoardStatus.Configuration.NotificationLabels, ",")

	checkBoardStatus.averageTemperatureMeasurement, err = time.ParseDuration(checkBoardStatus.Configuration.AverageTemperatureMeasurementDuration)
	if err != nil {
		return fmt.Errorf("AverageTemperatureMeasurementDuration failed to be parsed: %v", err)
	}

	checkBoardStatus.notificationSubscriptionRESTRetryInterval, err = time.ParseDuration(checkBoardStatus.Configuration.NotificationSubscriptionRESTRetryIntervalDuration)
	if err != nil {
		return fmt.Errorf("NotificationSubscriptionRESTRetryIntervalDuration failed to be parsed: %v", err)
	}

	checkBoardStatus.notificationThrottle, err = time.ParseDuration(checkBoardStatus.Configuration.NotificationThrottleDuration)
	if err != nil {
		return fmt.Errorf("NotificationThrottleDuration failed to be parsed: %v", err)
	}

	checkBoardStatus.restCommandTimeout, err = time.ParseDuration(checkBoardStatus.Configuration.RESTCommandTimeoutDuration)
	if err != nil {
		return fmt.Errorf("RESTCommandTimeoutDuration failed to be parsed: %v", err)
	}

	return nil
}
