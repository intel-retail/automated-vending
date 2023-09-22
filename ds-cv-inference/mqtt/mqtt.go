// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package mqtt

import (
	"ds-cv-inference/inference"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	commandTopic  = "Inference/CommandTopic"
	responseTopic = "Inference/ResponseTopic"
	dataTopic     = "Inference/DataTopic"
	retryCount    = 5
	waitTime      = 1 * time.Second
)

var (
	doorStatusLock           DoorStatusLock
	inferenceDoorOpenChannel chan bool
	inferenceDeltasChannel   chan []byte
)

// DoorStatusLock holds the Door state (open|close)
type DoorStatusLock struct {
	mu                  sync.Mutex
	InferenceDoorStatus bool
}

// Connection holds the mqtt client interface
type Connection struct {
	MqttClient MQTT.Client
}

// NewMqttConnection starts a new mqtt connection
func NewMqttConnection(connectionString string) Connection {
	mc := Connection{}
	mc.Connect(connectionString)
	//Set Default Status for door status
	doorStatusLock.InferenceDoorStatus = false
	inferenceDeltasChannel = make(chan []byte)
	inferenceDoorOpenChannel = make(chan bool)

	// Starts inference
	go inference.StartInference(inferenceDeltasChannel, inferenceDoorOpenChannel)

	fmt.Println("New MQTT Connection Created")
	return mc
}

// define a function for the default message handler
var commandTopicFunction MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {

	var edgeXMessage map[string]string
	if err := json.Unmarshal(msg.Payload(), &edgeXMessage); err != nil {
		fmt.Println(http.StatusBadRequest, "Failed to unmarshal body")
		return
	}

	fmt.Printf("received message: %v+", edgeXMessage)

	words := strings.Split(msg.Topic(), "/")
	if len(words) != 5 {
		fmt.Println(http.StatusBadRequest, fmt.Sprintf("mqtt command topic not formatted for EdgeX 3.0: %s", msg.Topic()))
		return
	}
	cmd := words[2]
	uuid := words[4]
	publishTopic := fmt.Sprintf("%s/%s", responseTopic, uuid)

	switch cmd {
	case "inferenceHeartbeat":
		{
			pingMessage := edgeXMessage
			pingMessage["inferenceHeartbeat"] = "inferencePong"

			pongMessage, err := json.Marshal(pingMessage)
			if err != nil {
				fmt.Println("Failed to marshal mqtt message")
			}
			token := client.Publish(publishTopic, 0, false, pongMessage)
			token.Wait()
		}
	case "inferenceDoorStatus":
		{
			pingMessage := edgeXMessage
			isDoorClosed := pingMessage["inferenceDoorStatus"]
			pingMessage["inferenceDoorStatus"] = "Got it!"

			pongMessage, err := json.Marshal(pingMessage)
			if err != nil {
				fmt.Println("Failed to marshal mqtt message")
			}
			token := client.Publish(publishTopic, 0, false, pongMessage)
			token.Wait()
			checkDoorStatus(isDoorClosed, client)
		}
	default:
		fmt.Println("Unknown cmd " + edgeXMessage["cmd"])
	}
}

func checkDoorStatus(isDoorClosed string, client MQTT.Client) {

	if isDoorClosed == "false" {
		//Check if Door is already open or not
		if !doorStatusLock.InferenceDoorStatus {
			//lock check for shared variable InferenceDoorStatus
			doorStatusLock.mu.Lock()
			doorStatusLock.InferenceDoorStatus = true
			doorStatusLock.mu.Unlock()
		} else {
			fmt.Println("Door is already open !!!")
			return
		}

		fmt.Println("Door is open !!!")
		// Send open signal to inference
		inferenceDoorOpenChannel <- true

	} else if isDoorClosed == "true" {

		//lock check for shared variable InferenceDoorStatus
		doorStatusLock.mu.Lock()
		doorStatusLock.InferenceDoorStatus = false
		doorStatusLock.mu.Unlock()

		// Send false signal to inference
		inferenceDoorOpenChannel <- false

		fmt.Println("1 Door is closed !!!")
		// Receive delta data from inference
		delta := <-inferenceDeltasChannel
		if len(delta) != 0 {
			SendDeltaData(client, delta)
		}
	}

}

// SendDeltaData publishes the delta data back to mqtt broker
func SendDeltaData(client MQTT.Client, delta []byte) {

	cmdSKUDelta := "inferenceSkuDelta"
	publishTopic := fmt.Sprintf("%s/%s/%s", dataTopic, "Inference-device", cmdSKUDelta)
	edgeXMessage := make(map[string]string)
	edgeXMessage["method"] = "get"
	edgeXMessage[cmdSKUDelta] = string(delta)

	deltaMessage, _ := json.Marshal(edgeXMessage)
	fmt.Println("Final deltaMessage is ", string(deltaMessage))
	token := client.Publish(publishTopic, 0, false, deltaMessage)
	token.Wait()
}

func (mqttCon *Connection) Connect(connectionString string) {
	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := MQTT.NewClientOptions().AddBroker(connectionString)
	opts.SetClientID("ds-cv-inference")
	//create and start a client using the above ClientOptions
	mqttCon.MqttClient = MQTT.NewClient(opts)
	attempts := 0
	fmt.Println("Attempting to connect to mqtt broker at " + connectionString)
	for attempts < retryCount {
		if token := mqttCon.MqttClient.Connect(); token.Wait() && token.Error() == nil {
			break
		}
		time.Sleep(waitTime)
		fmt.Println("Failed to connect to mqtt broker")
		attempts++
	}
}

func (mqttCon *Connection) SubscribeToAutomatedVending() {
	//subscribe to the topic inference/CommandTopic and handle messages in the commandTopicFunction
	attempts := 0
	for attempts < retryCount {
		if token := mqttCon.MqttClient.Subscribe(commandTopic, 0, commandTopicFunction); token.Wait() && token.Error() == nil {
			fmt.Println("subscribe successful")
			break
		}
		time.Sleep(waitTime)
		fmt.Println("Failed to subscribe to channel")
		attempts++
	}
	fmt.Println("---done subscribing to " + commandTopic)
}

func (mqttCon *Connection) Subscribe(channel string, recvMesg func(client MQTT.Client, msg MQTT.Message)) {
	//subscribe to the channel and request messages to be delivered
	//at a maximum qos of zero, wait for the receipt to confirm the subscription
	if token := mqttCon.MqttClient.Subscribe(channel, 0, recvMesg); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

func (mqttCon *Connection) Unsubscribe(channel string) {
	//unsubscribe from channel sample
	if token := mqttCon.MqttClient.Unsubscribe(channel); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

func (mqttCon *Connection) Disconnect() {
	mqttCon.MqttClient.Disconnect(250)
}
