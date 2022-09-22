// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"fmt"
	"net/http"

	"as-controller-board-status/functions"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

type Controller struct {
	lc          logger.LoggingClient
	service     interfaces.ApplicationService
	boardStatus *functions.CheckBoardStatus
}

func NewController(lc logger.LoggingClient, service interfaces.ApplicationService, boardStatus *functions.CheckBoardStatus) Controller {
	return Controller{
		lc:          lc,
		service:     service,
		boardStatus: boardStatus,
	}
}

func (c *Controller) AddAllRoutes() error {
	// Add the "status" REST API route
	err := c.service.AddRoute("/status", c.GetStatus, http.MethodGet, http.MethodOptions)
	if err != nil {
		return fmt.Errorf("error adding route: %s", err.Error())
	}
	return nil
}

// GetStatus is a REST API endpoint that enables a web UI or some other downstream
// service to inquire about the status of the upstream Automated Checkout hardware interface(s).
func (c *Controller) GetStatus(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		controllerBoardStatusJSON, err := utilities.GetAsJSON(c.boardStatus.GetControllerBoardStatus())
		if err != nil {
			errMsg := "Failed to serialize the controller board's current state"
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
			c.lc.Errorf("%s: %s", errMsg, err.Error())
			return
		}

		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, controllerBoardStatusJSON, false)
		c.lc.Info("GetStatus successfully!")
	})
}
