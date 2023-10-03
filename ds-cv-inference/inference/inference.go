// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package inference

import (
	"container/ring"
	"ds-cv-inference/mjpeg"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"

	"gocv.io/x/gocv"
)

const font = gocv.FontHersheySimplex

var (
	// Products
	labels = []string{
		"background_label",
		"undefined",
		"sprite",
		"kool-aid",
		"extra",
		"ocelo",
		"finish",
		"mtn_dew",
		"best_foods",
		"gatorade",
		"heinz",
		"ruffles",
		"pringles",
		"del_monte",
	}

	white  = color.RGBA{255, 250, 250, 0}
	blue   = color.RGBA{36, 122, 208, 0}
	purple = color.RGBA{145, 61, 136, 1}
	green  = color.RGBA{30, 130, 76, 1}
	red    = color.RGBA{231, 76, 60, 1}

	colorsMap = map[string]color.RGBA{
		"pringles": purple,
		"sprite":   green,
		"gatorade": red,
		"ruffles":  blue,
	}

	// Stream hold a state of the mjpeg streamer
	Stream *mjpeg.Stream

	// StreamChannel used to send a gocv.Mat to the streamer
	StreamChannel chan gocv.Mat
)

// AppConfig holds configuration variables
var AppConfig config

type config struct {
	Directory  string
	Model      string
	ConfigFile string
	Confidence float64
	SKUMapping map[string]string
}

type deltaValue struct {
	SKU   string `json:"SKU"`
	Delta int    `json:"delta"`
}

// StartInference func loads a DNN model and starts the inference
func StartInference(inferenceDeltasChannel chan []byte, InferenceDoorOpenChannel chan bool) {

	directory := AppConfig.Directory
	model := AppConfig.Model
	config := AppConfig.ConfigFile
	confidenceThreshold := AppConfig.Confidence
	skuMap := AppConfig.SKUMapping

	imageSequence, err := constructImageSequence(directory)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", directory)
		return
	}

	// read openVINO product detection model
	net := gocv.ReadNet(model, config)
	if net.Empty() {
		fmt.Printf("Error reading network model from : %v %v\n", model, config)
		return
	}
	defer net.Close()

	// OpenVINO backend
	if err := net.SetPreferableBackend(gocv.NetBackendOpenVINO); err != nil {
		fmt.Printf("Unable to set Net backend: %v\n", gocv.NetBackendOpenVINO)
		return
	}

	if err := net.SetPreferableTarget(gocv.NetTargetCPU); err != nil {
		fmt.Printf("Unable to set Prefereable target: %v\n", gocv.NetTargetCPU)
		return
	}

	fmt.Println("Start reading...")
	fromImages(imageSequence, net, inferenceDeltasChannel, InferenceDoorOpenChannel, confidenceThreshold, skuMap)

}

func fromImages(imageSequence *ring.Ring, net gocv.Net, inferenceDeltasChannel chan []byte, InferenceDoorOpenChannel chan bool, confidence float64, skuMap map[string]string) {

	doorStatus := true
	dict := make(map[string]int)

INFERENCE:

	img, detections := netDetections(fmt.Sprintf("%v", imageSequence.Value), net)
	defer detections.Close()

	// Execute this section when door opens
	if doorStatus {
		dict = make(map[string]int)
		performDetection(&img, detections, dict, confidence)
	}

	// Execute this section when door closes
	if !doorStatus {
		fmt.Println("2 Door is closed !!!")

		newDict := make(map[string]int)
		performDetection(&img, detections, newDict, confidence)
		calculatedDelta := calculateDelta(dict, newDict, skuMap)
		fmt.Print("calculated delta: ")
		fmt.Println(calculatedDelta)
		if deltaBytes, err := json.Marshal(calculatedDelta); err == nil {
			inferenceDeltasChannel <- deltaBytes
		}
	}

	for {
		select {
		case doorStatus = <-InferenceDoorOpenChannel:
			if !doorStatus {
				imageSequence = imageSequence.Next()
			}
			goto INFERENCE
		default:
			// send image to mjpeg streamer
			StreamChannel <- img
		}
	}
}

func netDetections(imagePath string, net gocv.Net) (gocv.Mat, gocv.Mat) {
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		fmt.Printf("Error reading image from: %v\n", imagePath)
		return gocv.Mat{}, gocv.Mat{}
	}

	gocv.Resize(img, &img, image.Pt(900, 700), 0, 0, gocv.InterpolationCubic)

	blob := gocv.BlobFromImage(img, 1.0, image.Pt(900, 700), gocv.NewScalar(0, 0, 0, 0), true, false)
	defer blob.Close()

	// feed the blob into the detector
	net.SetInput(blob, "")

	// run a forward pass thru the network
	detBlob := net.Forward("")
	defer detBlob.Close()

	detections := gocv.GetBlobChannel(detBlob, 0, 0)

	return img, detections
}

// performDetection analyzes the results from the detector network,
// which produces an output blob with a shape 1x1xNx7
// where N is the number of detections, and each detection
// is a vector of float values
// [image_id, label, conf, x_min, y_min, x_max, y_max]
func performDetection(frame *gocv.Mat, results gocv.Mat, dict map[string]int, confidenceThreshold float64) {

	for i := 0; i < results.Total(); i += 7 {
		confidence := results.GetFloatAt(0, i+2)
		if confidence > float32(confidenceThreshold) {

			// Draw rectangle
			labelID := results.GetFloatAt(0, i+1)
			labelName := labels[int(labelID)]
			xMin := int(results.GetFloatAt(0, i+3) * float32(frame.Cols()))
			yMin := int(results.GetFloatAt(0, i+4) * float32(frame.Rows()))
			xMax := int(results.GetFloatAt(0, i+5) * float32(frame.Cols()))
			yMax := int(results.GetFloatAt(0, i+6) * float32(frame.Rows()))
			rect := image.Rect(xMin, yMin, xMax, yMax)
			gocv.Rectangle(frame, rect, colorsMap[labelName], 2)

			// Draw label
			pt := image.Pt(rect.Min.X, rect.Min.Y-10)
			textSize := gocv.GetTextSize(labelName, font, 0.6, 1)
			textRect := image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+textSize.X+2, rect.Min.Y-10-textSize.Y-2)
			gocv.Rectangle(frame, textRect, colorsMap[labelName], -1)
			gocv.PutText(frame, labelName, pt, font, 0.6, white, 1)

			//populate inference if label not in map
			key := labelName
			populateDict(dict, key)

		}
	}
}

func populateDict(dict map[string]int, key string) {

	if _, checkKeyInDict := dict[key]; !checkKeyInDict {
		dict[key] = 1
	} else {
		dict[key] = dict[key] + 1
	}
	fmt.Println("PopulateDict ", dict)
}

func calculateDelta(originalCount map[string]int, afterCount map[string]int, skuMap map[string]string) []deltaValue {
	calculatedDeltas := []deltaValue{}

	for key, value := range originalCount {
		mappedSkuKey := skuMap[key]
		if afterValue, ok := afterCount[key]; ok {
			if value == afterCount[key] {
				continue
			} else {
				newDelta := deltaValue{
					SKU:   mappedSkuKey,
					Delta: afterValue - value,
				}
				calculatedDeltas = append(calculatedDeltas, newDelta)
			}
		} else {
			newDelta := deltaValue{
				SKU:   mappedSkuKey,
				Delta: value * -1,
			}
			calculatedDeltas = append(calculatedDeltas, newDelta)
		}
	}

	for key, value := range afterCount {
		mappedSkuKey := skuMap[key]
		if _, ok := originalCount[key]; !ok {

			newDelta := deltaValue{
				SKU:   mappedSkuKey,
				Delta: value,
			}
			calculatedDeltas = append(calculatedDeltas, newDelta)
		}
	}

	return calculatedDeltas
}

func countImages(directory string) (int, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		fmt.Printf("Error opening directory: %v\n", directory)
		return 0, err
	}
	return len(files), nil
}

func constructImageSequence(directory string) (*ring.Ring, error) {

	// Construct image sequence
	imageCount, err := countImages(directory)
	if err != nil {
		return nil, err
	}
	imageSequence := ring.New(imageCount)

	// Initialize Ring
	for i := 0; i < imageCount; i++ {
		imageSequence.Value = fmt.Sprintf("%s/%d.jpg", directory, i)
		imageSequence = imageSequence.Next()
	}

	return imageSequence, nil
}
