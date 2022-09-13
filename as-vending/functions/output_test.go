// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	err := m.Run()
	os.Exit(err)
}

// TestGetMaintenanceMode tests the HTTP GET endpoint '/maintenanceMode' to
// verify that it reports the correct value of MaintenanceMode in its instance
// of VendingState.
func TestGetMaintenanceMode(t *testing.T) {
	maintModeTrue := MaintenanceMode{
		MaintenanceMode: true,
	}
	maintModeFalse := MaintenanceMode{
		MaintenanceMode: false,
	}
	t.Run("TestGetMaintenanceMode MaintenanceMode=True", func(t *testing.T) {
		var vendingState VendingState
		var maintModeAPIResponse MaintenanceMode

		// set the vendingState's MaintenanceMode boolean accordingly
		vendingState.MaintenanceMode = true

		req := httptest.NewRequest("GET", "/maintenanceMode", nil)
		w := httptest.NewRecorder()

		// run the actual function in question
		vendingState.GetMaintenanceMode(w, req)

		// parse the response
		resp := w.Result()
		_, err := utilities.ParseJSONHTTPResponseContent(resp.Body, &maintModeAPIResponse)
		require.NoError(t, err)

		defer resp.Body.Close()
		assert.Equal(t, maintModeAPIResponse, maintModeTrue, "Received a maintenance mode response that was different than anticipated")
	})
	t.Run("TestGetMaintenanceMode MaintenanceMode=False", func(t *testing.T) {
		var vendingState VendingState
		var maintModeAPIResponse MaintenanceMode

		// set the vendingState's MaintenanceMode boolean accordingly
		vendingState.MaintenanceMode = false

		req := httptest.NewRequest("GET", "/maintenanceMode", nil)
		w := httptest.NewRecorder()

		// run the actual function in question
		vendingState.GetMaintenanceMode(w, req)

		// parse the response
		resp := w.Result()
		_, err := utilities.ParseJSONHTTPResponseContent(resp.Body, &maintModeAPIResponse)
		require.NoError(t, err)

		defer resp.Body.Close()
		assert.Equal(t, maintModeAPIResponse, maintModeFalse, "Received a maintenance mode response that was different than anticipated")
	})
	t.Run("TestGetMaintenanceMode OPTIONS", func(t *testing.T) {
		var vendingState VendingState

		req := httptest.NewRequest("OPTIONS", "/maintenanceMode", nil)
		w := httptest.NewRecorder()

		vendingState.GetMaintenanceMode(w, req)

		// parse the response
		resp := w.Result()
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "Parsing response body threw error")
		assert.Equal(t, string(body), "", "Response body was not an empty string, expected it to be empty for a pre-flight CORS OPTIONS response")
		assert.Equal(t, resp.Status, "200 OK", "OPTIONS request did not return 200")
	})
}

func TestCheckInferenceStatus(t *testing.T) {
	testCases := []struct {
		TestCaseName string
		statusCode   int
		Expected     bool
	}{
		{"Successful case", http.StatusOK, true},
		{"Negative case", http.StatusInternalServerError, false},
	}

	for _, tc := range testCases {

		t.Run(tc.TestCaseName, func(t *testing.T) {

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, err := w.Write([]byte{})
				if err != nil {
					t.Fatal(err.Error())
				}
			}))
			defer testServer.Close()
			assert.Equal(t, tc.Expected, checkInferenceStatus(logger.NewMockClient(), testServer.URL), "Expected value to match output")
		})
	}

}

func TestGetCardAuthInfo(t *testing.T) {
	testCases := []struct {
		TestCaseName string
		statusCode   int
		cardID       string
		Expected     string
	}{
		{"Successful case", http.StatusOK, "1234567890", "1234567890"},
		{"Internal error case", http.StatusInternalServerError, "1234567890", ""},
	}

	var vendingState VendingState

	for _, tc := range testCases {

		t.Run(tc.TestCaseName, func(t *testing.T) {

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				output := OutputData{
					CardID: tc.cardID,
				}
				authDataJSON, _ := utilities.GetAsJSON(output)
				utilities.WriteJSONHTTPResponse(w, r, tc.statusCode, authDataJSON, false)
			}))

			defer testServer.Close()
			vendingState.getCardAuthInfo(logger.NewMockClient(), testServer.URL, tc.cardID)
			assert.Equal(t, tc.Expected, vendingState.CurrentUserData.CardID, "Expected value to match output")
		})
	}
}

func TestResetDoorLock(t *testing.T) {
	stopChannel := make(chan int)
	var vendingState VendingState
	vendingState.ThreadStopChannel = stopChannel

	request, _ := http.NewRequest("POST", "", nil)
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(vendingState.ResetDoorLock)
	handler.ServeHTTP(recorder, request)

	assert.Equal(t, false, vendingState.MaintenanceMode, "MaintanceMode should be false")
	assert.Equal(t, false, vendingState.CVWorkflowStarted, "CVWorkflowStarted should be false")
	assert.Equal(t, false, vendingState.DoorClosed, "DoorClosed should be false")
	assert.Equal(t, false, vendingState.DoorClosedDuringCVWorkflow, "DoorClosedDuringCVWorkflow should be false")
	assert.Equal(t, false, vendingState.DoorOpenedDuringCVWorkflow, "DoorOpenedDuringCVWorkflow should be false")
	assert.Equal(t, false, vendingState.InferenceDataReceived, "InferenceDataReceived should be false")
}

func TestDisplayLedger(t *testing.T) {
	// Http test servers
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte{})
		if err != nil {
			t.Fatal(err.Error())
		}
	}))

	vendingState := VendingState{
		Configuration: &ServiceConfiguration{
			LCDRowLength:                      20,
			DeviceControllerBoarddisplayReset: testServer.URL,
			DeviceControllerBoarddisplayRow1:  testServer.URL,
		},
	}

	ledger := Ledger{
		LineItems: []LineItem{{ProductName: "itemX", ItemCount: 2, ItemPrice: 1.50, SKU: "1234"}},
	}
	err := vendingState.displayLedger(logger.NewMockClient(), ledger)
	assert.NoError(t, err)
}

// TODO: BoardStatus handler needs to return proper http status code for unit tests
func TestBoardStatus(t *testing.T) {
	doorOpenStopChannel := make(chan int)
	doorCloseStopChannel := make(chan int)

	vendingState := VendingState{
		DoorClosed:                     false,
		CVWorkflowStarted:              true,
		DoorOpenWaitThreadStopChannel:  doorOpenStopChannel,
		DoorCloseWaitThreadStopChannel: doorCloseStopChannel,
		Configuration:                  new(ServiceConfiguration),
	}
	boardStatus := ControllerBoardStatus{
		MaxTemperatureStatus: true,
		MinTemperatureStatus: true,
		DoorClosed:           true,
	}

	b, _ := json.Marshal(boardStatus)

	request, _ := http.NewRequest("POST", "", bytes.NewBuffer(b))
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(vendingState.BoardStatus)
	handler.ServeHTTP(recorder, request)
}

func TestHandleMqttDeviceReading(t *testing.T) {
	testCases := []struct {
		TestCaseName string
		statusCode   int
		Expected     error
	}{
		{"Successful case", http.StatusOK, nil},
		{"Internal error case", http.StatusInternalServerError, fmt.Errorf("error sending command: received status code: 500 Internal Server Error")},
		{"Bad request case", http.StatusBadRequest, fmt.Errorf("error sending command: received status code: 400 Bad Request")},
	}

	event := models.Event{
		DeviceName: InferenceMQTTDevice,
		Readings: []models.Reading{
			models.SimpleReading{
				BaseReading: models.BaseReading{
					DeviceName: "test-reading",
				},
				Value: `[{"SKU": "HXI86WHU", "delta": -2}]`,
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.TestCaseName, func(t *testing.T) {

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				output := Ledger{
					IsPaid:        false,
					LineItems:     []LineItem{},
					TransactionID: 123,
					LineTotal:     20.5,
				}

				outputJSON, _ := utilities.GetAsJSON(output)

				writeError := false
				writeContentType := "json"
				if tc.Expected != nil {
					writeError = true
					writeContentType = "string"
				}

				httpResponse := utilities.HTTPResponse{
					Content:     outputJSON,
					ContentType: writeContentType,
					StatusCode:  tc.statusCode,
					Error:       writeError,
				}

				httpResponse.WriteHTTPResponse(w, r)
			}))

			// VendingState initialization
			inferenceStopChannel := make(chan int)
			stopChannel := make(chan int)

			vendingState := VendingState{
				InferenceWaitThreadStopChannel: inferenceStopChannel,
				ThreadStopChannel:              stopChannel,
				CurrentUserData:                OutputData{RoleID: 1},
				Configuration: &ServiceConfiguration{
					InventoryService:                  testServer.URL,
					InventoryAuditLogService:          testServer.URL,
					DeviceControllerBoarddisplayReset: testServer.URL,
					DeviceControllerBoarddisplayRow1:  testServer.URL,
					LedgerService:                     testServer.URL,
				},
			}

			_, err := vendingState.HandleMqttDeviceReading(logger.NewMockClient(), event)

			e, ok := err.(error)
			if ok {
				assert.Equal(t, tc.Expected, e)
			}

		})
	}
}

func TestVerifyDoorAccess(t *testing.T) {
	testCases := []struct {
		TestCaseName    string
		StatusCode      int
		MaintenanceMode bool
		RoleID          int
	}{
		{"Successful case", http.StatusOK, false, 1},
		{"MaintanceMode on", http.StatusOK, true, 1},
		{"Role 3", http.StatusOK, false, 3},
	}

	// VendingState initialization
	inferenceStopChannel := make(chan int)
	stopChannel := make(chan int)

	event := models.Event{
		DeviceName: "ds-card-reader",
		Readings: []models.Reading{
			models.SimpleReading{
				BaseReading: models.BaseReading{
					DeviceName: "ds-card-reader",
				},
				Value: `[{"SKU": "HXI86WHU", "delta": -2}]`,
			},
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(nil)
		assert.NoError(t, err)
	}))

	for _, tc := range testCases {

		t.Run(tc.TestCaseName, func(t *testing.T) {

			authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				output := OutputData{
					RoleID: tc.RoleID,
				}
				authDataJSON, _ := utilities.GetAsJSON(output)
				utilities.WriteJSONHTTPResponse(w, r, http.StatusOK, authDataJSON, false)
			}))

			vendingState := VendingState{
				InferenceWaitThreadStopChannel: inferenceStopChannel,
				ThreadStopChannel:              stopChannel,
				CurrentUserData:                OutputData{RoleID: 1},
				CVWorkflowStarted:              false,
				MaintenanceMode:                tc.MaintenanceMode,
				Configuration: &ServiceConfiguration{
					InferenceHeartbeat:               testServer.URL,
					DeviceControllerBoarddisplayRow1: testServer.URL,
					DeviceControllerBoarddisplayRow2: testServer.URL,
					DeviceControllerBoarddisplayRow3: testServer.URL,
					DeviceControllerBoardLock1:       testServer.URL,
					AuthenticationEndpoint:           authServer.URL,
				},
			}

			_, err := vendingState.VerifyDoorAccess(logger.NewMockClient(), event)

			e, ok := err.(error)
			if ok {
				assert.NoError(t, e)
			}
		})
	}
}
