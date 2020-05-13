// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// buildSubscriptionMessage returns a map containing all of the fields required
// by EdgeX's notification service in order to send an email.
// For reference please visit:
// https://nexus.edgexfoundry.org/content/sites/docs/staging/master/docs/_build/html/support-notifications.html
func buildSubscriptionMessage(edgexconfig ControllerBoardStatusAppSettings) (result map[string]interface{}) {
	// Build out the subscription message
	result = map[string]interface{}{
		"slug":     edgexconfig.NotificationSlug,
		"receiver": edgexconfig.NotificationReceiver,
		"subscribedCategories": []string{
			edgexconfig.NotificationCategory,
		},
		"subscribedLabels": []string{
			edgexconfig.NotificationCategory,
		},
		"channels": []map[string]interface{}{
			{
				"type":          "EMAIL",
				"mailAddresses": edgexconfig.NotificationEmailAddresses,
			},
		},
	}

	return result
}

// PostSubscriptionToAPI attempts to perform an HTTP POST REST API call to the
// EdgeX notifications service. It will retry up to a specified number of times
// if there is an error. It considers the subscription successful if
// the API response is http.StatusCreated or http.StatusConflict.
func PostSubscriptionToAPI(edgexconfig ControllerBoardStatusAppSettings, subscriptionMessage map[string]interface{}) (err error) {
	// Serialize the subscriptionMessage so that it can be sent as part of a
	// REST request
	subscriptionMessageBytes, err := json.Marshal(subscriptionMessage)
	if err != nil {
		return fmt.Errorf("Failed to serialize the subscription message: %v", err.Error())
	}

	var resp *http.Response

	// Try no more than maxRetries times to post to the EdgeX notification service API.
	for i := 0; i < edgexconfig.NotificationSubscriptionMaxRESTRetries; i++ {
		resp, err = http.Post(edgexconfig.SubscriptionHost, ApplicationJSONContentType, bytes.NewBuffer(subscriptionMessageBytes))
		if err != nil {
			return fmt.Errorf("Failed to submit REST request to subscription API endpoint: %v", err.Error())
		}
		defer resp.Body.Close()

		// if the response has succeeded, it's either StatusCreated or StatusConflict
		isResponseSuccessful := (resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict)
		if isResponseSuccessful {
			return nil
		}

		// Wait before doing another REST call
		time.Sleep(edgexconfig.NotificationSubscriptionRESTRetryInterval)
	}

	// Read the response body so we can return the API response info
	// to the end user. We don't really care about error checking at this point,
	// since we're going to throw an error next anyway.
	respBody, _ := ioutil.ReadAll(resp.Body)

	return fmt.Errorf("REST request to subscribe to the notification service failed after %v attempts. The last API response returned a %v status code, and the response body was: %v", edgexconfig.NotificationSubscriptionMaxRESTRetries, resp.StatusCode, string(respBody))
}

// SubscribeToNotificationService configures an email notification and submits
// it to the EdgeX notification service
func SubscribeToNotificationService(edgexconfig ControllerBoardStatusAppSettings) error {
	// Build out the subscription message based on the validated app settings
	subscriptionMessage := buildSubscriptionMessage(edgexconfig)

	// Try to make the API call a few times until it works.
	err := PostSubscriptionToAPI(edgexconfig, subscriptionMessage)

	if err != nil {
		return fmt.Errorf("Failed to subscribe to the EdgeX notification service due to an error thrown while performing the HTTP POST subscription to the notification service: %v", err.Error())
	}

	return nil
}

// SendNotification performs a REST API call to the EdgeX notification service
// that will trigger the service to send a notification.
func SendNotification(edgexconfig ControllerBoardStatusAppSettings, message interface{}) error {
	notificationMessage := map[string]interface{}{
		"slug":     edgexconfig.NotificationSlugPrefix + time.Now().String(),
		"sender":   edgexconfig.NotificationSender,
		"category": edgexconfig.NotificationCategory,
		"severity": edgexconfig.NotificationSeverity,
		"content":  message,
		"labels":   edgexconfig.NotificationLabels,
	}

	notificationMessageBytes, err := json.Marshal(notificationMessage)
	if err != nil {
		return fmt.Errorf("Failed to marshal the notification message into a JSON byte array: %v", err.Error())
	}

	resp, err := http.Post(edgexconfig.NotificationHost, ApplicationJSONContentType, bytes.NewBuffer(notificationMessageBytes))
	if err != nil {
		return fmt.Errorf("Failed to perform REST POST API call to send a notification to \"%v\", error: %v", edgexconfig.NotificationHost, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("The REST API HTTP status code response from the server when attempting to send a notification was not %v, instead got: %v", http.StatusAccepted, resp.StatusCode)
	}

	return nil
}
