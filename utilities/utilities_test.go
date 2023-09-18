// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package utilities

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	assert "github.com/stretchr/testify/assert"
)

// TestUUIDGeneration tests the GenUUID function from the helpers package
func TestUUIDGeneration(t *testing.T) {
	t.Run("GenUUID test", func(t *testing.T) {
		testUUID := GenUUID()
		_, err := uuid.Parse(testUUID)
		if err != nil {
			t.Errorf("GenUUID failed to generate a valid UUID")
		}
	})
}

// TestValidUUID tests the ValidUUID function from the helpers package
func TestValidUUID(t *testing.T) {
	t.Run("ValidUUID test", func(t *testing.T) {
		testUUID := "c1ecd6ae-d750-451d-9070-e592cf949cd7"
		isValidUUID := ValidUUID(testUUID)
		if isValidUUID != true {
			t.Errorf("ValidUUID failed to validate a valid UUID")
		}
	})
}

// TestWriteToJSONFile will write a string
func TestWriteToJSONFile(t *testing.T) {
	t.Run("WriteToJSONFile test", func(t *testing.T) {
		filename := "TestWriteToJSONFile.json"
		JSONContents := `{"validJson":true,"extremelyValidJson":"very true"}`

		// The JSONContents string will have backslash escapes once it gets written
		// to a file, which is expected behavior. We have to emulate it here
		marshaledJSONContentsTestString, err := json.Marshal(JSONContents)
		if err != nil {
			t.Errorf("Failed to marshal the JSONContents test string")
		}

		// Make the call to write the JSONContents string to file
		err = WriteToJSONFile(filename, JSONContents, 0644)
		if err != nil {
			t.Errorf("An error occurred with valid inputs")
		}

		// test by loading the file
		dat, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Errorf("Failed to read from the recently written filename " + filename)
		}

		// test by validating the file contents match our test string
		if string(marshaledJSONContentsTestString) != string(dat) {
			t.Errorf("The contents of the recently written file " + filename + " did not match what was expected:\n" + string(marshaledJSONContentsTestString) + "\n" + string(dat))
		}
		// clean up our test
		err = os.Remove(filename)
		if err != nil {
			t.Errorf("Failed to clean up the test file " + filename)
		}
	})
	t.Run("WriteToJSONFile test invalid GetAsJSON call", func(t *testing.T) {
		filename := "TestWriteToJSONFile_invalid.json"
		// Test calling the GetAsJSON function
		testInterface := map[string]interface{}{
			"invalidChan": make(chan int),
		}
		err := WriteToJSONFile(filename, testInterface, 0644)
		if err == nil {
			t.Errorf("WriteToJSONFile GetAsJSON call did not throw error for invalid input when it should have")
		}

		if strings.Contains(err.Error(), "Failed to marshal into JSON string") == false {
			t.Errorf("WriteToJSONFile GetAsJSON call did not yield proper error response for invalid JSON input")
		}
	})
	t.Run("WriteToJSONFile test invalid permissions", func(t *testing.T) {
		invalidJSONFilename := "./asdf/asdf/"
		// Test calling the GetAsJSON function
		testInterface := map[string]interface{}{
			"validValue": "0",
		}
		err := WriteToJSONFile(invalidJSONFilename, testInterface, 0644)
		if err == nil {
			t.Errorf("WriteToJSONFile did not throw error when passed invalid filename")
		}

		if strings.Contains(err.Error(), "Failed to write JSON to file "+invalidJSONFilename) == false {
			t.Errorf("WriteToJSONFile invalid filename error message was not as expected")
		}
	})
}

type TestJSONFileStruct struct {
	ValidJSON          bool   `json:"validJson"`
	ExtremelyValidJSON string `json:"extremelyValidJson"`
}

// TestLoadFromJSONFile will write a known JSON string to a
// JSON file and read from it.
func TestLoadFromJSONFile(t *testing.T) {
	t.Run("LoadFromJSONFile test", func(t *testing.T) {
		filename := "TestLoadFromJSONFile.json"
		JSONContents := `{"validJson":true,"extremelyValidJson":"very true"}`
		// write to a JSON file using JSON package
		err := ioutil.WriteFile(filename, []byte(JSONContents), 0644)
		if err != nil {
			t.Errorf("Failed to write to test file " + filename)
		}

		// read from the file using our function
		var JSONContentsFromFile TestJSONFileStruct
		err = LoadFromJSONFile(filename, &JSONContentsFromFile)
		if err != nil {
			t.Errorf("Failed to call the LoadFromJSONFile function: " + err.Error())
		}

		// convert JSONContentsFromFile to string
		JSONContentsFromFileStr, err := json.Marshal(JSONContentsFromFile)
		if err != nil {
			t.Errorf("Failed to marshal JSONContentsFromFile into JSONContentsFromFileStr")
		}

		// compare the two strings to see if they match
		if string(JSONContentsFromFileStr) != JSONContents {
			t.Errorf("LoadFromJSONFile failed to match expected test string:\nJSONContents: " + JSONContents + "\nJSONContentsFromFile: " + string(JSONContentsFromFileStr))
		}
		// clean up our test
		err = os.Remove(filename)
		if err != nil {
			t.Errorf("Failed to clean up the test file " + filename)
		}
	})
	t.Run("LoadFromJSONFile test invalid filename", func(t *testing.T) {
		nonExistentFilename := "TestLoadFromJSONFile_not_existing.json"

		// read from the file using our function
		var JSONContentsFromFile TestJSONFileStruct
		err := LoadFromJSONFile(nonExistentFilename, &JSONContentsFromFile)
		if err == nil {
			t.Errorf("LoadFromJSONFile expected error condition upon entering invalid filename")
		}
		if strings.Contains(err.Error(), "Failed to read from JSON file "+nonExistentFilename) == false {
			t.Errorf("LoadFromJSONFile expected proper error message upon entering invalid filename")
		}
	})
	t.Run("LoadFromJSONFile test invalid unmarshal", func(t *testing.T) {
		filename := "TestLoadFromInvalidJSONFile.json"
		invalidJSONContents := `{"validJson":false,"extremelyValidJson":no`
		// write to a JSON file using JSON package
		err := ioutil.WriteFile(filename, []byte(invalidJSONContents), 0644)
		if err != nil {
			t.Errorf("Failed to write to test file " + filename)
		}

		// read from the file using our function
		var JSONContentsFromFile TestJSONFileStruct
		err = LoadFromJSONFile(filename, &JSONContentsFromFile)
		if err == nil {
			t.Errorf("LoadFromJSONFile expected error condition upon reading invalid JSON file")
		}
		if strings.Contains(err.Error(), "Failed to unmarshal from JSON file "+filename) == false {
			t.Errorf("LoadFromJSONFile expected proper error message upon reading invalid JSON file")
		}
		// clean up our test
		err = os.Remove(filename)
		if err != nil {
			t.Errorf("Failed to clean up the test file " + filename)
		}
	})
}

// TestGetAsJSON will compare an interface with an expected JSON
// version of that interface.
func TestGetAsJSON(t *testing.T) {
	t.Run("GetAsJSON test", func(t *testing.T) {
		// set up the map to test
		inputInterface := make(map[string]interface{})
		inputInterface["string1"] = "test"
		inputInterface["bool1"] = false
		inputInterface["num1"] = 0.29
		inputInterface["interface1"] = make(map[string]interface{})

		// set up the JSON str to test
		// Note that keys must be alphabetical
		testJSONStr := `{"bool1":false,"interface1":{},"num1":0.29,"string1":"test"}`

		// Test calling the GetAsJSON function
		testedJSONStr, err := GetAsJSON(inputInterface)
		if err != nil {
			t.Errorf("GetAsJSON call failed: " + err.Error())
		}

		// Perform the match test
		if testedJSONStr != testJSONStr {
			t.Errorf("GetAsJSON failed to match test strings:\n" + testedJSONStr + "\n" + testJSONStr)
		}
	})
	t.Run("GetAsJSON test invalid input", func(t *testing.T) {
		// Test calling the GetAsJSON function
		testInterface := map[string]interface{}{
			"invalidChan": make(chan int),
		}
		_, err := GetAsJSON(testInterface)
		if err == nil {
			t.Errorf("GetAsJSON call did not throw error for invalid input when it should have")
		}

		if strings.Contains(err.Error(), "Failed to marshal into JSON string") == false {
			t.Errorf("GetAsJSON call did not yield proper error response for invalid JSON input")
		}
	})
}

// TestGetHTTPResponseTemplate will test to see if the HTTPRespone struct is returned
// with expected values.
func TestGetHTTPResponseTemplate(t *testing.T) {
	t.Run("GetHTTPResponseTemplate test", func(t *testing.T) {
		testHTTPResponse := GetHTTPResponseTemplate()

		// construct a new HTTPResponse struct
		newHTTPResponse := HTTPResponse{
			Content:     "",
			ContentType: "string",
			StatusCode:  http.StatusOK,
			Error:       false,
		}

		if testHTTPResponse.Content != newHTTPResponse.Content {
			t.Errorf("GetHTTPResponseTemplate yielded an unexpected Content property")
		}
		if testHTTPResponse.ContentType != newHTTPResponse.ContentType {
			t.Errorf("GetHTTPResponseTemplate yielded an unexpected ContentType property")
		}
		if testHTTPResponse.StatusCode != newHTTPResponse.StatusCode {
			t.Errorf("GetHTTPResponseTemplate yielded an unexpected StatusCode property")
		}
		if testHTTPResponse.Error != newHTTPResponse.Error {
			t.Errorf("GetHTTPResponseTemplate yielded an unexpected Error property")
		}
	})
}

// TestGetAsString will test to see if the GetAsString function
// returns a JSON string of an HTTPResponse
func TestGetAsString(t *testing.T) {
	t.Run("GetAsString test", func(t *testing.T) {
		testHTTPResponse := GetHTTPResponseTemplate()
		testHTTPResponse.StatusCode = http.StatusUnauthorized
		testHTTPResponse.ContentType = "json"
		testHTTPResponse.Content = `{"foo":"bar"}`
		testHTTPResponse.Error = true

		// call the function
		testJSONStr := testHTTPResponse.GetAsString()

		// define the expected JSON string
		expectedJSONStr := `{"content":"{\"foo\":\"bar\"}","contentType":"json","statusCode":401,"error":true}`

		// compare the expected & test results
		if expectedJSONStr != testJSONStr {
			t.Errorf("GetAsString failed to match expected and test strings:\n" + expectedJSONStr + "\n" + testJSONStr)
		}
	})
	t.Run("GetAsString test invalid input", func(t *testing.T) {
		testHTTPResponse := GetHTTPResponseTemplate()
		testHTTPResponse.StatusCode = http.StatusUnauthorized
		testHTTPResponse.ContentType = "json"
		testHTTPResponse.Content = make(chan int) // this cannot be marshaled
		testHTTPResponse.Error = true

		// call the function
		testJSONStr := testHTTPResponse.GetAsString()

		// define the expected JSON string
		expectedJSONStr := `{"content":"Failed to process response.","contentType":"string","statusCode":500,"error":true}`

		// compare the expected & test results
		if expectedJSONStr != testJSONStr {
			t.Errorf("GetAsString failed to match expected and test strings:\n" + expectedJSONStr + "\n" + testJSONStr)
		}
	})
}

// TestSetHTTPResponseFields will test whether the function properly sets each input
// to the expected field in the provided HTTPResponse struct.
func TestSetHTTPResponseFields(t *testing.T) {
	t.Run("SetHTTPResponseFields test", func(t *testing.T) {
		testHTTPResponse := GetHTTPResponseTemplate()

		// test the function call
		testHTTPResponse.SetHTTPResponseFields(http.StatusUnauthorized, "foo", "string", true)

		if testHTTPResponse.StatusCode != http.StatusUnauthorized {
			t.Errorf("SetHTTPResponseFields failed to set property StatusCode")
		}
		if testHTTPResponse.Content != "foo" {
			t.Errorf("SetHTTPResponseFields failed to set property Content")
		}
		if testHTTPResponse.ContentType != "string" {
			t.Errorf("SetHTTPResponseFields failed to set property ContentType")
		}
		if testHTTPResponse.Error != true {
			t.Errorf("SetHTTPResponseFields failed to set property Error")
		}
	})
}

// TestSetJSONHTTPResponseFields will test whether the function properly sets each input
// to the expected field in the provided HTTPResponse struct.
func TestSetJSONHTTPResponseFields(t *testing.T) {
	t.Run("SetJSONHTTPResponseFields test", func(t *testing.T) {
		testHTTPResponse := GetHTTPResponseTemplate()

		// test the function call
		testHTTPResponse.SetJSONHTTPResponseFields(http.StatusUnauthorized, "foo", true)

		if testHTTPResponse.StatusCode != http.StatusUnauthorized {
			t.Errorf("SetJSONHTTPResponseFields failed to set property StatusCode")
		}
		if testHTTPResponse.Content != "foo" {
			t.Errorf("SetJSONHTTPResponseFields failed to set property Content")
		}
		if testHTTPResponse.ContentType != "json" {
			t.Errorf("SetJSONHTTPResponseFields failed to set property ContentType")
		}
		if testHTTPResponse.Error != true {
			t.Errorf("SetJSONHTTPResponseFields failed to set property Error")
		}
	})
}

// TestStringHTTPResponseFields will test whether the function properly sets each input
// to the expected field in the provided HTTPResponse struct.
func TestStringHTTPResponseFields(t *testing.T) {
	t.Run("SetStringHTTPResponseFields test", func(t *testing.T) {
		testHTTPResponse := GetHTTPResponseTemplate()

		// test the function call
		testHTTPResponse.SetStringHTTPResponseFields(http.StatusUnauthorized, "foo", true)

		if testHTTPResponse.StatusCode != http.StatusUnauthorized {
			t.Errorf("SetStringHTTPResponseFields failed to set property StatusCode")
		}
		if testHTTPResponse.Content != "foo" {
			t.Errorf("SetStringHTTPResponseFields failed to set property Content")
		}
		if testHTTPResponse.ContentType != "string" {
			t.Errorf("SetStringHTTPResponseFields failed to set property ContentType")
		}
		if testHTTPResponse.Error != true {
			t.Errorf("SetStringHTTPResponseFields failed to set property Error")
		}
	})
}

// TestProcessCORS will test whether the ProcessCORS function properly sets
// the expected headers but also properly executes the passed-in HTTPFunc.
func TestProcessCORS(t *testing.T) {
	testMethods := []string{
		"DELETE",
		"GET",
		"OPTIONS",
		"PATCH",
		"POST",
		"PUT",
	}
	for _, testMethod := range testMethods {
		testName := "TestProcessCORS test " + testMethod
		t.Run(testName, func(t *testing.T) {
			req := httptest.NewRequest(testMethod, "http://localhost:48096", nil)
			w := httptest.NewRecorder()

			// the function to test
			ProcessCORS(w, req, func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, `{"content":"{\"foo\":\"bar\"}","contentType":"json","statusCode":200,"error":true}`)
				w.WriteHeader(http.StatusOK)
			})

			resp := w.Result()
			// check the status
			if resp.StatusCode != http.StatusOK {
				t.Errorf(testName + " failed to properly set status code")
			}

			// check the headers
			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf(testName + " failed to properly set header Content-Type")
			}
			if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
				t.Errorf(testName + " failed to properly set header Access-Control-Allow-Origin")
			}
			if resp.Header.Get("Access-Control-Allow-Methods") != "POST, GET, OPTIONS, PUT, DELETE" {
				t.Errorf(testName + " failed to properly set header Access-Control-Allow-Methods")
			}
			if resp.Header.Get("Access-Control-Allow-Headers") != AllowedHeaders {
				t.Errorf(testName + " failed to properly set header Access-Control-Allow-Headers")
			}

			// no content expected for OPTIONS
			if testMethod == "OPTIONS" {
				return
			}

			body, _ := ioutil.ReadAll(resp.Body)

			// unmarshal the response into an HTTPResponse struct
			responseContent := HTTPResponse{}
			err := json.Unmarshal(body, &responseContent)
			if err != nil {
				t.Errorf(testName + " failed to properly unmarshal JSON response")
			}
		})
	}
}

// TestWriteHTTPResponse will test whether the WriteHTTPResponse function will
// correctly mutate the incoming HTTPResponse
func TestWriteHTTPResponse(t *testing.T) {
	testHTTPResponses := []HTTPResponse{
		{
			Content:     `{"foo":"bar"}`,
			ContentType: "json",
			StatusCode:  200,
			Error:       false,
		}, {
			Content:     `string text`,
			ContentType: "string",
			StatusCode:  200,
			Error:       false,
		}, {
			Content:     `{"foo2":"bar"}`,
			ContentType: "json",
			StatusCode:  500,
			Error:       true,
		}, {
			Content:     `string text 2`,
			ContentType: "string",
			StatusCode:  500,
			Error:       false,
		},
	}
	for _, testHTTPResponse := range testHTTPResponses {
		testName := "TestWriteHTTPResponse test " + testHTTPResponse.Content.(string)
		t.Run(testName, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://localhost:48096", nil)
			w := httptest.NewRecorder()

			// test the function here
			testHTTPResponse.WriteHTTPResponse(w, req)

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			// unmarshal the response into an HTTPResponse struct
			responseContent := HTTPResponse{}
			err := json.Unmarshal(body, &responseContent)
			if err != nil {
				t.Errorf(testName + " failed to properly unmarshal JSON response")
			}

			// check the status
			if resp.StatusCode != testHTTPResponse.StatusCode {
				t.Errorf(testName + " failed to properly set status code")
			}
			if responseContent.Content != testHTTPResponse.Content {
				t.Errorf(testName + " failed to properly set Content")
			}
			if responseContent.ContentType != testHTTPResponse.ContentType {
				t.Errorf(testName + " failed to properly set ContentType")
			}
			if responseContent.StatusCode != testHTTPResponse.StatusCode {
				t.Errorf(testName + " failed to properly set StatusCode")
			}
			if responseContent.Error != testHTTPResponse.Error {
				t.Errorf(testName + " failed to properly set Error")
			}
		})
	}
}

// TestWriteStringHTTPResponse will test whether the WriteStringHTTPResponse function will
// correctly create & mutate an HTTPResponse
func TestWriteStringHTTPResponse(t *testing.T) {
	testHTTPResponses := []HTTPResponse{
		{
			Content:     `{"foo":"bar"}`,
			ContentType: "json",
			StatusCode:  200,
			Error:       false,
		}, {
			Content:     `string text`,
			ContentType: "string",
			StatusCode:  200,
			Error:       false,
		}, {
			Content:     `{"foo2":"bar"}`,
			ContentType: "json",
			StatusCode:  500,
			Error:       true,
		}, {
			Content:     `string text 2`,
			ContentType: "string",
			StatusCode:  500,
			Error:       false,
		},
	}
	for _, testHTTPResponse := range testHTTPResponses {
		testNameWriteStringHTTPResponse := "TestWriteStringHTTPResponse test " + testHTTPResponse.Content.(string)
		t.Run(testNameWriteStringHTTPResponse, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://localhost:48096", nil)
			w := httptest.NewRecorder()

			// test the function here
			WriteStringHTTPResponse(w, req, testHTTPResponse.StatusCode, testHTTPResponse.Content, testHTTPResponse.Error)

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			// unmarshal the response into an HTTPResponse struct
			responseContent := HTTPResponse{}
			err := json.Unmarshal(body, &responseContent)
			if err != nil {
				t.Errorf(testNameWriteStringHTTPResponse + " failed to properly unmarshal JSON response")
			}

			// check the status
			if resp.StatusCode != testHTTPResponse.StatusCode {
				t.Errorf(testNameWriteStringHTTPResponse + " failed to properly set status code")
			}
			// validate the response mutation
			if responseContent.Content != testHTTPResponse.Content {
				t.Errorf(testNameWriteStringHTTPResponse + " failed to properly set Content")
			}
			if responseContent.ContentType != HTTPResponseTypeString {
				t.Errorf(testNameWriteStringHTTPResponse + " failed to properly set ContentType")
			}
			if responseContent.StatusCode != testHTTPResponse.StatusCode {
				t.Errorf(testNameWriteStringHTTPResponse + " failed to properly set StatusCode")
			}
			if responseContent.Error != testHTTPResponse.Error {
				t.Errorf(testNameWriteStringHTTPResponse + " failed to properly set Error")
			}
		})
	}
}

// TestWriteJSONHTTPResponse will test whether the WriteJSONHTTPResponse function will
// correctly create & mutate an HTTPResponse
func TestWriteJSONHTTPResponse(t *testing.T) {
	testHTTPResponses := []HTTPResponse{}
	testHTTPResponses = append(testHTTPResponses,
		HTTPResponse{
			Content:     `{"foo":"bar"}`,
			ContentType: "json",
			StatusCode:  200,
			Error:       false,
		}, HTTPResponse{
			Content:     `string text`,
			ContentType: "string",
			StatusCode:  200,
			Error:       false,
		}, HTTPResponse{
			Content:     `{"foo2":"bar"}`,
			ContentType: "json",
			StatusCode:  500,
			Error:       true,
		}, HTTPResponse{
			Content:     `string text 2`,
			ContentType: "string",
			StatusCode:  500,
			Error:       false,
		})
	for _, testHTTPResponse := range testHTTPResponses {
		testNameWriteJSONHTTPResponse := "TestWriteJSONHTTPResponse test " + testHTTPResponse.Content.(string)
		t.Run(testNameWriteJSONHTTPResponse, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://localhost:48096", nil)
			w := httptest.NewRecorder()

			// test the function here
			WriteJSONHTTPResponse(w, req, testHTTPResponse.StatusCode, testHTTPResponse.Content, testHTTPResponse.Error)

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			// unmarshal the response into an HTTPResponse struct
			responseContent := HTTPResponse{}
			err := json.Unmarshal(body, &responseContent)
			if err != nil {
				t.Errorf(testNameWriteJSONHTTPResponse + " failed to properly unmarshal JSON response")
			}

			// check the status
			if resp.StatusCode != testHTTPResponse.StatusCode {
				t.Errorf(testNameWriteJSONHTTPResponse + " failed to properly set status code")
			}
			// validate the response mutation
			if responseContent.Content != testHTTPResponse.Content {
				t.Errorf(testNameWriteJSONHTTPResponse + " failed to properly set Content")
			}
			if responseContent.ContentType != HTTPResponseTypeJSON {
				t.Errorf(testNameWriteJSONHTTPResponse + " failed to properly set ContentType")
			}
			if responseContent.StatusCode != testHTTPResponse.StatusCode {
				t.Errorf(testNameWriteJSONHTTPResponse + " failed to properly set StatusCode")
			}
			if responseContent.Error != testHTTPResponse.Error {
				t.Errorf(testNameWriteJSONHTTPResponse + " failed to properly set Error")
			}
		})
	}
}

// TestDateFunction will test whether the date function properly returns
// the date in the expected format.
func TestDateFunction(t *testing.T) {
	t.Run("SetDateFunction test", func(t *testing.T) {
		testDateStr := NowDate()
		_, err := time.Parse(dtFormat, testDateStr)

		if err != nil {
			t.Errorf("TestDateFunction failed to return date in expected format")
		}
	})
}

// TestStruct is a test type that is used for simple unmarshal tests
type TestStruct struct {
	Foo string `json:"foo"`
}

// UnmarshalableStruct is a test type that is used for breaking unmarshal tests
type UnmarshalableStruct struct {
	Foo2 chan int
}

// EmptyBuffer is an empty struct that is intended for testing with
// ioutil.ReadCloser to produce error conditions
// https://play.golang.org/p/ZhkyijksQ7j
type EmptyBuffer struct {
}

// Close allows ioutil.ReadCloser to call this function as part of a test case
// upon closure of the EmptyBuffer
func (buff EmptyBuffer) Close() error {
	return nil
}

// Read tells any instance of EmptyBuffer to throw an error when
// ioutil.ReadCloser attempts to read
func (buff EmptyBuffer) Read(anything []byte) (b int, err error) {
	return 0, fmt.Errorf("Injected Failure Condition")
}

// InvalidJSONBuffer is an empty struct that is intended for testing with
// ioutil.ReadCloser to produce error conditions
type InvalidJSONBuffer struct {
}

// Close allows ioutil.ReadCloser to call this function as part of a test case
// upon closure of the InvalidJSONBuffer
func (buff InvalidJSONBuffer) Close() error {
	return nil
}

// Read tells any instance of InvalidJSONBuffer to throw an "F" character
// upon read
func (buff InvalidJSONBuffer) Read(p []byte) (n int, err error) {
	if n > 0 {
		buff.Close()
	}
	return 70, io.EOF
}

// TestParseJSONHTTPResponseContent will test if the function properly
// extracts an interface from an HTTPResponse from an HTTP response (yes, you
// read that correctly)
func TestParseJSONHTTPResponseContent(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		testHTTPResponse := HTTPResponse{
			Content:     `{"foo":"bar"}`,
			ContentType: "json",
			StatusCode:  http.StatusOK,
			Error:       false,
		}
		req := httptest.NewRequest("GET", "http://localhost:48096", nil)
		w := httptest.NewRecorder()

		// test the function here
		WriteJSONHTTPResponse(w, req, testHTTPResponse.StatusCode, testHTTPResponse.Content, testHTTPResponse.Error)

		resp := w.Result()

		var testStruct = TestStruct{}

		response, err := ParseJSONHTTPResponseContent(resp.Body, &testStruct)

		if err != nil {
			t.Errorf("Error running ParseJSONHTTPResponseContent: " + err.Error())
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK, got: " + strconv.Itoa(response.StatusCode))
		}
		if response.Error != false {
			t.Errorf("Expected error=false, got something else")
		}
		if response.ContentType != HTTPResponseTypeJSON {
			t.Errorf("Expected content type=" + HTTPResponseTypeJSON + ", got " + response.ContentType)
		}

		if testStruct.Foo != "bar" {
			t.Errorf("Expected testStruct.Foo == bar, got: " + testStruct.Foo)
		}
	})
	t.Run("ioutil ReadAll Failure", func(t *testing.T) {
		blankHTTPResponse := HTTPResponse{}
		req := httptest.NewRequest("GET", "http://localhost:48096", nil)
		w := httptest.NewRecorder()

		// test the function here
		WriteJSONHTTPResponse(w, req, blankHTTPResponse.StatusCode, blankHTTPResponse.Content, blankHTTPResponse.Error)

		// inject our buffer error condition
		resp := w.Result()
		var buf EmptyBuffer
		resp.Body = buf

		var testStruct = TestStruct{}

		response, err := ParseJSONHTTPResponseContent(resp.Body, &testStruct)

		if err == nil {
			t.Errorf("Expected ioutil error: " + err.Error())
		}

		if response.StatusCode != blankHTTPResponse.StatusCode {
			t.Errorf("Expected status OK, got: " + strconv.Itoa(response.StatusCode))
		}
		if response.Error != blankHTTPResponse.Error {
			t.Errorf("Expected error=false, got something else")
		}
		if response.ContentType != blankHTTPResponse.ContentType {
			t.Errorf("Expected content type=" + HTTPResponseTypeJSON + ", got " + response.ContentType)
		}
	})
	t.Run("Invalid HTTPResponse unmarshal error", func(t *testing.T) {
		blankHTTPResponse := HTTPResponse{}
		req := httptest.NewRequest("GET", "http://localhost:48096", nil)
		w := httptest.NewRecorder()

		// test the function here
		WriteJSONHTTPResponse(w, req, blankHTTPResponse.StatusCode, blankHTTPResponse.Content, blankHTTPResponse.Error)

		// inject our buffer error condition
		resp := w.Result()
		var buf InvalidJSONBuffer
		resp.Body = buf

		var testStruct = TestStruct{}

		response, err := ParseJSONHTTPResponseContent(resp.Body, &testStruct)

		if err == nil {
			t.Errorf("Expected ioutil error: " + err.Error())
		}

		if response.StatusCode != blankHTTPResponse.StatusCode {
			t.Errorf("Expected status OK, got: " + strconv.Itoa(response.StatusCode))
		}
		if response.Error != blankHTTPResponse.Error {
			t.Errorf("Expected error=false, got something else")
		}
		if response.ContentType != blankHTTPResponse.ContentType {
			t.Errorf("Expected content type=" + HTTPResponseTypeJSON + ", got " + response.ContentType)
		}
	})
	t.Run("Invalid ContentType error", func(t *testing.T) {
		testHTTPResponse := HTTPResponse{
			Content:     "test",
			ContentType: HTTPResponseTypeString,
		}
		req := httptest.NewRequest("GET", "http://localhost:48096", nil)
		w := httptest.NewRecorder()

		// test the function here
		WriteStringHTTPResponse(w, req, testHTTPResponse.StatusCode, testHTTPResponse.Content, testHTTPResponse.Error)

		resp := w.Result()

		var testStruct = TestStruct{}

		response, err := ParseJSONHTTPResponseContent(resp.Body, &testStruct)

		if err == nil {
			t.Errorf("Expected an error error but did not get one")
		}

		if strings.Contains(err.Error(), "HTTPResponse ContentType is not "+HTTPResponseTypeJSON) != true {
			t.Errorf("Expected to receive error for invalid string ContentType: " + err.Error())
		}

		if response.StatusCode != testHTTPResponse.StatusCode {
			t.Errorf("Expected status OK, got: " + strconv.Itoa(response.StatusCode))
		}
		if response.Error != testHTTPResponse.Error {
			t.Errorf("Expected error=false, got something else")
		}
		if response.ContentType != testHTTPResponse.ContentType {
			t.Errorf("Expected content type=" + testHTTPResponse.ContentType + ", got " + response.ContentType)
		}
	})
	t.Run("Invalid Content interface assertion", func(t *testing.T) {
		testHTTPResponse := HTTPResponse{
			Content: nil,
		}
		req := httptest.NewRequest("GET", "http://localhost:48096", nil)
		w := httptest.NewRecorder()

		// test the function here
		WriteJSONHTTPResponse(w, req, testHTTPResponse.StatusCode, testHTTPResponse.Content, testHTTPResponse.Error)

		resp := w.Result()

		var testStruct = TestStruct{}

		response, err := ParseJSONHTTPResponseContent(resp.Body, &testStruct)

		if err == nil {
			t.Errorf("Expected an error error but did not get one")
		}

		if strings.Contains(err.Error(), "HTTPResponse Content must be a string, instead got <nil>") != true {
			t.Errorf("Expected to receive error for invalid Content interface type: " + err.Error())
		}

		if response.StatusCode != testHTTPResponse.StatusCode {
			t.Errorf("Expected status OK, got: " + strconv.Itoa(response.StatusCode))
		}
		if response.Error != testHTTPResponse.Error {
			t.Errorf("Expected error=false, got something else")
		}
		if response.ContentType != HTTPResponseTypeJSON {
			t.Errorf("Expected content type=" + HTTPResponseTypeJSON + ", got " + response.ContentType)
		}
	})
	t.Run("Invalid final unmarshal target", func(t *testing.T) {
		testHTTPResponse := HTTPResponse{
			Content: `{"foo2":"bar"}`,
		}
		req := httptest.NewRequest("GET", "http://localhost:48096", nil)
		w := httptest.NewRecorder()

		// test the function here
		WriteJSONHTTPResponse(w, req, testHTTPResponse.StatusCode, testHTTPResponse.Content, testHTTPResponse.Error)

		resp := w.Result()

		var testStruct UnmarshalableStruct

		response, err := ParseJSONHTTPResponseContent(resp.Body, &testStruct)
		fmt.Println(testStruct)

		if err == nil {
			fmt.Println("true")
			t.Errorf("Expected an error error but did not get one")
		}

		if strings.Contains(err.Error(), "Failed to unmarshal HTTPResponse Content into outputJSONContent") != true {
			t.Errorf("Expected to receive error for invalid final unmarshal target: " + err.Error())
		}

		if response.StatusCode != testHTTPResponse.StatusCode {
			t.Errorf("Expected status OK, got: " + strconv.Itoa(response.StatusCode))
		}
		if response.Error != testHTTPResponse.Error {
			t.Errorf("Expected error=false, got something else")
		}
		if response.ContentType != HTTPResponseTypeJSON {
			t.Errorf("Expected content type=" + HTTPResponseTypeJSON + ", got " + response.ContentType)
		}
	})
}

// TestDeleteEmptyAndTrim validates that the DeleteEmptyAndTrim function
// properly trims whitespace from each entry in a string slice
func TestDeleteEmptyAndTrim(t *testing.T) {
	assert := assert.New(t)
	input := []string{" abc ", "def ", "gh i "}
	output := []string{"abc", "def", "gh i"}
	assert.Equal(output, DeleteEmptyAndTrim(input))
}
