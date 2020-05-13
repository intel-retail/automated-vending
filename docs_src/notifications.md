# How to Send Notifications through EdgeX (optional)

This section provides instructions to help you configure the EdgeX notifications service to send alerts through SMS, email, REST API calls, and others.

Notifications work as follows:

1. When either the minimum or maximum temperature thresholds (defined in `as-controller-board-status`) have been exceeded (calculated as an average temperature over a configurable duration), the service enters maintenance mode and begins the process of sending an alert

2. The `as-controller-board-status` service sends these alerts as email messages through the EdgeX notification service using REST API calls

To change the message type from email to a different medium, the `as-controller-board-status` service should be updated to use a different notification type.

## Step 1: Set Environment Variables

Set environment variable overrides for `Smtp_Host` and `Smtp_Port` in 'config-seed', which will inject these variables into the notification service's registry.

Additional notification service configuration properties are [here](https://fuji-docs.edgexfoundry.org/Ch-AlertsNotifications.html#configuration-properties "EdgeX Alerts & Notifications").

## Step 2: Add code to the config-seed Environment Section

The code snippet below is a docker-compose example that sends an email notification. Add this code to the `config-seed` service's environment section in `docker-compose.yml`.

``` yaml
environment:
  <<: *common-variables
  Smtp_Host: <host name>
  Smtp_Port: 25
  Smtp_Password: <password if applicable>
  Smtp_Sender: <some email>
  Smtp_Subject: Automated Checkout Notification
```

## Step 3: Add SMTP Server to compose file (optional)

The snipped below adds a development SMTP server smtp4dev to your `docker-compose.yml`.
Skip this step if you want to use Gmail or another server.

``` yaml
smtp-server:
  image: rnwood/smtp4dev:linux-amd64-v3
  ports:
    - "3000:80"
    - "2525:25"
  restart: "on-failure:5"
  container_name: smtp-server
  networks:
    - automated-checkout_default # the name of this network may be different for your setup
```
