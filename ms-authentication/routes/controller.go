// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

type Controller struct {
	service interfaces.ApplicationService
	lc      logger.LoggingClient
}

func NewController(service interfaces.ApplicationService) Controller {
	return Controller{
		service: service,
		lc:      service.LoggingClient(),
	}
}

func (c *Controller) AddAllRoutes() error {
	err := c.service.AddRoute("/authentication/{cardid}", c.AuthenticationGet, "GET")
	if errWithMsg := errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}
	return nil
}
func errorAddRouteHandler(err error) error {
	if err != nil {
		return fmt.Errorf("error adding route: %s", err.Error())
	}
	return nil
}
