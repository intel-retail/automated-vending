// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package utilities

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HTTPFunc is a function type that matches what is typically passed
// into an HTTP API endpoint
type HTTPFunc func(http.ResponseWriter, *http.Request)

// HTTPResponse All API HTTP responses should follow this format.
type HTTPResponse struct {
	Content     interface{} `json:"content"`
	ContentType string      `json:"contentType"`
	StatusCode  int         `json:"statusCode"`
	Error       bool        `json:"error"`
}

// HTTPResponseTypeJSON is a const that allows for efficiently setting
// HTTPResponse structs' ContentType fields to json
const HTTPResponseTypeJSON = "json"

// HTTPResponseTypeString is a const that allows for efficiently setting
// HTTPResponse structs' ContentType fields to string
const HTTPResponseTypeString = "string"

// AllowedHeaders is the set of headers we will process in the ProcessCORS function.
const AllowedHeaders = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token"

//dtFormat is the ISO8601 date format
const dtFormat = "2006-01-02T15:04:05.000000"

// NowDate simply returns the date in ISO8601 format
func NowDate() string {
	return time.Now().Format(dtFormat) // %Y-%m-%dT%H:%M:%S.%f
}

// GenUUID returns a fresh UUID as a string
func GenUUID() string {
	return uuid.New().String()
}

// ValidUUID checks to see if a given string is a valid UUID
func ValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

// ProcessCORS is a decorator function that enables CORS preflight responses and
// sets CORS headers.
// Usage:
// func GetSomething(writer http.ResponseWriter, req *http.Request) {
//     helpers.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
//         // do some logic with writer and req
//     })
// }
func ProcessCORS(writer http.ResponseWriter, req *http.Request, fn HTTPFunc) {
	writer.Header().Set("Content-Type", "application/json")

	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	writer.Header().Set("Access-Control-Allow-Headers", AllowedHeaders)

	// handle preflight options requests
	if req.Method == "OPTIONS" {
		writer.WriteHeader(http.StatusOK)
		return
	}

	fn(writer, req)
}

// WriteToJSONFile takes an input filename and writes JSON from any interface
// The "perm" parameter should generally be 0644
func WriteToJSONFile(filename string, target interface{}, perm os.FileMode) (err error) {
	result, err := GetAsJSON(target)
	if err != nil {
		return fmt.Errorf("Failed to convert input interface to JSON before writing to %v: %v", filename, err.Error())
	}

	err = ioutil.WriteFile(filename, []byte(result), perm)
	if err != nil {
		return fmt.Errorf("Failed to write JSON to file %v: %v", filename, err.Error())
	}

	return nil
}

// LoadFromJSONFile takes an input filename and loads from JSON into any interface
// Make sure that the "target" parameter is used like this
// err := LoadFromJSONFile(filename, &target)
func LoadFromJSONFile(filename string, target interface{}) (err error) {
	// Read from local file inventory.json
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Failed to read from JSON file %v: %v", filename, err.Error())
	}

	// Unmarshal the string contents of inventory.json into a proper structure
	if err := json.Unmarshal(dat, &target); err != nil {
		return fmt.Errorf("Failed to unmarshal from JSON file %v: %v", filename, err.Error())
	}

	return nil
}

// GetAsJSON converts any interface into a JSON string
func GetAsJSON(input interface{}) (output string, err error) {
	result, err := json.Marshal(input)

	if err != nil {
		return "", fmt.Errorf("Failed to marshal into JSON string: %s", err.Error())
	}

	return string(result), nil
}

// GetHTTPResponseTemplate gets a template HTTPResponse.
// You can use this as a starting point for all HTTP responses
func GetHTTPResponseTemplate() HTTPResponse {
	newHTTPResponse := HTTPResponse{
		Content:     "",
		ContentType: "string",
		StatusCode:  http.StatusOK,
		Error:       false,
	}
	return newHTTPResponse
}

// GetAsString gets an HTTPResponse as a JSON string.
// If the JSON marshal fails, it will return a template error HTTPResponse
func (response HTTPResponse) GetAsString() string {
	result, err := GetAsJSON(response)
	if err != nil {
		fmt.Println("Failed to convert HTTP response to string: " + err.Error())
		blankHTTPResponse := GetHTTPResponseTemplate()
		blankHTTPResponse.Error = true
		blankHTTPResponse.StatusCode = http.StatusInternalServerError
		blankHTTPResponse.ContentType = "string"
		blankHTTPResponse.Content = "Failed to process response."
		// BUG: This can recurse infinitely if there is something
		// wrong with the GetAsJSON() function
		result = blankHTTPResponse.GetAsString()
	}
	return result
}

// WriteHTTPResponse is a helpful shorthand for writing out a prepared HTTPResponse
func (response *HTTPResponse) WriteHTTPResponse(writer http.ResponseWriter, req *http.Request) {
	responseBytes := []byte(response.GetAsString())
	logOutput := fmt.Sprintf("%v HTTP %v: %v %v Error=%v, Content Type=%v, Response Body Length=%v B", NowDate(), response.StatusCode, req.Method, req.URL, response.Error, response.ContentType, len(responseBytes))
	if response.Error == true {
		logOutput = fmt.Sprintf(logOutput, ", Content:", response.Content)
	}
	fmt.Println(logOutput)
	// Write the HTTP status header
	writer.WriteHeader(response.StatusCode)
	// Write item data back to caller
	writer.Write(responseBytes)
}

// SetHTTPResponseFields sets the four HTTPResponse fields in a single line.
func (response *HTTPResponse) SetHTTPResponseFields(statusCode int, content interface{}, contentType string, error bool) {
	response.StatusCode = statusCode
	response.Error = error
	response.Content = content
	response.ContentType = contentType
}

// SetJSONHTTPResponseFields sets the three common HTTPResponse fields, and the ContentType to JSON
func (response *HTTPResponse) SetJSONHTTPResponseFields(statusCode int, content interface{}, error bool) {
	response.SetHTTPResponseFields(statusCode, content, HTTPResponseTypeJSON, error)
}

// SetStringHTTPResponseFields sets the three common HTTPResponse fields, and the ContentType to string
func (response *HTTPResponse) SetStringHTTPResponseFields(statusCode int, content interface{}, error bool) {
	response.SetHTTPResponseFields(statusCode, content, HTTPResponseTypeString, error)
}

// WriteStringHTTPResponse is a one-liner meant to quickly build out a string
// HTTPResponse, and then respond according to the given inputs
func WriteStringHTTPResponse(w http.ResponseWriter, req *http.Request, statusCode int, content interface{}, error bool) {
	response := GetHTTPResponseTemplate()
	response.SetStringHTTPResponseFields(statusCode, content, error)
	response.WriteHTTPResponse(w, req)
}

// WriteJSONHTTPResponse is a one-liner meant to quickly build out a JSON
// HTTPResponse, and then respond according to the given inputs
func WriteJSONHTTPResponse(w http.ResponseWriter, req *http.Request, statusCode int, content interface{}, error bool) {
	response := GetHTTPResponseTemplate()
	response.SetJSONHTTPResponseFields(statusCode, content, error)
	response.WriteHTTPResponse(w, req)
}

// ParseJSONHTTPResponseContent converts a raw http.Response (i.e. resp.Body) into
// a fully unmarshaled interface, including the content interface
func ParseJSONHTTPResponseContent(respBody io.ReadCloser, outputJSONContent interface{}) (response HTTPResponse, err error) {
	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("Failed to read response body: " + err.Error())
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("Failed to unmarshal response body as an HTTPResponse: " + err.Error())
	}

	if response.ContentType != HTTPResponseTypeJSON {
		return response, fmt.Errorf("HTTPResponse ContentType is not " + HTTPResponseTypeJSON)
	}

	// check if the content is a string otherwise this won't work
	switch t := response.Content.(type) {
	case string:
		break
	default:
		return response, fmt.Errorf("HTTPResponse Content must be a string, instead got %v", t)
	}

	err = json.Unmarshal([]byte(response.Content.(string)), &outputJSONContent)
	if err != nil {
		return response, fmt.Errorf("Failed to unmarshal HTTPResponse Content into outputJSONContent: " + err.Error())
	}

	return
}

// DeleteEmptyAndTrim takes an input string slice and trims whitespace
// surrounding each entry in the slice
func DeleteEmptyAndTrim(s []string) []string {
	var r []string
	for _, str := range s {
		str = strings.TrimSpace(str)
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
