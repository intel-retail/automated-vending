// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"as-controller-board-status/config"
	"as-controller-board-status/functions"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestController_AddAllRoutes(t *testing.T) {

	tests := []struct {
		name         string
		badservice   bool
		want         int
		failAddRoute bool
	}{
		{
			name:         "valid case",
			failAddRoute: false,
		},
		{
			name:         "invalid case",
			failAddRoute: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAppService := &mocks.ApplicationService{}
			if !tt.failAddRoute {
				mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			} else {
				mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("fail"))
			}

			c := &Controller{
				lc:      logger.NewMockClient(),
				service: mockAppService,
			}

			err := c.AddAllRoutes()

			if tt.failAddRoute {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestController_GetStatus(t *testing.T) {

	boardStatus := functions.CheckBoardStatus{
		DoorClosed:    true, // Set default door state to closed
		Configuration: &config.ControllerBoardStatusConfig{},
	}

	tests := []struct {
		name                  string
		controllerBoardStatus functions.ControllerBoardStatus
		expectedContent       string
		expectedStatus        int
		RESTURL               string
	}{
		// TODO: Add test cases.
		{
			name:                  "Success",
			controllerBoardStatus: functions.ControllerBoardStatus{},
			expectedStatus:        http.StatusOK,
			RESTURL:               "/status",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.RESTURL, nil)
			w := httptest.NewRecorder()
			c := &Controller{
				lc:          logger.NewMockClient(),
				service:     &mocks.ApplicationService{},
				boardStatus: &boardStatus,
			}

			c.boardStatus.ControllerBoardStatus = &tt.controllerBoardStatus

			c.GetStatus(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)
			responseContent := utilities.HTTPResponse{}
			err = json.Unmarshal(body, &responseContent)

			assert.NoError(t, err)
		})
	}
}
