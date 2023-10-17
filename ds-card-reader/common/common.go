// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package common

// CommandCardReaderStatus is the string value of the command that will be
// sent to the card reader driver that performs a health check of the underlying
// device
const (
	CommandCardReaderStatus = "status"
	CommandCardNumber       = "card-number"
)
