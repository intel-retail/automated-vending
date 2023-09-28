// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// ApplicationJSONContentType is a const holding the common HTTP
	// Content-Type header value, "application/json"
	ApplicationJSONContentType = "application/json"
)

// RESTCommandJSON submits a REST API call a specified restURL using the
// specified restMethod, and will serialize the inputInterface into JSON
// and submit it as part of the outbound REST request.
func (boardStatus *CheckBoardStatus) RESTCommandJSON(restURL string, restMethod string, inputInterface interface{}) (err error) {
	// Serialize the inputInterface
	inputInterfaceJSON, err := json.Marshal(inputInterface)

	if err != nil {
		return fmt.Errorf("failed to serialize the input interface as JSON: %v", err.Error())
	}

	// Build out the request
	req, err := http.NewRequest(restMethod, restURL, bytes.NewBuffer(inputInterfaceJSON))
	if err != nil {
		return fmt.Errorf("failed to build the REST %v request for the URL %v due to error: %v", restMethod, restURL, err.Error())
	}
	client := &http.Client{
		Timeout: boardStatus.restCommandTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit REST %v request due to error: %v", restMethod, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Convert the response body into a string so that it can be returned as part of the error
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(resp.Body)
		if err != nil {
			return fmt.Errorf("did not receive an HTTP 200 status OK response from %v, instead got a response code of %v, and the response body could not be serialized", restURL, resp.StatusCode)
		}
		return fmt.Errorf("did not receive an HTTP 200 status OK response from %v, instead got a response code of %v, and the response body was: %v", restURL, resp.StatusCode, buf.String())
	}

	return nil
}
