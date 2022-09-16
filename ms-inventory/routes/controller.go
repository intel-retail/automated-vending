// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type Controller struct {
	lc             logger.LoggingClient
	service        interfaces.ApplicationService
	inventoryItems Products
	auditLog       AuditLog
}

func NewController(lc logger.LoggingClient, service interfaces.ApplicationService) Controller {
	return Controller{
		lc:      lc,
		service: service,
	}
}

func (c *Controller) AddAllRoutes() error {
	var err error

	err = c.service.AddRoute("/inventory", c.InventoryGet, "GET")
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/inventory", c.InventoryPost, "POST", "OPTIONS")
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/inventory/delta", c.DeltaInventorySKUPost, "POST", "OPTIONS")
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/inventory/{sku}", c.InventoryItemGet, "GET")
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/inventory/{sku}", c.InventoryDelete, "DELETE", "OPTIONS")
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/auditlog", c.AuditLogGetAll, "GET", "OPTIONS")
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/auditlog", c.AuditLogPost, "POST")
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/auditlog/{entry}", c.AuditLogGetEntry, "GET", "OPTIONS")
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/auditlog/{entry}", c.AuditLogDelete, "DELETE")
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	return nil

}

func (c *Controller) errorAddRouteHandler(err error) error {
	if err != nil {
		c.lc.Errorf("error adding route: %s", err.Error())
		return fmt.Errorf("error adding route: %s", err.Error())
	}
	return nil
}
