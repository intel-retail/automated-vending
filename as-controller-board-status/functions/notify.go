// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
)

// SubscribeToNotificationService configures an email notification and submits
// it to the EdgeX notification service
func (boardStatus CheckBoardStatus) SubscribeToNotificationService() error {
	dto := dtos.Subscription{
		Id:   uuid.NewString(),
		Name: boardStatus.Configuration.NotificationName,
		Channels: []dtos.Address{
			{
				Type:         "EMAIL",
				EmailAddress: dtos.EmailAddress{Recipients: boardStatus.notificationEmailAddresses},
			},
		},
		Receiver: boardStatus.Configuration.NotificationReceiver,
		Labels: []string{
			boardStatus.Configuration.NotificationCategory,
		},
		Categories: []string{
			boardStatus.Configuration.NotificationCategory,
		},
		AdminState: boardStatus.Configuration.SubscriptionAdminState,
	}
	reqs := []requests.AddSubscriptionRequest{requests.NewAddSubscriptionRequest(dto)}
	_, err := boardStatus.SubscriptionClient.Add(context.Background(), reqs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to the EdgeX notification service: %s", err.Error())
	}

	return nil
}

func (boardStatus CheckBoardStatus) SendNotification(message string) error {
	dto := dtos.NewNotification(boardStatus.notificationLabels,
		boardStatus.Configuration.NotificationCategory,
		message,
		boardStatus.Configuration.NotificationSender,
		boardStatus.Configuration.NotificationSeverity,
	)

	req := requests.NewAddNotificationRequest(dto)
	_, err := boardStatus.NotificationClient.SendNotification(context.Background(), []requests.AddNotificationRequest{req})
	if err != nil {
		return fmt.Errorf("failed to send the notification: %s", err.Error())
	}

	return nil
}
