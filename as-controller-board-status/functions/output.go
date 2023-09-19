// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
)

const (
	minimum = "minimum"
	maximum = "maximum"
	// ControllerBoardDeviceServiceDeviceName is the name of the EdgeX device
	// corresponding to our upstream event source.
	ControllerBoardDeviceServiceDeviceName = "controller-board"
)

// CheckControllerBoardStatus is an EdgeX function that is passed into the EdgeX SDK's function pipeline.
// It is a decision function that allows for multiple devices to have their events processed
// correctly by this application service. In this case, only one unique type of EdgeX device will come
// through to this function, but in general this is basically a template function that is also followed
// in other services in the Automated Vending project.
func (boardStatus *CheckBoardStatus) CheckControllerBoardStatus(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		// We didn't receive a result
		return false, nil
	}

	// Declare shorthand for the LoggingClient
	lc := ctx.LoggingClient()

	event := data.(dtos.Event)

	if event.DeviceName == ControllerBoardDeviceServiceDeviceName {
		for _, eventReading := range event.Readings {
			if len(eventReading.Value) < 1 {
				return false, fmt.Errorf("event reading was empty")
			}

			lc.Debugf("Received event reading value: %s", eventReading.Value)

			// Unmarshal the event reading data into the global controllerBoardStatus variable
			err := json.Unmarshal([]byte(eventReading.Value), &boardStatus.ControllerBoardStatus)
			if err != nil {
				lc.Errorf("Failed to unmarshal controller board data %s: %s", eventReading.Value, err.Error())
				return false, nil
			}

			// Check if the temperature thresholds have been exceeded
			err = boardStatus.processTemperature(lc, boardStatus.ControllerBoardStatus.Temperature)
			if err != nil {
				lc.Errorf("Encountered error while checking temperature thresholds: %s", err.Error())
			}

			// Check if the door open/closed state requires action
			err = boardStatus.processVendingDoorState(lc, boardStatus.ControllerBoardStatus.DoorClosed)
			if err != nil {
				lc.Errorf("Encountered error while checking the open/closed state of the door: %s", err.Error())
			}
		}
	}

	return true, event // Continues the functions pipeline execution with the current event
}

// processTemperatureMeasurements takes a single temperature measurement (which
// presumably is coming straight from an EdgeX reading) and adds it to the
// boardStatus.Measurements slice
func (boardStatus *CheckBoardStatus) processTemperatureMeasurements(temperature float64) float64 {
	// Start by storing the latest data from the controller board as a
	// temperature & time measurement
	newMeasurement := TempMeasurement{
		Timestamp:   time.Now(),
		Measurement: temperature,
	}

	// Update the list of measurements to include the new measurement from
	// EdgeX
	boardStatus.Measurements = append(boardStatus.Measurements, newMeasurement)

	avgTemp, cutIndex := AvgTemp(boardStatus.Measurements, boardStatus.averageTemperatureMeasurement)

	// Only keep track of the measurements used to calculate the latest average
	// temperature
	boardStatus.Measurements = boardStatus.Measurements[:cutIndex]

	return avgTemp
}

func (controllerBoardStatus *ControllerBoardStatus) updateThresholdsFromAverageTemperature(avgTemp float64, maxTemp float64, minTemp float64) {
	// If the average temperature over the last X duration exceeds
	// the maximum threshold temperature as configured in the application
	// settings, switch the state accordingly
	if avgTemp >= maxTemp {
		controllerBoardStatus.MaxTemperatureStatus = true
	} else {
		controllerBoardStatus.MaxTemperatureStatus = false
	}

	// Similarly, switch the state accordingly if the minimum threshold
	// as defined in the settings is greater than the average temperature
	if avgTemp <= minTemp {
		controllerBoardStatus.MinTemperatureStatus = true
	} else {
		controllerBoardStatus.MinTemperatureStatus = false
	}
}

// getTempThresholdExceededMessage builds out a "max/min temperature threshold
// exceeded" message string and returns it
func getTempThresholdExceededMessage(minOrMax string, avgTemp float64, tempThreshold float64) (string, error) {
	if minOrMax != maximum && minOrMax != minimum {
		return "", fmt.Errorf("Please specify minOrMax as \"%v\" or \"%v\", the value given was \"%v\"", maximum, minimum, minOrMax)
	}
	resultMessage := fmt.Sprintf("The internal automated vending's temperature is currently %.2f, and this temperature exceeds the configured %v temperature threshold of %v degrees. The automated vending needs maintenance as of: %s", avgTemp, minOrMax, tempThreshold, time.Now().Format("_2 Jan, Mon | 3:04PM MST"))
	return resultMessage, nil
}

// sendTempThresholdExceededNotification leverages the SendNotification
// function to submit an EdgeX REST call to send a notification to a user.
// It does not check if a notification needs to be sent, it simply sends it
func (boardStatus *CheckBoardStatus) sendTempThresholdExceededNotification(message string) error {
	err := boardStatus.SendNotification(message)
	if err != nil {
		return fmt.Errorf("Encountered error sending notification for exceeding temperature threshold: %v", err.Error())
	}
	boardStatus.LastNotified = time.Now()
	return nil
}

// sendTempThresholdExceededNotifications reviews the state of
// controllerBoardStatus and sends notifications accordingly
func (boardStatus *CheckBoardStatus) sendTempThresholdExceededNotifications(avgTemp float64) error {
	// For efficient coding, build out a simple map that contains keys
	// only if there is a message that needs to be sent when the
	// min/max thresholds are exceeded, then loop over that map
	messages := make(map[string]float64)
	if boardStatus.ControllerBoardStatus.MaxTemperatureStatus {
		messages[maximum] = boardStatus.Configuration.MaxTemperatureThreshold
	}
	if boardStatus.ControllerBoardStatus.MinTemperatureStatus {
		messages[minimum] = boardStatus.Configuration.MinTemperatureThreshold
	}
	for minMaxStr, tempThresholdValueFloat := range messages {
		// Build the message out
		tempThresholdMessage, err := getTempThresholdExceededMessage(minMaxStr, avgTemp, tempThresholdValueFloat)
		if err != nil {
			return fmt.Errorf("Encountered error building out the %v temperature threshold message: %v", minMaxStr, err.Error())
		}
		// Send the notification
		err = boardStatus.sendTempThresholdExceededNotification(tempThresholdMessage)
		if err != nil {
			return fmt.Errorf("Encountered error sending the %v temperature threshold message: %v", minMaxStr, err.Error())
		}
	}
	return nil
}

// processTemperature checks to see if we've exceeded any temperature thresholds
// and submits EdgeX REST commands accordingly
func (boardStatus *CheckBoardStatus) processTemperature(lc logger.LoggingClient, temperature float64) error {
	avgTemp := boardStatus.processTemperatureMeasurements(temperature)

	// Update the min/max temperature status readout for the global controller
	// board status according to the how the average temperature compares to
	// the configured min/max temperature threshold values
	boardStatus.ControllerBoardStatus.updateThresholdsFromAverageTemperature(avgTemp, boardStatus.Configuration.MaxTemperatureThreshold, boardStatus.Configuration.MinTemperatureThreshold)

	// Take note of whether or not we've sent a notification within a duration
	// not allowable by the user's configuration
	notificationSentRecently := (boardStatus.notificationThrottle > time.Since(boardStatus.LastNotified))

	// Send a notification if the temperature has exceeded thresholds,
	// and if we have not sent a notification recently
	if !notificationSentRecently {
		err := boardStatus.sendTempThresholdExceededNotifications(avgTemp)
		if err != nil {
			return fmt.Errorf("Failed to send temperature threshold exceeded notification(s) due to error: %v", err.Error())
		}
	}

	// If either the minimum or maximum temperature thresholds have been
	// exceeded, send the current state to the central service so it can
	// react accordingly
	if boardStatus.ControllerBoardStatus.MinTemperatureStatus || boardStatus.ControllerBoardStatus.MaxTemperatureStatus {
		lc.Info("Pushing controller board status to central vending service due to a temperature threshold being exceeded")
		err := boardStatus.RESTCommandJSON(boardStatus.Configuration.VendingEndpoint, http.MethodPost, boardStatus.ControllerBoardStatus)
		if err != nil {
			return fmt.Errorf("Encountered error sending the controller board's status to the central vending endpoint: %v", err.Error())
		}
	}
	return nil
}

// AvgTemp takes a slice of temperature measurements and returns a proper
// average value of the values in the slice.
func AvgTemp(measurements []TempMeasurement, duration time.Duration) (float64, int) {
	var z int
	var mCount float64
	var tempSum, avgTemp float64 = 0.00, 0.00

	// Sort the slice so that correct number of measurements can be averaged
	sort.Slice(measurements, func(x, y int) bool {
		return measurements[x].Timestamp.After(measurements[y].Timestamp)
	})

	for z < len(measurements) {
		if measurements[z].Timestamp.Before(measurements[0].Timestamp.Add(duration)) {
			mCount = float64(z)
			avgTemp = tempSum / mCount
			break
		}
		tempSum = tempSum + measurements[z].Measurement
		avgTemp = (tempSum) / float64(z+1)
		z = z + 1
	}
	return avgTemp, z
}

// processVendingDoorState checks to see if the vending door state has changed
// and if it has changed, it will then submit EdgeX commands (REST calls)
// to the MQTT device service and the central vending state endpoint.
func (boardStatus *CheckBoardStatus) processVendingDoorState(lc logger.LoggingClient, doorClosed bool) error {
	if boardStatus.DoorClosed != doorClosed {
		// Set the boardStatus's DoorClosed value to the new value
		boardStatus.DoorClosed = doorClosed
		lc.Infof("The door closed status has changed to: %t", doorClosed)

		// Set the door closed state and make sure MinTemp and MaxTemp status
		// are false to avoid triggering a false temperature event
		err := boardStatus.RESTCommandJSON(boardStatus.Configuration.VendingEndpoint, http.MethodPost, ControllerBoardStatus{
			DoorClosed:           doorClosed,
			MinTemperatureStatus: false,
			MaxTemperatureStatus: false,
		})
		if err != nil {
			return fmt.Errorf("failed to submit the controller board's status to the central vending state service: %v", err.Error())
		}

		// Prepare a message to be sent to the MQTT bus. Depending on the state
		// of the door, this message may trigger a CV inference
		err = boardStatus.RESTCommandJSON(boardStatus.Configuration.DoorStatusCommandEndpoint, http.MethodPut, VendingDoorStatus{
			VendingDoorStatus: strconv.FormatBool(doorClosed),
		})
		if err != nil {
			return fmt.Errorf("failed to submit the vending door state to the MQTT device service: %v", err.Error())
		}
	}

	return nil
}
