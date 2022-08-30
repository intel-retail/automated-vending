// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"ms-ledger/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type Route struct {
	lc            logger.LoggingClient
	serviceConfig *config.ServiceConfig
}

func NewRoute(lc logger.LoggingClient, serviceConfig *config.ServiceConfig) Route {
	return Route{
		lc:            lc,
		serviceConfig: serviceConfig,
	}
}
