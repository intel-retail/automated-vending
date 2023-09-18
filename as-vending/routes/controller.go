// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"as-vending/functions"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

type Controller struct {
	lc           logger.LoggingClient
	service      interfaces.ApplicationService
	vendingState *functions.VendingState
}

func NewController(lc logger.LoggingClient, service interfaces.ApplicationService, vendingState *functions.VendingState) Controller {
	return Controller{
		lc:           lc,
		service:      service,
		vendingState: vendingState,
	}
}

func (c *Controller) AddAllRoutes() error {
	var err error

	err = c.service.AddRoute("/boardStatus", c.BoardStatus, http.MethodPost)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/resetDoorLock", c.ResetDoorLock, http.MethodPost)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/maintenanceMode", c.GetMaintenanceMode, http.MethodGet)
	if errWithMsg := c.errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	return nil

}

// GetMaintenanceMode will return a JSON response containing the boolean state
// of the vendingState's maintenance mode.
func (c *Controller) GetMaintenanceMode(writer http.ResponseWriter, req *http.Request) {
	mm, _ := utilities.GetAsJSON(functions.MaintenanceMode{MaintenanceMode: c.vendingState.MaintenanceMode})
	utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, mm, false)
}

func (c *Controller) errorAddRouteHandler(err error) error {
	errorMsg := "error adding route: %s"
	if err != nil {
		c.lc.Errorf(errorMsg, err.Error())
		return fmt.Errorf(errorMsg, err.Error())
	}
	return nil
}

// ResetDoorLock endpoint to reset all door lock states
func (c *Controller) ResetDoorLock(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	// Check the HTTP Request's form values
	returnval := "reset the door lock"

	close(c.vendingState.ThreadStopChannel)
	c.vendingState.ThreadStopChannel = make(chan int)

	c.vendingState.MaintenanceMode = false
	c.vendingState.CVWorkflowStarted = false
	c.vendingState.DoorClosed = true
	c.vendingState.DoorClosedDuringCVWorkflow = false
	c.vendingState.DoorOpenedDuringCVWorkflow = false
	c.vendingState.InferenceDataReceived = false

	c.lc.Infof("Maintenance card scanned")
	c.lc.Debugf("workflow: %t", c.vendingState.CVWorkflowStarted)
	c.lc.Debugf("maintenance mode: %t", c.vendingState.MaintenanceMode)
	c.lc.Debugf("open: %t", c.vendingState.DoorOpenedDuringCVWorkflow)
	c.lc.Debugf("closed: %t", c.vendingState.DoorClosedDuringCVWorkflow)
	c.lc.Debugf("Inference: %t", c.vendingState.InferenceDataReceived)
	c.lc.Debugf("door: %t", c.vendingState.DoorClosed)

	// Write the HTTP status header
	writer.WriteHeader(http.StatusOK)

	_, writeErr := writer.Write([]byte(returnval))
	if writeErr != nil {
		c.lc.Errorf("Failed to write item data back to caller")
	}

}

// BoardStatus endpoint that handles board status events from the controller board status application service
func (c *Controller) BoardStatus(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	var status int

	// Read request body
	body := make([]byte, req.ContentLength)
	_, err := io.ReadFull(req.Body, body)
	if err != nil {
		c.lc.Errorf("Failed to read request data")
	}

	// Unmarshal the string contents of request into a proper structure
	var boardStatus functions.ControllerBoardStatus
	if err := json.Unmarshal(body, &boardStatus); err != nil {
		c.lc.Errorf("Failed to read request data")
	}
	returnval := "Board status received but maintenance mode was not set"
	status = http.StatusOK

	// Check controller board MinTemperatureStatus state. If it's true then a minimum temperature event has happened
	if boardStatus.MinTemperatureStatus {
		returnval = string("Temperature status received and maintenance mode was set")
		status = http.StatusOK
		c.lc.Error("Cooler temperature exceeds the minimum temperature threshold. The cooler needs maintenance.")
		c.vendingState.MaintenanceMode = true
	}
	// Check controller board MaxTemperatureStatus state. If it's true then a maximum temperature event has happened
	if boardStatus.MaxTemperatureStatus {
		returnval = string("Temperature status received and maintenance mode was set")
		status = http.StatusOK
		c.lc.Error("Cooler temperature exceeds the maximum temperature threshold. The cooler needs maintenance.")
		c.vendingState.MaintenanceMode = true
	}

	// Check to see if the board closed state is different than the previous state. If it is we need to update the state and
	// set the related properties.
	if c.vendingState.DoorClosed != boardStatus.DoorClosed {
		c.lc.Errorf("Successfully updated the door event. Door closed: %v", boardStatus.DoorClosed)
		returnval = string("Door closed change event was received ")
		status = http.StatusOK //FIXME: This is an issue
		c.vendingState.DoorClosed = boardStatus.DoorClosed
		if c.vendingState.CVWorkflowStarted {
			// If the door was opened then we want to wait for the door closed event
			if !boardStatus.DoorClosed {
				c.vendingState.DoorOpenedDuringCVWorkflow = true
				// Stop the open wait thread since the door is now opened
				close(c.vendingState.DoorOpenWaitThreadStopChannel)
				c.vendingState.DoorOpenWaitThreadStopChannel = make(chan int)

				// Wait for door closed event. If the door isn't closed within the timeout
				// then leave the workflow, remove the user data, and enter maintenance mode
				go func() {
					c.lc.Infof("Door Opened: wait for %v seconds", c.vendingState.DoorCloseStateTimeout)
					for {
						select {
						case <-time.After(c.vendingState.DoorCloseStateTimeout):
							{
								if !c.vendingState.DoorClosedDuringCVWorkflow {
									c.lc.Error("Door Opened: Failed")
									c.vendingState.CVWorkflowStarted = false
									c.vendingState.CurrentUserData = functions.OutputData{}
									c.vendingState.MaintenanceMode = true
								}
								return
							}
						case <-c.vendingState.DoorCloseWaitThreadStopChannel:
							c.lc.Info("Stopped the door closed wait thread")
							return

						case <-c.vendingState.ThreadStopChannel:
							c.lc.Info("Globally stopped the door closed wait thread")
							return
						}
					}
				}()
			}
			// If the door was closed we want to wait for the inference event
			if boardStatus.DoorClosed {
				c.vendingState.DoorClosedDuringCVWorkflow = true
				// Stop the open wait thread since the door is now opened
				close(c.vendingState.DoorCloseWaitThreadStopChannel)
				c.vendingState.DoorCloseWaitThreadStopChannel = make(chan int)

				// Wait for the inference data to be received. If we don't receive any inference data with the timeout
				// then leave the workflow, remove the user data, and enter maintenance mode
				go func() {
					c.lc.Infof("Door Closed: wait for %v seconds", c.vendingState.InferenceTimeout)
					for {
						select {
						case <-time.After(c.vendingState.InferenceTimeout):
							{
								if !c.vendingState.InferenceDataReceived {
									c.lc.Error("Door Closed: Failed")
									c.vendingState.CVWorkflowStarted = false
									c.vendingState.CurrentUserData = functions.OutputData{}
									c.vendingState.MaintenanceMode = true
								}
								return
							}
						case <-c.vendingState.InferenceWaitThreadStopChannel:
							c.lc.Info("Stopped the inference wait thread")
							return

						case <-c.vendingState.ThreadStopChannel:
							c.lc.Info("Globally stopped the inference wait thread")
							return
						}
					}
				}()
			}
		}
	}

	// Write the HTTP status header
	writer.WriteHeader(status)

	_, writeErr := writer.Write([]byte(returnval))
	if writeErr != nil {
		c.lc.Error("Failed to write item data back to caller")
	}
}
