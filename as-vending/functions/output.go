// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

const (
	InferenceMQTTDevice = "Inference-MQTT-device"
	DsCardReader        = "ds-card-reader"
)

// DeviceHelper is an EdgeX function that is passed into the EdgeX SDK's function pipeline.
// It is a decision function that allows for multiple devices to have their events processed
// correctly by this application service.
func (vendingState *VendingState) DeviceHelper(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		// We didn't receive a result
		return false, nil
	}

	event := data.(dtos.Event)

	switch event.DeviceName {
	case DsCardReader:
		{
			return vendingState.VerifyDoorAccess(ctx.LoggingClient(), event)
		}
	case InferenceMQTTDevice:
		{
			return vendingState.HandleMqttDeviceReading(ctx.LoggingClient(), event)
		}
	default:
		{
			return false, nil
		}
	}
}

// HandleMqttDeviceReading is an EdgeX function that simply handles events coming from
// the MQTT device service.
func (vendingState *VendingState) HandleMqttDeviceReading(lc logger.LoggingClient, event dtos.Event) (bool, interface{}) {
	if event.DeviceName == InferenceMQTTDevice {

		lc.Debugf("Inference mqtt device")
		lc.Debugf("workflow: +%v", vendingState.CVWorkflowStarted)
		lc.Debugf("maintenance mode: +%v", vendingState.MaintenanceMode)
		lc.Debugf("open: +%v", vendingState.DoorOpenedDuringCVWorkflow)
		lc.Debugf("closed: +%v", vendingState.DoorClosedDuringCVWorkflow)
		lc.Debugf("Inference: +%v ", vendingState.InferenceDataReceived)
		lc.Debugf("door: +%v", vendingState.DoorClosed)

		lc.Debug("Processing reading from MQTT device service")
		for _, eventReading := range event.Readings {
			if len(eventReading.Value) < 1 {
				return false, fmt.Errorf("event reading was empty")
			}
			switch eventReading.DeviceName {
			case "inferenceSkuDelta":
				{
					fmt.Println("Inference Started")
					var skuDelta []deltaSKU

					if err := json.Unmarshal([]byte(eventReading.Value), &skuDelta); err != nil {
						lc.Errorf("HandleMqttDeviceReading failed to unmarshal skuDelta message for %s: %v", eventReading.Value, err)
						fmt.Println("Inference Failed")
						return false, err
					}

					// do some things with the skuDelta
					// example:
					// [{"SKU": "HXI86WHU", "delta": -2}]
					deltaLedger := deltaLedger{
						AccountID: vendingState.CurrentUserData.AccountID,
						DeltaSKUs: skuDelta,
					}

					vendingState.InferenceDataReceived = true
					// Stop the open wait thread since the door is now opened
					close(vendingState.InferenceWaitThreadStopChannel)
					vendingState.InferenceWaitThreadStopChannel = make(chan int)

					if vendingState.CurrentUserData.RoleID == 1 {
						// POST the deltaLedger json string to the ledger endpoint
						outputBytes, err := json.Marshal(deltaLedger)
						if err != nil {
							lc.Errorf("HandleMqttDeviceReading failed to marshal deltaLedger: %v", err)
							return false, err
						}

						lc.Info("Sending SKU delta to ledger service")
						// send SKU delta to ledger service and get back current ledger information
						resp, err := sendHttpRequest(lc, http.MethodPost, vendingState.Configuration.LedgerService, outputBytes)
						if err != nil {
							lc.Errorf("Ledger service failed: %s", err.Error())
							return false, err
						}

						defer resp.Body.Close()

						lc.Info("Successfully updated the user's ledger")

						var currentLedger Ledger
						_, err = utilities.ParseJSONHTTPResponseContent(resp.Body, &currentLedger)
						if err != nil {
							return false, fmt.Errorf("Unable to unmarshal ledger response")
						}

						// Display Ledger Total on LCD
						if displayErr := vendingState.displayLedger(lc, eventReading.DeviceName, currentLedger); displayErr != nil {
							return false, displayErr
						}

					}

					// POST the deltaLedger json string to the inventory endpoint
					outputBytes, err := json.Marshal(deltaLedger.DeltaSKUs)
					if err != nil {
						return false, fmt.Errorf("HandleMqttDeviceReading failed to marshal deltaLedger.DeltaSKUs")
					}

					lc.Info("Sending SKU delta to inventory service")
					inventoryResp, err := sendHttpRequest(lc, http.MethodPost, vendingState.Configuration.InventoryService, outputBytes)
					if err != nil {
						return false, err
					}
					defer inventoryResp.Body.Close()
					// Post an audit log entry for this transaction, regardless of ledger or not
					auditLogEntry := AuditLogEntry{
						AccountID:      vendingState.CurrentUserData.AccountID,
						CardID:         vendingState.CurrentUserData.CardID,
						RoleID:         vendingState.CurrentUserData.RoleID,
						PersonID:       vendingState.CurrentUserData.PersonID,
						InventoryDelta: deltaLedger.DeltaSKUs,
						CreatedAt:      time.Now().UnixNano(),
					}

					outputBytes, err = json.Marshal(auditLogEntry)
					if err != nil {
						return false, err
					}

					lc.Info("Sending audit log entry to inventory service")
					auditResp, err := sendHttpRequest(lc, http.MethodPost, vendingState.Configuration.InventoryAuditLogService, outputBytes)
					if err != nil {
						return false, err
					}
					defer auditResp.Body.Close()
					vendingState.CurrentUserData = OutputData{}
					vendingState.CVWorkflowStarted = false
					lc.Info("Inference complete and workflow status reset")
					// Close all thread to ensure all threads are cleaned up before the next card is scanned.
					close(vendingState.ThreadStopChannel)
					vendingState.ThreadStopChannel = make(chan int)
				}
			default:
				{
					lc.Info("Received an event with an unknown name")
					return false, nil
				}
			}
		}
	}

	return false, nil
}

// VerifyDoorAccess will take the card reader events and verify the read card id against the white list
// If the card is valid the function will send the unlock message to the device-controller-board device service
func (vendingState *VendingState) VerifyDoorAccess(lc logger.LoggingClient, event dtos.Event) (bool, interface{}) {

	lc.Infof("new card scanned")
	lc.Infof("workflow: +%v", vendingState.CVWorkflowStarted)
	lc.Infof("maintenance mode: +%v", vendingState.MaintenanceMode)
	lc.Infof("open: +%v", vendingState.DoorOpenedDuringCVWorkflow)
	lc.Infof("closed: +%v", vendingState.DoorClosedDuringCVWorkflow)
	lc.Infof("Inference: +%v ", vendingState.InferenceDataReceived)
	lc.Infof("door: +%v", vendingState.DoorClosed)

	if event.DeviceName == DsCardReader && !vendingState.CVWorkflowStarted {
		lc.Info("Verify the card reader input against the white list")

		lc.Infof("Card Scanned")
		lc.Infof("workflow: +%v", vendingState.CVWorkflowStarted)
		lc.Infof("maintenance mode: +%v", vendingState.MaintenanceMode)
		lc.Infof("open: +%v", vendingState.DoorOpenedDuringCVWorkflow)
		lc.Infof("closed: +%v", vendingState.DoorClosedDuringCVWorkflow)
		lc.Infof("Inference: +%v ", vendingState.InferenceDataReceived)
		lc.Infof("door: +%v", vendingState.DoorClosed)

		// check to see if inference is running and set maintenance mode accordingly
		if !vendingState.MaintenanceMode {
			vendingState.MaintenanceMode = !vendingState.checkInferenceStatus(lc, vendingState.Configuration.InferenceHeartbeat, event.DeviceName)
		}

		for _, eventReading := range event.Readings {
			if len(eventReading.Value) < 1 {
				return false, fmt.Errorf("event reading was empty, devicename: %s, resourcename: %s", eventReading.DeviceName, eventReading.ResourceName)
			}

			// Retrieve & Hit auth endpoint
			vendingState.getCardAuthInfo(lc, vendingState.Configuration.AuthenticationEndpoint, eventReading.Value)

			switch vendingState.CurrentUserData.RoleID {
			// Check the role of the card scanned. Role 1 = customer and Role 2 = item stocker
			case 1, 2:
				{
					if !vendingState.MaintenanceMode {
						lc.Infof("%s readable value from %s is %s", eventReading.ResourceName, eventReading.DeviceName, eventReading.Value)
						// display "hello" on row 2
						settings := make(map[string]string)
						settings["displayRow2"] = "hello"
						err := vendingState.SendCommand(lc, http.MethodPut, eventReading.DeviceName, vendingState.Configuration.DeviceControllerBoarddisplayRow2, settings)
						if err != nil {
							return false, err
						}

						settings = make(map[string]string)
						settings["displayRow3"] = eventReading.Value
						// display the card number on row 3
						err = vendingState.SendCommand(lc, http.MethodPut, eventReading.DeviceName, vendingState.Configuration.DeviceControllerBoarddisplayRow3, settings)
						if err != nil {
							return false, err
						}

						settings = make(map[string]string)
						settings["lock1"] = "true"
						// unlock
						err = vendingState.SendCommand(lc, http.MethodPut, eventReading.DeviceName, vendingState.Configuration.DeviceControllerBoardLock1, settings)
						if err != nil {
							return false, err
						}

						// Start the workflow state and set all of the thread states to false
						vendingState.CVWorkflowStarted = true
						vendingState.DoorClosedDuringCVWorkflow = false
						vendingState.DoorOpenedDuringCVWorkflow = false
						vendingState.InferenceDataReceived = false

						// Wait for the door open event to be received. If we don't receive the door open event within the timeout
						// then leave the workflow state and remove all user data
						go func() {
							for {
								select {
								case <-time.After(vendingState.Configuration.DoorOpenStateTimeout):
									if !vendingState.DoorOpenedDuringCVWorkflow {
										lc.Info("door wasn't opened so we reset")
										vendingState.CVWorkflowStarted = false
										vendingState.CurrentUserData = OutputData{}
									}

									lc.Infof("Card Scan")
									lc.Infof("workflow: +%v", vendingState.CVWorkflowStarted)
									lc.Infof("maintenance mode: +%v", vendingState.MaintenanceMode)
									lc.Infof("open: +%v", vendingState.DoorOpenedDuringCVWorkflow)
									lc.Infof("closed: +%v", vendingState.DoorClosedDuringCVWorkflow)
									lc.Infof("Inference: +%v ", vendingState.InferenceDataReceived)
									lc.Infof("door: +%v", vendingState.DoorClosed)
									return

								case <-vendingState.DoorOpenWaitThreadStopChannel:
									lc.Info("Stopped the door open wait thread")
									return

								case <-vendingState.ThreadStopChannel:
									lc.Info("Globally stopped the door open wait thread")
									return
								}
							}
						}()
					} else {
						settings := make(map[string]string)
						settings["displayRow1"] = "Out of Order"
						// display out of order when door waiting state is set to false
						err := vendingState.SendCommand(lc, http.MethodPut, eventReading.DeviceName, vendingState.Configuration.DeviceControllerBoarddisplayRow1, settings)
						if err != nil {
							return false, err
						}
					}

				}
			// Check the role of the card scanned. Role 3 = maintainer
			case 3:
				{
					close(vendingState.ThreadStopChannel)
					vendingState.ThreadStopChannel = make(chan int)

					lc.Infof("%s readable value from %s is %s", eventReading.ResourceName, eventReading.DeviceName, eventReading.Value)

					// display text "Maintenance Mode" in row 2
					settings := make(map[string]string)
					settings["displayRow2"] = "Maintenance Mode"
					err := vendingState.SendCommand(lc, http.MethodPut, eventReading.DeviceName, vendingState.Configuration.DeviceControllerBoarddisplayRow2, settings)
					if err != nil {
						return false, err
					}

					// display any reading value in row 3
					settings = make(map[string]string)
					settings["displayRow3"] = eventReading.Value
					err = vendingState.SendCommand(lc, http.MethodPut, eventReading.DeviceName, vendingState.Configuration.DeviceControllerBoarddisplayRow3, settings)
					if err != nil {
						return false, err
					}

					// send lock command
					settings = make(map[string]string)
					settings["lock1"] = "true"
					err = vendingState.SendCommand(lc, http.MethodPut, eventReading.DeviceName, vendingState.Configuration.DeviceControllerBoardLock1, settings)
					if err != nil {
						return false, err
					}

					vendingState.MaintenanceMode = false
					vendingState.CVWorkflowStarted = false
					vendingState.DoorClosedDuringCVWorkflow = false
					vendingState.DoorOpenedDuringCVWorkflow = false
					vendingState.InferenceDataReceived = false
					lc.Infof("Maintenance Scan")
					lc.Infof("workflow: +%v", vendingState.CVWorkflowStarted)
					lc.Infof("maintenance mode: +%v", vendingState.MaintenanceMode)
					lc.Infof("open: +%v", vendingState.DoorOpenedDuringCVWorkflow)
					lc.Infof("closed: +%v", vendingState.DoorClosedDuringCVWorkflow)
					lc.Infof("Inference: +%v ", vendingState.InferenceDataReceived)
					lc.Infof("door: +%v", vendingState.DoorClosed)
				}
			default:
				// display "Unauthorized" on display row 2
				settings := make(map[string]string)
				settings["displayRow2"] = "Unauthorized"
				err := vendingState.SendCommand(lc, http.MethodPut, eventReading.DeviceName, vendingState.Configuration.DeviceControllerBoarddisplayRow2, settings)
				if err != nil {
					return false, err
				}
				lc.Infof("Invalid card: %s", eventReading.Value)
			}
		}
	}
	return true, event // Continues the functions pipeline execution with the current event
}

func (vendingState *VendingState) checkInferenceStatus(lc logger.LoggingClient, heartbeatEndPoint string, deviceName string) bool {
	err := vendingState.SendCommand(lc, http.MethodGet, deviceName, heartbeatEndPoint, nil)
	if err != nil {
		lc.Errorf("error checking inference status: %v", err)
		return false
	}

	return true
}

func (vendingState *VendingState) getCardAuthInfo(lc logger.LoggingClient, authEndpoint string, cardID string) {
	// Push the authenticated user info to the current vendingState
	// First, reset it, then populate it at the end of the function
	vendingState.CurrentUserData = OutputData{}

	resp, err := sendHttpRequest(lc, http.MethodGet, authEndpoint+"/"+cardID, []byte(""))
	if err != nil {
		lc.Infof("Unauthorized card: %s", cardID)
		return
	}

	defer resp.Body.Close()

	var auth OutputData
	_, err = utilities.ParseJSONHTTPResponseContent(resp.Body, &auth)
	if err != nil {
		lc.Errorf("Could not read response body from AuthenticationEndpoint: %v", err)
		return
	}

	// Set the door waiting state to false while processing a use
	vendingState.CurrentUserData = auth
	lc.Info("Successfully found user data for card " + cardID)
}

func (vendingState *VendingState) displayLedger(lc logger.LoggingClient, deviceName string, ledger Ledger) error {
	settings := make(map[string]string)
	settings["displayReset"] = ""
	// reset LCD
	err := vendingState.SendCommand(lc, http.MethodPut, deviceName, vendingState.Configuration.DeviceControllerBoarddisplayReset, settings)
	if err != nil {
		return fmt.Errorf("sendCommand returned error for %v : %v", vendingState.Configuration.DeviceControllerBoarddisplayReset, err.Error())
	}

	// Loop through lineItems in Ledger and display on LCD
	for _, lineItem := range ledger.LineItems {
		// Line Item is item count and product name, and truncated or padded with whitespaces so it is same length as LCD Row
		displayLineItem := fmt.Sprintf("%-[1]*.[1]*s", vendingState.Configuration.LCDRowLength, strconv.Itoa(lineItem.ItemCount)+" "+lineItem.ProductName)
		settings := make(map[string]string)
		settings["displayRow1"] = displayLineItem
		// push line item to LCD and pause three seconds before displaying next line item
		err := vendingState.SendCommand(lc, http.MethodPut, deviceName, vendingState.Configuration.DeviceControllerBoarddisplayRow1, settings)
		if err != nil {
			return fmt.Errorf("sendCommand returned nil for %v : %v", vendingState.Configuration.DeviceControllerBoarddisplayRow1, err.Error())
		}

		time.Sleep(3 * time.Second)
	}

	//reset the LCD
	settings = make(map[string]string)
	settings["displayReset"] = ""
	err = vendingState.SendCommand(lc, http.MethodPut, deviceName, vendingState.Configuration.DeviceControllerBoarddisplayReset, settings)
	if err != nil {
		return fmt.Errorf("sendCommand returned nil for %v : %v", vendingState.Configuration.DeviceControllerBoarddisplayReset, err.Error())
	}

	//display ledger.LineTotal from in currency format
	displayLedgerTotal := "Total: $" + fmt.Sprintf("%3.2f", ledger.LineTotal)
	settings = make(map[string]string)
	settings["displayRow1"] = displayLedgerTotal
	err = vendingState.SendCommand(lc, http.MethodPut, deviceName, vendingState.Configuration.DeviceControllerBoarddisplayRow1, settings)
	if err != nil {
		return fmt.Errorf("sendCommand returned nil for %v : %v", vendingState.Configuration.DeviceControllerBoarddisplayRow1, err.Error())
	}

	return nil
}

func (vendingState *VendingState) SendCommand(lc logger.LoggingClient, actionName string, deviceName string,
	commandName string, settings map[string]string) error {
	lc.Debug("Sending Command")

	commandClient := vendingState.Service.CommandClient()
	if commandClient == nil {
		return errors.New("error command service missing from client's configuration")
	}

	switch actionName {
	case http.MethodPut:
		lc.Infof("executing %s action", actionName)
		lc.Infof("Sending command '%s' for device '%s'", commandName, deviceName)

		response, err := commandClient.IssueSetCommandByName(context.Background(), deviceName, commandName, settings)
		//(res dtoCommon.BaseResponse, err errors.EdgeX)
		if err != nil {
			return fmt.Errorf("failed to send '%s' set command to '%s' device: %s", commandName, deviceName, err.Error())
		}

		lc.Infof("response status: %d", response.StatusCode)

	case http.MethodGet:
		lc.Infof("executing %s action", actionName)
		lc.Infof("Sending command '%s' for device '%s'", commandName, deviceName)
		response, err := commandClient.IssueGetCommandByName(context.Background(), deviceName, commandName, "no", "yes")
		if err != nil {
			return fmt.Errorf("failed to send '%s' get command to '%s' device: %s", commandName, deviceName, err.Error())
		}
		lc.Infof("response status: %d", response.StatusCode)

	default:
		lc.Errorf("Invalid action requested: %s", actionName)
		return errors.New("Invalid action requested: " + actionName)
	}

	return nil
}

// sendHttpRequest will make an http request to an EdgeX command endpoint
func sendHttpRequest(lc logger.LoggingClient, method string, commandURL string, inputBytes []byte) (*http.Response, error) {

	lc.Debugf("sending command to edgex endpoint: %v", commandURL)

	// Create the http request based on the parameters
	request, _ := http.NewRequest(method, commandURL, bytes.NewBuffer(inputBytes))
	timeout := 60 * time.Second
	client := &http.Client{
		Timeout: timeout,
	}

	// Execute the http request
	resp, err := client.Do(request)
	if err != nil {
		return resp, fmt.Errorf("error sending command: %v", err.Error())
	}

	// Check the status code and return any errors
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("error sending command: received status code: %v", resp.Status)
	}

	return resp, nil
}

// BoardStatus endpoint that handles board status events from the controller board status application service
func (vendingState *VendingState) BoardStatus(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	var status int

	// Read request body
	body := make([]byte, req.ContentLength)
	_, err := io.ReadFull(req.Body, body)
	if err != nil {
		fmt.Printf("Failed to read request data\n")
	}

	// Unmarshal the string contents of request into a proper structure
	var boardStatus ControllerBoardStatus
	if err := json.Unmarshal(body, &boardStatus); err != nil {
		fmt.Printf("Failed to read request data\n")
	}
	returnval := "Board status received but maintenance mode was not set"
	status = http.StatusOK

	// Check controller board MinTemperatureStatus state. If it's true then a minimum temperature event has happened
	if boardStatus.MinTemperatureStatus {
		returnval = string("Temperature status received and maintenance mode was set")
		status = http.StatusOK
		fmt.Println("Cooler temperature exceeds the minimum temperature threshold. The cooler needs maintenance.")
		vendingState.MaintenanceMode = true
	}
	// Check controller board MaxTemperatureStatus state. If it's true then a maximum temperature event has happened
	if boardStatus.MaxTemperatureStatus {
		returnval = string("Temperature status received and maintenance mode was set")
		status = http.StatusOK
		fmt.Println("Cooler temperature exceeds the maximum temperature threshold. The cooler needs maintenance.")
		vendingState.MaintenanceMode = true
	}

	// Check to see if the board closed state is different than the previous state. If it is we need to update the state and
	// set the related properties.
	if vendingState.DoorClosed != boardStatus.DoorClosed {
		fmt.Println("Successfully updated the door event. Door closed:", boardStatus.DoorClosed)
		returnval = string("Door closed change event was received ")
		status = http.StatusOK //FIXME: This is an issue
		vendingState.DoorClosed = boardStatus.DoorClosed
		if vendingState.CVWorkflowStarted {
			// If the door was opened then we want to wait for the door closed event
			if !boardStatus.DoorClosed {
				vendingState.DoorOpenedDuringCVWorkflow = true
				// Stop the open wait thread since the door is now opened
				close(vendingState.DoorOpenWaitThreadStopChannel)
				vendingState.DoorOpenWaitThreadStopChannel = make(chan int)

				// Wait for door closed event. If the door isn't closed within the timeout
				// then leave the workflow, remove the user data, and enter maintenance mode
				go func() {
					fmt.Println("Door Opened: wait for ", vendingState.Configuration.DoorCloseStateTimeout, " seconds")
					for {
						select {
						case <-time.After(vendingState.Configuration.DoorCloseStateTimeout):
							{
								if !vendingState.DoorClosedDuringCVWorkflow {
									fmt.Println("Door Opened: Failed")
									vendingState.CVWorkflowStarted = false
									vendingState.CurrentUserData = OutputData{}
									vendingState.MaintenanceMode = true
								}
								return
							}
						case <-vendingState.DoorCloseWaitThreadStopChannel:
							fmt.Println("Stopped the door closed wait thread")
							return

						case <-vendingState.ThreadStopChannel:
							fmt.Println("Globally stopped the door closed wait thread")
							return
						}
					}
				}()
			}
			// If the door was closed we want to wait for the inference event
			if boardStatus.DoorClosed {
				vendingState.DoorClosedDuringCVWorkflow = true
				// Stop the open wait thread since the door is now opened
				close(vendingState.DoorCloseWaitThreadStopChannel)
				vendingState.DoorCloseWaitThreadStopChannel = make(chan int)

				// Wait for the inference data to be received. If we don't receive any inference data with the timeout
				// then leave the workflow, remove the user data, and enter maintenance mode
				go func() {
					fmt.Println("Door Closed: wait for ", vendingState.Configuration.InferenceTimeout, " seconds")
					for {
						select {
						case <-time.After(vendingState.Configuration.InferenceTimeout):
							{
								if !vendingState.InferenceDataReceived {
									fmt.Println("Door Closed: Failed")
									vendingState.CVWorkflowStarted = false
									vendingState.CurrentUserData = OutputData{}
									vendingState.MaintenanceMode = true
								}
								return
							}
						case <-vendingState.InferenceWaitThreadStopChannel:
							fmt.Println("Stopped the inference wait thread")
							return

						case <-vendingState.ThreadStopChannel:
							fmt.Println("Globally stopped the inference wait thread")
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
		fmt.Printf("Failed to write item data back to caller\n")
	}
}

// ResetDoorLock endpoint to reset all door lock states
func (vendingState *VendingState) ResetDoorLock(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	// Check the HTTP Request's form values
	returnval := "reset the door lock"

	close(vendingState.ThreadStopChannel)
	vendingState.ThreadStopChannel = make(chan int)

	vendingState.MaintenanceMode = false
	vendingState.CVWorkflowStarted = false
	vendingState.DoorClosed = false
	vendingState.DoorClosedDuringCVWorkflow = false
	vendingState.DoorOpenedDuringCVWorkflow = false
	vendingState.InferenceDataReceived = false

	fmt.Println("Maintenance card scanned")
	fmt.Printf("workflow: %t", vendingState.CVWorkflowStarted)
	fmt.Printf("maintenance mode: %t", vendingState.MaintenanceMode)
	fmt.Printf("open: %t", vendingState.DoorOpenedDuringCVWorkflow)
	fmt.Printf("closed: %t", vendingState.DoorClosedDuringCVWorkflow)
	fmt.Printf("Inference: %t", vendingState.InferenceDataReceived)
	fmt.Printf("door: %t", vendingState.DoorClosed)

	// Write the HTTP status header
	writer.WriteHeader(http.StatusOK)

	_, writeErr := writer.Write([]byte(returnval))
	if writeErr != nil {
		fmt.Printf("Failed to write item data back to caller\n")
	}

}

// GetMaintenanceMode will return a JSON response containing the boolean state
// of the vendingState's maintenance mode.
func (vendingState *VendingState) GetMaintenanceMode(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		mm, _ := utilities.GetAsJSON(MaintenanceMode{MaintenanceMode: vendingState.MaintenanceMode})
		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, mm, false)
	})
}
