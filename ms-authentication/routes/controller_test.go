// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
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
			mockAppService.On("LoggingClient").Return(logger.NewMockClient())
			if !tt.failAddRoute {
				mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			} else {
				mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("fail"))
			}

			c := NewController(mockAppService)

			err := c.AddAllRoutes()

			if tt.failAddRoute {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
