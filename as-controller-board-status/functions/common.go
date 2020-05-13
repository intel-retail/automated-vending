// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"bytes"
	"fmt"
	"net/http"

	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

const (
	// RESTPost is a const used for REST commands using the specified method.
	RESTPost = "POST"
	// RESTPut is a const used for REST commands using the specified method.
	RESTPut = "PUT"
	// RESTGet is a const used for REST commands using the specified method.
	RESTGet = "GET"
	// ApplicationJSONContentType is a const holding the common HTTP
	// Content-Type header value, "application/json"
	ApplicationJSONContentType = "application/json"
)

// RESTCommandJSON submits a REST API call a specified restURL using the
// specified restMethod, and will serialize the inputInterface into JSON
// and submit it as part of the outbound REST request.
func (edgexconfig *ControllerBoardStatusAppSettings) RESTCommandJSON(restURL string, restMethod string, inputInterface interface{}) (err error) {
	// Serialize the inputInterface
	inputInterfaceJSON, err := utilities.GetAsJSON(inputInterface)
	if err != nil {
		return fmt.Errorf("Failed to serialize the input interface as JSON: %v", err.Error())
	}

	// Build out the request
	req, err := http.NewRequest(restMethod, restURL, bytes.NewBuffer([]byte(inputInterfaceJSON)))
	if err != nil {
		return fmt.Errorf("Failed to build the REST %v request for the URL %v due to error: %v", restMethod, restURL, err.Error())
	}
	client := &http.Client{
		Timeout: edgexconfig.RESTCommandTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to submit REST %v request due to error: %v", restMethod, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Convert the response body into a string so that it can be returned as part of the error
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(resp.Body)
		if err != nil {
			return fmt.Errorf("Did not receive an HTTP 200 status OK response from %v, instead got a response code of %v, and the response body could not be serialized", restURL, resp.StatusCode)
		}
		return fmt.Errorf("Did not receive an HTTP 200 status OK response from %v, instead got a response code of %v, and the response body was: %v", restURL, resp.StatusCode, buf.String())
	}

	return nil
}
