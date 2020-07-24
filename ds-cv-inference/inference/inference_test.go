package inference

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gocv.io/x/gocv"
)

func TestCalculateDelta(t *testing.T) {

	originalCount := map[string]int{"sprite": 2, "gatorade": 2, "pringles": 1}
	afterCount := map[string]int{"sprite": 1, "gatorade": 1, "ruffles": 2}
	skuMapping := map[string]string{"sprite": "4900002470", "gatorade": "4900002510", "pringles": "4900002525", "ruffles": "4900002520"}
	expectedDelta := []deltaValue{
		{SKU: "4900002525", Delta: -1},
		{SKU: "4900002470", Delta: -1},
		{SKU: "4900002510", Delta: -1},
		{SKU: "4900002520", Delta: 2},
	}

	actualDelta := calculateDelta(originalCount, afterCount, skuMapping)

	fmt.Println(expectedDelta)
	fmt.Println(actualDelta)

	require.ElementsMatch(t, expectedDelta, actualDelta)

}

func TestCountImagesNegative(t *testing.T) {

	_, err := countImages("")
	require.Error(t, err, "Expecting countImages() to fail.")

}

func TestConstructImageSequence(t *testing.T) {

	imageDirectory := "../images"

	// Count number of images in the directory
	count, err := countImages(imageDirectory)
	require.NoError(t, err, "Unable to count images in directory %v", imageDirectory)

	// Construct Ring
	imageRing, err := constructImageSequence(imageDirectory)
	require.NoError(t, err, "Error constructing image sequence ring", imageDirectory)

	require.Equal(t, imageRing.Len(), count, "image ring length (%d) is different than image count (%d)", imageRing.Len(), count)

}

func TestNetDetections(t *testing.T) {

	model := "../product-detection-0001/FP32/product-detection-0001.bin"
	config := "../product-detection-0001/FP32/product-detection-0001.xml"
	image := "image_test.jpg"

	// read openVINO product detection model
	net := gocv.ReadNet(model, config)
	require.Equal(t, net.Empty(), false, "Error reading network model from : %v %v\n", model, config)
	defer net.Close()

	// OpenVINO backend
	err := net.SetPreferableBackend(gocv.NetBackendOpenVINO)
	require.NoError(t, err, "Unable to set Net backend: %v\n", gocv.NetBackendOpenVINO)

	err = net.SetPreferableTarget(gocv.NetTargetCPU)
	require.NoError(t, err, "Unable to set Prefereable target: %v\n", gocv.NetTargetCPU)

	img, detections := netDetections(image, net)
	defer detections.Close()

	elements := 0
	for i := 0; i < detections.Total(); i += 7 {
		confidence := detections.GetFloatAt(0, i+2)
		if confidence > 0.85 {
			elements++
		}
	}

	require.Equal(t, elements, 7, "Number of actual detections (%d) doesn't match with expected detections (7)", elements)

	actualDict := make(map[string]int)
	performDetection(&img, detections, actualDict, 0.85)
	expectedDict := map[string]int{"sprite": 2, "ruffles": 1, "gatorade": 2, "pringles": 2}

	require.Equal(t, expectedDict, actualDict, "actualDict doesn't match with expectedDict")

}
