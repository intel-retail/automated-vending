// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"ds-controller-board/driver"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/startup"
)

const (
	version     string = "1.0"
	serviceName string = "ds-controller-board"
)

func main() {
	d := driver.NewControllerBoardDeviceDriver()
	startup.Bootstrap(serviceName, version, d)
}
