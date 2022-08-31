// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/mock"
)

func TestController_AddAllRoutes(t *testing.T) {

	inventoryServer := newInventoryTestServer(t)

	tests := []struct {
		name       string
		badservice bool
		want       int
	}{
		// TODO: Add test cases.
		{
			name:       "valid case",
			badservice: false,
			want:       0,
		},
		{
			name:       "invalid case",
			badservice: true,
			want:       1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAppService := &mocks.ApplicationService{}
			if !tt.badservice {
				mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			} else {
				mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("fail"))
			}

			c := &Controller{
				lc:                logger.NewMockClient(),
				service:           mockAppService,
				inventoryEndpoint: inventoryServer.URL,
			}

			if got := c.AddAllRoutes(); got != tt.want {
				t.Errorf("Controller.AddAllRoutes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_errorAddRouteHandler(t *testing.T) {

	type args struct {
		lc  logger.LoggingClient
		err error
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
		{
			name: "valid case",
			args: args{
				lc:  logger.NewMockClient(),
				err: nil,
			},
			want: 0,
		},
		{
			name: "error case",
			args: args{
				lc:  logger.NewMockClient(),
				err: fmt.Errorf("fail"),
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errorAddRouteHandler(tt.args.lc, tt.args.err); got != tt.want {
				t.Errorf("errorAddRouteHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
