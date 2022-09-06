// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package common

// CardReaderConfig holds the configurable options for a automated checkout
// card reader device
type CardReaderConfig struct {
	DeviceName       string
	DeviceSearchPath string
	VID              uint16
	PID              uint16
	SimulateDevice   bool
}

// CommandCardReaderStatus is the string value of the command that will be
// sent to the card reader driver that performs a health check of the underlying
// device
const (
	CommandCardReaderStatus = "status"
	CommandCardNumber       = "card-number"
)
