// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"ds-cv-inference/inference"
	"ds-cv-inference/mjpeg"
	"ds-cv-inference/mqtt"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gocv.io/x/gocv"
)

func main() {

	// Flags
	directory := flag.String("dir", "./images", "Images directory.")
	mqttAddress := flag.String("mqtt", "localhost:1883", "Mqtt address.")
	model := flag.String("model", "product-detection-0001/FP32/product-detection-0001.bin", "Model file path.")
	configFile := flag.String("config", "product-detection-0001/FP32/product-detection-0001.xml", "XML model config file path.")
	confidence := flag.Float64("confidence", 0.85, "Confidence threshold.")
	skuMapping := flag.String("skuMapping", "skumapping.json", "SKU Mapping JSON file path")

	flag.Parse()

	//Read skuMappingJSON
	skuMappingJSONFile, err := os.Open(*skuMapping)
	if err != nil {
		fmt.Printf("Error reading from SKU Mapping file: %v\n", *skuMapping)
		os.Exit(1)
	}
	defer skuMappingJSONFile.Close()

	skuMappingJSONByte, _ := io.ReadAll(skuMappingJSONFile)

	inferenceInit(*directory, *model, *configFile, *confidence, skuMappingJSONByte)

	mqttConnection := mqtt.NewMqttConnection(*mqttAddress)
	mqttConnection.SubscribeToAutomatedVending()
	defer mqttConnection.Disconnect()

	inference.Stream = mjpeg.NewStream()
	inference.StreamChannel = make(chan gocv.Mat)

	go updateMjpegServer()

	http.Handle("/", inference.Stream)
	log.Fatal(http.ListenAndServe(":9005", nil))
}

func updateMjpegServer() {

	for img := range inference.StreamChannel {
		buf, err := gocv.IMEncode(".jpg", img)
		if err != nil {
			fmt.Println("error on IMEncode JPG with image: ", err.Error())
			os.Exit(1)
		}
		defer buf.Close()

		inference.Stream.UpdateJPEG(buf.GetBytes())

	}
}

// initialize inference config variables
func inferenceInit(directory string, model string, config string, confidence float64, skuMappingJSONByte []byte) {
	inference.AppConfig.Directory = directory
	inference.AppConfig.Model = model
	inference.AppConfig.ConfigFile = config
	inference.AppConfig.Confidence = confidence

	//Unmarshal JSON into map
	skuMap := make(map[string]string)
	if err := json.Unmarshal(skuMappingJSONByte, &skuMap); err != nil {
		fmt.Println("Error Unmarshaling sku mapping json file")
		return
	}

	inference.AppConfig.SKUMapping = skuMap
}
