// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

const (
	InferenceMQTTDevice = "Inference-MQTT-device"
	DefaultTimeOuts     = 15
)

// DeviceHelper is an EdgeX function that is passed into the EdgeX SDK's function pipeline.
// It is a decision function that allows for multiple devices to have their events processed
// correctly by this application service.
func (vendingState *VendingState) DeviceHelper(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	event := params[0].(models.Event)

	switch event.Device {
	case "ds-card-reader":
		{
			return vendingState.VerifyDoorAccess(edgexcontext, event)
		}
	case InferenceMQTTDevice:
		{
			return vendingState.HandleMqttDeviceReading(edgexcontext, event)
		}
	default:
		{
			return false, nil
		}
	}
}

// HandleMqttDeviceReading is an EdgeX function that simply handles events coming from
// the MQTT device service.
func (vendingState *VendingState) HandleMqttDeviceReading(edgexcontext *appcontext.Context, event models.Event) (bool, interface{}) {
	if event.Device == InferenceMQTTDevice {
		for _, eventReading := range event.Readings {
			fmt.Println(eventReading.Value)
		}
		edgexcontext.LoggingClient.Debug("Inference mqtt device",
			"workflow:", vendingState.CVWorkflowStarted,
			"maintenance mode:", vendingState.MaintenanceMode,
			"open:", vendingState.DoorOpenedDuringCVWorkflow,
			"closed:", vendingState.DoorClosedDuringCVWorkflow,
			"Inference:", vendingState.InferenceDataReceived,
			"door:", vendingState.DoorClosed,
		)

		edgexcontext.LoggingClient.Debug("Processing reading from MQTT device service")
		for _, eventReading := range event.Readings {
			switch eventReading.Name {
			case "inferenceSkuDelta":
				{
					fmt.Println("Inference Started")
					var skuDelta []deltaSKU
					if err := json.Unmarshal([]byte(eventReading.Value), &skuDelta); err != nil {
						edgexcontext.LoggingClient.Error("HandleMqttDeviceReading failed to unmarshal skuDelta message")
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
						ledgerCommandURL, ok := edgexcontext.Configuration.ApplicationSettings["LedgerService"]
						if !ok {
							return false, fmt.Errorf("LedgerService application API endpoint setting not found")
						}

						outputBytes, err := json.Marshal(deltaLedger)
						if err != nil {
							edgexcontext.LoggingClient.Error("HandleMqttDeviceReading failed to marshal deltaLedger")
							return false, err
						}

						edgexcontext.LoggingClient.Info("Sending SKU delta to ledger service")

						// send SKU delta to ledger service and get back current ledger information
						resp, err := sendCommand(edgexcontext, "POST", ledgerCommandURL, outputBytes)
						if err != nil {
							edgexcontext.LoggingClient.Error("Ledger service failed: %v", err.Error())
							return false, err
						}

						defer resp.Body.Close()

						edgexcontext.LoggingClient.Info("Successfully updated the user's ledger")

						var currentLedger Ledger
						_, err = utilities.ParseJSONHTTPResponseContent(resp.Body, &currentLedger)
						if err != nil {
							return false, fmt.Errorf("Unable to unmarshal ledger response")
						}

						// Display Ledger Total on LCD
						if displayErr := displayLedger(edgexcontext, currentLedger); displayErr != nil {
							return false, displayErr
						}

					}

					// POST the deltaLedger json string to the inventory endpoint
					inventoryCommandURL, ok := edgexcontext.Configuration.ApplicationSettings["InventoryService"]
					if !ok {
						return false, fmt.Errorf("InventoryService application API endpoint setting not found")
					}

					outputBytes, err := json.Marshal(deltaLedger.DeltaSKUs)
					if err != nil {
						return false, fmt.Errorf("HandleMqttDeviceReading failed to marshal deltaLedger.DeltaSKUs")
					}

					edgexcontext.LoggingClient.Info("Sending SKU delta to inventory service")
					inventoryResp, err := sendCommand(edgexcontext, "POST", inventoryCommandURL, outputBytes)
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

					auditLogCommandURL, ok := edgexcontext.Configuration.ApplicationSettings["InventoryAuditLogService"]
					if !ok {
						return false, fmt.Errorf("InventoryAuditLogService application API endpoint setting not found")
					}

					outputBytes, err = json.Marshal(auditLogEntry)
					if err != nil {
						return false, err
					}

					edgexcontext.LoggingClient.Info("Sending audit log entry to inventory service")
					auditResp, err := sendCommand(edgexcontext, "POST", auditLogCommandURL, outputBytes)
					if err != nil {
						return false, err
					}
					defer auditResp.Body.Close()
					vendingState.CurrentUserData = OutputData{}
					vendingState.CVWorkflowStarted = false
					edgexcontext.LoggingClient.Info("Inference complete and workflow status reset")
					// Close all thread to ensure all threads are cleaned up before the next card is scanned.
					close(vendingState.ThreadStopChannel)
					vendingState.ThreadStopChannel = make(chan int)
				}
			default:
				{
					edgexcontext.LoggingClient.Info("Received an event with an unknown name")
					return false, nil
				}
			}
		}
	}

	return false, nil
}

// VerifyDoorAccess will take the card reader events and verify the read card id against the white list
// If the card is valid the function will send the unlock message to the device-controller-board device service
func (vendingState *VendingState) VerifyDoorAccess(edgexcontext *appcontext.Context, event models.Event) (bool, interface{}) {
	edgexcontext.LoggingClient.Info("new card scanned", "workflow:", vendingState.CVWorkflowStarted, "maintenance mode:", vendingState.MaintenanceMode, "open:", vendingState.DoorOpenedDuringCVWorkflow, "closed:", vendingState.DoorClosedDuringCVWorkflow, "Inference:", vendingState.InferenceDataReceived, "door:", vendingState.DoorClosed)

	if event.Device == "ds-card-reader" && !vendingState.CVWorkflowStarted {
		edgexcontext.LoggingClient.Info("Verify the card reader input against the white list")

		edgexcontext.LoggingClient.Info("Card Scanned", "workflow:", vendingState.CVWorkflowStarted, "maintenance mode:", vendingState.MaintenanceMode, "open:", vendingState.DoorOpenedDuringCVWorkflow, "closed:", vendingState.DoorClosedDuringCVWorkflow, "Inference:", vendingState.InferenceDataReceived, "door:", vendingState.DoorClosed)
		// check to see if inference is running and set maintenance mode accordingly
		if !vendingState.MaintenanceMode {
			inferenceHeartbeatEndpoint, ok := edgexcontext.Configuration.ApplicationSettings["InferenceHeartbeat"]
			if !ok {
				return false, fmt.Errorf("inferenceHeartbeatEndpoint Application setting not found")
			}
			vendingState.MaintenanceMode = !checkInferenceStatus(edgexcontext, inferenceHeartbeatEndpoint)
		}

		// Retrieve & Hit auth endpoint
		authEndpoint, ok := edgexcontext.Configuration.ApplicationSettings["AuthenticationEndpoint"]
		if !ok {
			return false, fmt.Errorf("AuthenticationEndpoint Application setting not found")
		}

		for _, eventReading := range event.Readings {
			vendingState.getCardAuthInfo(edgexcontext, authEndpoint, eventReading.Value)

			switch vendingState.CurrentUserData.RoleID {
			// Check the role of the card scanned. Role 1 = customer and Role 2 = item stocker
			case 1, 2:
				{
					if !vendingState.MaintenanceMode {
						edgexcontext.LoggingClient.Info(eventReading.Name + " readable value from " + event.Device + " is " + eventReading.Value)
						// Display row 2 command URL
						commandURL, ok := edgexcontext.Configuration.ApplicationSettings["DeviceControllerBoarddisplayRow2"]
						if !ok {
							return false, fmt.Errorf("DeviceControllerBoard Application setting not found")
						}
						// Send lock command
						resp, err := sendCommand(edgexcontext, "PUT", commandURL, []byte("{\"displayRow2\":\"hello\"}"))
						if err != nil {
							return false, err
						}
						defer resp.Body.Close()

						// Display row 3 command URL
						commandURL, ok = edgexcontext.Configuration.ApplicationSettings["DeviceControllerBoarddisplayRow3"]
						if !ok {
							return false, fmt.Errorf("DeviceControllerBoard Application setting not found")
						}
						// Send lock command
						respRow3, err := sendCommand(edgexcontext, "PUT", commandURL, []byte("{\"displayRow3\":\""+eventReading.Value+"\"}"))
						if err != nil {
							return false, err
						}
						defer respRow3.Body.Close()

						// Lock command URL
						commandURL, ok = edgexcontext.Configuration.ApplicationSettings["DeviceControllerBoardLock1"]
						if !ok {
							return false, fmt.Errorf("DeviceControllerBoard Application setting not found")
						}
						// Send lock command
						respLock, err := sendCommand(edgexcontext, "PUT", commandURL, []byte("{\"lock1\":\"true\"}"))
						if err != nil {
							return false, err
						}
						defer respLock.Body.Close()

						// Start the workflow state and set all of the thread states to false
						vendingState.CVWorkflowStarted = true
						vendingState.DoorClosedDuringCVWorkflow = false
						vendingState.DoorOpenedDuringCVWorkflow = false
						vendingState.InferenceDataReceived = false
						// Get the latest timeouts from the toml configuration
						SetDefaultTimeouts(vendingState, edgexcontext.Configuration.ApplicationSettings, edgexcontext.LoggingClient)

						// Wait for the door open event to be received. If we don't receive the door open event within the timeout
						// then leave the workflow state and remove all user data
						go func() {
							for {
								select {
								case <-time.After(time.Duration(vendingState.DoorOpenStateTimeout) * time.Second):
									if !vendingState.DoorOpenedDuringCVWorkflow {
										edgexcontext.LoggingClient.Info("door wasn't opened so we reset")
										vendingState.CVWorkflowStarted = false
										vendingState.CurrentUserData = OutputData{}
									}
									edgexcontext.LoggingClient.Info("Card scan: Waiting for open event", "workflow:", vendingState.CVWorkflowStarted, "maintenance mode:", vendingState.MaintenanceMode, "open:", vendingState.DoorOpenedDuringCVWorkflow, "closed:", vendingState.DoorClosedDuringCVWorkflow, "Inference:", vendingState.InferenceDataReceived, "door:", vendingState.DoorClosed)
									return

								case <-vendingState.DoorOpenWaitThreadStopChannel:
									edgexcontext.LoggingClient.Info("Stopped the door open wait thread")
									return

								case <-vendingState.ThreadStopChannel:
									edgexcontext.LoggingClient.Info("Globally stopped the door open wait thread")
									return
								}
							}
						}()
					} else {
						// Display row 1 command URL
						commandURL, ok := edgexcontext.Configuration.ApplicationSettings["DeviceControllerBoarddisplayRow1"]
						if !ok {
							return false, fmt.Errorf("DeviceControllerBoard Application setting not found")
						}
						// Display out of order when door waiting state is set to false
						resp, err := sendCommand(edgexcontext, "PUT", commandURL, []byte("{\"displayRow1\":\"Out of Order\"}"))
						if err != nil {
							return false, err
						}
						defer resp.Body.Close()
					}

				}
			// Check the role of the card scanned. Role 3 = maintainer
			case 3:
				{
					close(vendingState.ThreadStopChannel)
					vendingState.ThreadStopChannel = make(chan int)

					edgexcontext.LoggingClient.Info(eventReading.Name + " readable value from " + event.Device + " is " + eventReading.Value)
					// Display row 2 command URL
					commandURL, ok := edgexcontext.Configuration.ApplicationSettings["DeviceControllerBoarddisplayRow2"]
					if !ok {
						return false, fmt.Errorf("DeviceControllerBoard Application setting not found")
					}
					// Send lock command
					respRow2, err := sendCommand(edgexcontext, "PUT", commandURL, []byte("{\"displayRow2\":\"Maintenance Mode\"}"))
					if err != nil {
						return false, err
					}
					defer respRow2.Body.Close()

					// Display row 3 command URL
					commandURL, ok = edgexcontext.Configuration.ApplicationSettings["DeviceControllerBoarddisplayRow3"]
					if !ok {
						return false, fmt.Errorf("DeviceControllerBoard Application setting not found")
					}
					// Send lock command
					respRow3, err := sendCommand(edgexcontext, "PUT", commandURL, []byte("{\"displayRow3\":\""+eventReading.Value+"\"}"))
					if err != nil {
						return false, err
					}
					defer respRow3.Body.Close()

					// Lock command URL
					commandURL, ok = edgexcontext.Configuration.ApplicationSettings["DeviceControllerBoardLock1"]
					if !ok {
						return false, fmt.Errorf("DeviceControllerBoard Application setting not found")
					}
					// Send lock command
					respLock, err := sendCommand(edgexcontext, "PUT", commandURL, []byte("{\"lock1\":\"true\"}"))
					if err != nil {
						return false, err
					}
					defer respLock.Body.Close()

					vendingState.MaintenanceMode = false
					vendingState.CVWorkflowStarted = false
					vendingState.DoorClosedDuringCVWorkflow = false
					vendingState.DoorOpenedDuringCVWorkflow = false
					vendingState.InferenceDataReceived = false
					edgexcontext.LoggingClient.Info("Maintenance Scan: ", "workflow:", vendingState.CVWorkflowStarted, "maintenance mode:", vendingState.MaintenanceMode, "open:", vendingState.DoorOpenedDuringCVWorkflow, "closed:", vendingState.DoorClosedDuringCVWorkflow, "Inference:", vendingState.InferenceDataReceived, "door:", vendingState.DoorClosed)
				}
			default:
				// Display row 2 command URL
				commandURL, ok := edgexcontext.Configuration.ApplicationSettings["DeviceControllerBoarddisplayRow2"]
				if !ok {
					return false, fmt.Errorf("DeviceControllerBoard Application setting not found")
				}
				resp, err := sendCommand(edgexcontext, "PUT", commandURL, []byte("{\"displayRow2\":\"Unauthorized\"}"))
				defer resp.Body.Close()
				if err != nil {
					return false, err
				}
				edgexcontext.LoggingClient.Info("Invalid card: " + eventReading.Value)
			}
		}
	}
	return true, event // Continues the functions pipeline execution with the current event
}

func checkInferenceStatus(edgexcontext *appcontext.Context, heartbeatEndPoint string) bool {
	resp, err := sendCommand(edgexcontext, "GET", heartbeatEndPoint, []byte(""))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return true
}

func (vendingState *VendingState) getCardAuthInfo(edgexcontext *appcontext.Context, authEndpoint string, cardID string) {
	// Push the authenticated user info to the current vendingState
	// First, reset it, then populate it at the end of the function
	vendingState.CurrentUserData = OutputData{}

	resp, err := sendCommand(edgexcontext, "GET", authEndpoint+"/"+cardID, []byte(""))
	if err != nil {
		edgexcontext.LoggingClient.Info("Unauthorized card: " + cardID)
		return
	}

	defer resp.Body.Close()

	var auth OutputData
	_, err = utilities.ParseJSONHTTPResponseContent(resp.Body, &auth)
	if err != nil {
		edgexcontext.LoggingClient.Error("Could not read response body from AuthenticationEndpoint")
		return
	}

	// Set the door waiting state to false while processing a use
	vendingState.CurrentUserData = auth
	edgexcontext.LoggingClient.Info("Successfully found user data for card " + cardID)
}

func displayLedger(edgexcontext *appcontext.Context, ledger Ledger) error {
	// string conversion error
	var strToIntErr error
	// Get LCDRowLength value
	LCDRowLength, strToIntErr := strconv.Atoi(edgexcontext.Configuration.ApplicationSettings["LCDRowLength"])
	if strToIntErr != nil {
		LCDRowLength = 19
		edgexcontext.LoggingClient.Error("LCD Row Length not set in configuration. Using default value of %v", LCDRowLength)
	}

	// Get commandURL to reset LCD
	displayResetCommandURL, ok := edgexcontext.Configuration.ApplicationSettings["DeviceControllerBoarddisplayReset"]
	if !ok {
		return fmt.Errorf("DeviceControllerBoarddisplayReset setting not found")
	}
	// reset LCD
	resp, err := sendCommand(edgexcontext, "PUT", displayResetCommandURL, []byte("{\"displayReset\":\"\"}"))
	if err != nil {
		return fmt.Errorf("sendCommand returned error for %v : %v", displayResetCommandURL, err.Error())
	}
	defer resp.Body.Close()

	// Get command url for LCD Row 1
	row1CommandURL, ok := edgexcontext.Configuration.ApplicationSettings["DeviceControllerBoarddisplayRow1"]
	if !ok {
		return fmt.Errorf("DeviceControllerBoard Application setting not found")
	}
	// Loop through lineItems in Ledger and display on LCD
	for _, lineItem := range ledger.LineItems {
		// Line Item is item count and product name, and truncated or padded with whitespaces so it is same length as LCD Row
		displayLineItem := fmt.Sprintf("%-[1]*.[1]*s", LCDRowLength, strconv.Itoa(lineItem.ItemCount)+" "+lineItem.ProductName)
		// push line item to LCD and pause three seconds before displaying next line item
		resp, err := sendCommand(edgexcontext, "PUT", row1CommandURL, []byte("{\"displayRow1\":\""+displayLineItem+"\"}"))
		if err != nil {
			return fmt.Errorf("sendCommand returned nil for %v : %v", row1CommandURL, err.Error())
		}
		defer resp.Body.Close()

		time.Sleep(3 * time.Second)
	}

	//reset the LCD
	respDisplayReset, err := sendCommand(edgexcontext, "PUT", displayResetCommandURL, []byte("{\"displayReset\":\"\"}"))
	if err != nil {
		return fmt.Errorf("sendCommand returned nil for %v : %v", displayResetCommandURL, err.Error())
	}
	defer respDisplayReset.Body.Close()

	//display ledger.LineTotal from in currency format
	displayLedgerTotal := "Total: $" + fmt.Sprintf("%3.2f", ledger.LineTotal)
	respDisplayRow, err := sendCommand(edgexcontext, "PUT", row1CommandURL, []byte("{\"displayRow1\":\""+displayLedgerTotal+"\"}"))
	if err != nil {
		return fmt.Errorf("sendCommand returned nil for %v : %v", row1CommandURL, err.Error())
	}
	defer respDisplayRow.Body.Close()

	return nil
}

// sendCommand will make a http request based on the input parameters
func sendCommand(edgexcontext *appcontext.Context, method string, commandURL string, inputBytes []byte) (*http.Response, error) {
	edgexcontext.LoggingClient.Debug("SendCommand")

	connectionTimeout := 15
	// Create the http request based on the parameters
	request, _ := http.NewRequest(method, commandURL, bytes.NewBuffer(inputBytes))
	timeout := time.Duration(connectionTimeout) * time.Second
	client := &http.Client{
		Timeout: timeout,
	}

	// Execute the http request
	resp, err := client.Do(request)
	if err != nil {
		return resp, fmt.Errorf("Error sending DeviceControllerBoard: %v", err.Error())
	}

	// Check the status code and return any errors
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("Error sending DeviceControllerBoard: Received status code %v", resp.Status)
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
					fmt.Println("Door Opened: wait for ", vendingState.DoorCloseStateTimeout, " seconds")
					for {
						select {
						case <-time.After(time.Duration(vendingState.DoorCloseStateTimeout) * time.Second):
							{
								if !vendingState.DoorClosedDuringCVWorkflow {
									fmt.Println("Door Opened: Failed")
									vendingState.CVWorkflowStarted = false
									vendingState.CurrentUserData = OutputData{}
									vendingState.MaintenanceMode = true
								}
								// TODO: remove print
								fmt.Println("Door Opened: waiting for door closed event", "workflow:", vendingState.CVWorkflowStarted, "maintenance mode:", vendingState.MaintenanceMode, "open:", vendingState.DoorOpenedDuringCVWorkflow, "closed:", vendingState.DoorClosedDuringCVWorkflow, "Inference:", vendingState.InferenceDataReceived, "door:", vendingState.DoorClosed)
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
					fmt.Println("Door Closed: wait for ", vendingState.InferenceTimeout, " seconds")
					for {
						select {
						case <-time.After(time.Duration(vendingState.InferenceTimeout) * time.Second):

							{
								if !vendingState.InferenceDataReceived {
									fmt.Println("Door Closed: Failed")
									vendingState.CVWorkflowStarted = false
									vendingState.CurrentUserData = OutputData{}
									vendingState.MaintenanceMode = true
								}
								// TODO: remove print
								fmt.Println("Door Closed: waiting for door closed event", "workflow:", vendingState.CVWorkflowStarted, "maintenance mode:", vendingState.MaintenanceMode, "open:", vendingState.DoorOpenedDuringCVWorkflow, "closed:", vendingState.DoorClosedDuringCVWorkflow, "Inference:", vendingState.InferenceDataReceived, "door:", vendingState.DoorClosed)
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

	fmt.Println("Maintenance card scanned: reset everything",
		"workflow:", vendingState.CVWorkflowStarted,
		"maintenance mode:", vendingState.MaintenanceMode,
		"open:", vendingState.DoorOpenedDuringCVWorkflow,
		"closed:", vendingState.DoorClosedDuringCVWorkflow,
		"Inference:", vendingState.InferenceDataReceived,
		"door:", vendingState.DoorClosed,
	)

	// Write the HTTP status header
	writer.WriteHeader(http.StatusOK)

	_, writeErr := writer.Write([]byte(returnval))
	if writeErr != nil {
		fmt.Printf("Failed to write item data back to caller\n")
	}

}

// SetDefaultTimeouts set the timeout values based on the toml configuration. If the value is not found use a default value
func SetDefaultTimeouts(vendingState *VendingState, appSettings map[string]string, loggingClient logger.LoggingClient) {
	// string conversion error
	var strToIntErr error

	// door open event
	openTimeout, strToIntErr := strconv.Atoi(appSettings["DoorOpenStateTimeout"])
	if strToIntErr != nil {
		openTimeout = DefaultTimeOuts
		loggingClient.Error(fmt.Sprintf("Door Open event timeout not set in configuration. Using default value of %v", openTimeout))
	}
	vendingState.DoorOpenStateTimeout = openTimeout

	// door close event
	closeTimeout, strToIntErr := strconv.Atoi(appSettings["DoorCloseStateTimeout"])
	if strToIntErr != nil {
		closeTimeout = DefaultTimeOuts
		loggingClient.Error(fmt.Sprintf("Door close event timeout not set in configuration. Using default value of %v", closeTimeout))
	}
	vendingState.DoorCloseStateTimeout = closeTimeout

	// inference event
	inferenceTimeout, strToIntErr := strconv.Atoi(appSettings["InferenceTimeout"])
	if strToIntErr != nil {
		inferenceTimeout = DefaultTimeOuts
		loggingClient.Error(fmt.Sprintf("Door close event timeout not set in configuration. Using default value of %v", inferenceTimeout))
	}
	vendingState.InferenceTimeout = inferenceTimeout
}

// GetMaintenanceMode will return a JSON response containing the boolean state
// of the vendingState's maintenance mode.
func (vendingState *VendingState) GetMaintenanceMode(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		mm, _ := utilities.GetAsJSON(MaintenanceMode{MaintenanceMode: vendingState.MaintenanceMode})
		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, mm, false)
	})
}
