// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"os"

	"ms-authentication/routes"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
)

const (
	serviceKey = "ms-authentication"
)

func main() {
	// TODO: See https://docs.edgexfoundry.org/2.2/microservices/application/ApplicationServices/
	//       for documentation on application services.
	var ok bool
	service, ok := pkg.NewAppService(serviceKey)
	if !ok {
		os.Exit(1)
	}
	lc := service.LoggingClient()

	if err := service.AddRoute("/authentication/{cardid}", routes.AuthenticationGet, "GET"); err != nil {
		lc.Errorf("Unable to add /authentication/{cardid} GET route: %s", err.Error())
		os.Exit(1)
	}

	if err := service.MakeItRun(); err != nil {
		lc.Errorf("MakeItRun returned error: %s", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
