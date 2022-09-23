package config

import (
	"fmt"
)

type ServiceConfig struct {
	Vending VendingConfig
}

type VendingConfig struct {
	AuthenticationEndpoint         string
	ControllerBoardDisplayResetCmd string
	ControllerBoardDisplayRow0Cmd  string
	ControllerBoardDisplayRow1Cmd  string
	ControllerBoardDisplayRow2Cmd  string
	ControllerBoardDisplayRow3Cmd  string
	ControllerBoardLock1Cmd        string
	ControllerBoardLock2Cmd        string
	CardReaderDeviceName           string
	InferenceDeviceName            string
	ControllerBoardDeviceName      string
	DoorCloseStateTimeoutDuration  string
	DoorOpenStateTimeoutDuration   string
	InferenceDoorStatusCmd         string
	InferenceHeartbeatCmd          string
	InferenceTimeoutDuration       string
	InventoryAuditLogService       string
	InventoryService               string
	LCDRowLength                   int
	LedgerService                  string
}

// UpdateFromRaw updates the service's full configuration from raw data received from
// the Service Provider.
func (c *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false
	}

	*c = *configuration

	return true
}

// Validate ensures your custom configuration has proper values.
func (ac *VendingConfig) Validate() error {
	if len(ac.AuthenticationEndpoint) == 0 {
		return fmt.Errorf("configuration AuthenticationEndpoint is empty")
	}

	if len(ac.ControllerBoardDisplayResetCmd) == 0 {
		return fmt.Errorf("configuration ControllerBoardDisplayResetCmd is empty")
	}

	if len(ac.ControllerBoardDisplayRow0Cmd) == 0 {
		return fmt.Errorf("configuration ControllerBoardDisplayRow0Cmd is empty")
	}

	if len(ac.ControllerBoardDisplayRow1Cmd) == 0 {
		return fmt.Errorf("configuration ControllerBoardDisplayRow1Cmd is empty")
	}

	if len(ac.ControllerBoardDisplayRow2Cmd) == 0 {
		return fmt.Errorf("configuration ControllerBoardDisplayRow2Cmd is empty")
	}

	if len(ac.ControllerBoardDisplayRow3Cmd) == 0 {
		return fmt.Errorf("configuration ControllerBoardDisplayRow3Cmd is empty")
	}

	if len(ac.ControllerBoardLock1Cmd) == 0 {
		return fmt.Errorf("configuration ControllerBoardLock1Cmd is empty")
	}

	if len(ac.ControllerBoardLock2Cmd) == 0 {
		return fmt.Errorf("configuration ControllerBoardLock2Cmd is empty")
	}

	if len(ac.CardReaderDeviceName) == 0 {
		return fmt.Errorf("configuration CardReaderDeviceName is empty")
	}

	if len(ac.InferenceDeviceName) == 0 {
		return fmt.Errorf("configuration InferenceDeviceName is empty")
	}

	if len(ac.ControllerBoardDeviceName) == 0 {
		return fmt.Errorf("configuration ControllerBoardDeviceName is empty")
	}

	if len(ac.DoorCloseStateTimeoutDuration) == 0 {
		return fmt.Errorf("configuration DoorCloseStateTimeoutDuration is empty")
	}

	if len(ac.DoorOpenStateTimeoutDuration) == 0 {
		return fmt.Errorf("configuration DoorOpenStateTimeoutDuration is empty")
	}

	if len(ac.InferenceDoorStatusCmd) == 0 {
		return fmt.Errorf("configuration InferenceDoorStatusCmd is empty")
	}

	if len(ac.InferenceHeartbeatCmd) == 0 {
		return fmt.Errorf("configuration InferenceHeartbeatCmd is empty")
	}

	if len(ac.InferenceTimeoutDuration) == 0 {
		return fmt.Errorf("configuration InferenceTimeoutDuration is empty")
	}

	if len(ac.InventoryAuditLogService) == 0 {
		return fmt.Errorf("configuration InventoryAuditLogService is empty")
	}

	if len(ac.InventoryService) == 0 {
		return fmt.Errorf("configuration InventoryService is empty")
	}

	if ac.LCDRowLength == 0 {
		return fmt.Errorf("configuration LCDRowLength is set to 0")
	}

	if len(ac.LedgerService) == 0 {
		return fmt.Errorf("configuration LedgerService is empty")
	}

	return nil
}
