// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
    "time"
)

// Config is the global device configuration, which is populated by values in
// the "Driver" section of res/configuration.toml
type Config struct {
    VirtualControllerBoard bool
    PID                    string
    VID                    string
    DisplayTimeout         time.Duration
    LockTimeout            time.Duration
}
