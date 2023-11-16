// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"fmt"
	"net/http"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

type Controller struct {
	lc                logger.LoggingClient
	service           interfaces.ApplicationService
	inventoryItems    Products
	auditLog          AuditLog
	auditLogFileName  string
	inventoryFileName string
}

func NewController(lc logger.LoggingClient, service interfaces.ApplicationService, auditLogFileName string, inventoryFileName string) Controller {
	return Controller{
		lc:                lc,
		service:           service,
		inventoryFileName: inventoryFileName,
		auditLogFileName:  auditLogFileName,
	}
}

func (c *Controller) AddAllRoutes() error {
	var err error

	err = c.service.AddRoute("/inventory", c.InventoryGet, http.MethodGet)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/inventory", c.InventoryPost, http.MethodPost)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/inventory/delta", c.DeltaInventorySKUPost, http.MethodPost)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/inventory/{sku}", c.InventoryItemGet, http.MethodGet)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/inventory/{sku}", c.InventoryDelete, http.MethodDelete)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/auditlog", c.AuditLogGetAll, http.MethodGet)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/auditlog", c.AuditLogPost, http.MethodPost)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/auditlog/{entry}", c.AuditLogGetEntry, http.MethodGet)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/auditlog/{entry}", c.AuditLogDelete, http.MethodDelete)
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
