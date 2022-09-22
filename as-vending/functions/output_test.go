// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"as-vending/config"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	client_mocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	edgexError "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	err := m.Run()
	os.Exit(err)
}

func TestCheckInferenceStatus(t *testing.T) {
	testCases := []struct {
		TestCaseName    string
		statusCode      int
		Expected        bool
		GetCommandError edgexError.EdgeX
	}{
		{"Successful case", http.StatusOK, true, nil},
		{"Negative case", http.StatusInternalServerError, false, edgexError.NewCommonEdgeXWrapper(errors.New("failed"))},
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

			mockCommandClient := &client_mocks.CommandClient{}
			resp := responses.NewEventResponse("", "", tc.statusCode, dtos.Event{})
			mockCommandClient.On("IssueGetCommandByName", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&resp, tc.GetCommandError)

			vendingstate := VendingState{
				CommandClient: mockCommandClient,
			}
			assert.Equal(t, tc.Expected, vendingstate.checkInferenceStatus(logger.NewMockClient(), testServer.URL, "test-device"), "Expected value to match output")
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

func TestDisplayLedger(t *testing.T) {

	mockCommandClient := &client_mocks.CommandClient{}
	resp := common.BaseResponse{
		StatusCode: http.StatusOK,
	}

	mockCommandClient.On("IssueSetCommandByName", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	vendingState := VendingState{
		Configuration: &config.AppCustomConfig{
			LCDRowLength:                   20,
			ControllerBoardDisplayResetCmd: "displayReset",
			ControllerBoardDisplayRow1Cmd:  "diplayrow1",
		},
		CommandClient: mockCommandClient,
	}

	ledger := Ledger{
		LineItems: []LineItem{{ProductName: "itemX", ItemCount: 2, ItemPrice: 1.50, SKU: "1234"}},
	}
	err := vendingState.displayLedger(logger.NewMockClient(), "test-device", ledger)
	assert.NoError(t, err)
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

	event := dtos.Event{
		DeviceName: InferenceMQTTDevice,
		Readings: []dtos.BaseReading{
			{
				DeviceName: "inferenceSkuDelta",
				SimpleReading: dtos.SimpleReading{
					Value: `[{"SKU": "HXI86WHU", "delta": -2}]`,
				},
			},
		},
	}

	mockCommandClient := &client_mocks.CommandClient{}
	resp := common.BaseResponse{
		StatusCode: http.StatusOK,
	}

	mockCommandClient.On("IssueSetCommandByName", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

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
				Configuration: &config.AppCustomConfig{
					InventoryService:               testServer.URL,
					InventoryAuditLogService:       testServer.URL,
					ControllerBoardDisplayResetCmd: "displayreset",
					ControllerBoardDisplayRow1Cmd:  "displayrow1",
					LedgerService:                  testServer.URL,
				},

				CommandClient: mockCommandClient,
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

	mockCommandClient := &client_mocks.CommandClient{}
	eventResp := responses.NewEventResponse("", "", http.StatusOK, dtos.Event{})
	resp := common.BaseResponse{
		StatusCode: http.StatusOK,
	}

	mockCommandClient.On("IssueSetCommandByName", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)
	mockCommandClient.On("IssueGetCommandByName", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&eventResp, nil)

	event := dtos.Event{
		DeviceName: "ds-card-reader",
		Readings: []dtos.BaseReading{
			{
				DeviceName: "ds-card-reader",
				SimpleReading: dtos.SimpleReading{
					Value: `[{"SKU": "HXI86WHU", "delta": -2}]`,
				},
			},
		},
	}

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
				Configuration: &config.AppCustomConfig{
					InferenceHeartbeatCmd:         "inferenceHeartbeat",
					ControllerBoardDisplayRow1Cmd: "displayrow1",
					ControllerBoardDisplayRow2Cmd: "displayrow2",
					ControllerBoardDisplayRow3Cmd: "displayrow3",
					ControllerBoardLock1Cmd:       "lock1",
					AuthenticationEndpoint:        authServer.URL,
				},

				CommandClient: mockCommandClient,
			}

			_, err := vendingState.VerifyDoorAccess(logger.NewMockClient(), event)

			e, ok := err.(error)
			if ok {
				assert.NoError(t, e)
			}
		})
	}
}
