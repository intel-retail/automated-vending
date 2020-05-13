// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"fmt"
	"strconv"
	"strings"
)

// Command is an enumeration of the commands accepted through the
// controller board's serial interface
var Command = struct {
	Lock1       string
	Lock2       string
	UnLock1     string
	UnLock2     string
	Status      string
	GetStatus   string
	DefaultDisp string
	Message0    string
	Message1    string
	Message2    string
	Message3    string
}{
	Lock1:       "L1\n",
	Lock2:       "L2\n",
	UnLock1:     "U1\n",
	UnLock2:     "U2\n",
	Status:      "Status\n",
	GetStatus:   "S\n",
	DefaultDisp: "D\n",
	Message0:    "M0",
	Message1:    "M1",
	Message2:    "M2",
	Message3:    "M3",
}

// StatusEvent is a struct to handle the mapping of status event values to their
// respective JSON values in a JSON object
type StatusEvent struct {
	Lock1Status int     `json:"lock1_status"`
	Lock2Status int     `json:"lock2_status"`
	DoorClosed  bool    `json:"door_closed"` // true means the door is closed and false means the door is open
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

// ParseStatus is a function used to marshal the serial output from the
// controller board into a StatusEvent struct
func ParseStatus(status string) (StatusEvent, error) {
	var sEvent StatusEvent // STATUS,L1,0,L2,0,D,0,T,78.58,H,19.54
	var err error
	s1 := strings.Split(status, ",")

	if len(s1) == 11 {
		sEvent.Lock1Status, err = strconv.Atoi(s1[2])
		if err != nil {
			return sEvent, fmt.Errorf("unable to parse Lock1 status: %v", err)
		}

		sEvent.Lock2Status, err = strconv.Atoi(s1[4])
		if err != nil {
			return sEvent, fmt.Errorf("unable to parse Lock2 status: %v", err)
		}

		sEvent.DoorClosed, err = strconv.ParseBool(s1[6])
		if err != nil {
			return sEvent, fmt.Errorf("unable to parse DoorClosed status: %v", err)
		}

		var temp float64
		temp, err = strconv.ParseFloat(s1[8], 64)
		sEvent.Temperature = temp
		if err != nil {
			return sEvent, fmt.Errorf("unable to parse Temperature status: %v", err)
		}

		// Remove "\r\n" from end of the line before calling strconv.ParseFloat
		s1[10] = strings.TrimSuffix(s1[10], "\r\n")
		var hum float64
		hum, err = strconv.ParseFloat(s1[10], 64)
		sEvent.Humidity = hum
		if err != nil {
			return sEvent, fmt.Errorf("unable to parse Humidity status: %v", err)
		}
	}

	return sEvent, nil
}
