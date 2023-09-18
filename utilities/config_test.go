// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package utilities

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

type testTableGetGenericErrorStruct struct {
	TestCaseName      string
	Output            string
	InputConfKey      string
	InputConfItemType reflect.Type
}

func prepGetGenericErrorTest() ([]testTableGetGenericErrorStruct, error) {
	testTimeDuration, err := time.ParseDuration("1s")
	if err != nil {
		return []testTableGetGenericErrorStruct{}, fmt.Errorf("Failed to set up time duration test value: %v", err.Error())
	}
	testTableGetGenericError := []testTableGetGenericErrorStruct{
		{
			TestCaseName:      "Time duration string input",
			Output:            fmt.Sprintf("The \"%v\" application setting has not been set. Please set this to an acceptable %v value in configuration.toml", "AverageTemperatureMeasurementDuration", "time.Duration"),
			InputConfKey:      "AverageTemperatureMeasurementDuration",
			InputConfItemType: reflect.TypeOf(testTimeDuration),
		},
		{
			TestCaseName:      "String slice input",
			Output:            fmt.Sprintf("The \"%v\" application setting has not been set. Please set this to an acceptable %v value in configuration.toml", "NotificationEmailAddresses", "[]string"),
			InputConfKey:      "NotificationEmailAddresses",
			InputConfItemType: reflect.TypeOf([]string{}),
		},
		{
			TestCaseName:      "Integer input",
			Output:            fmt.Sprintf("The \"%v\" application setting has not been set. Please set this to an acceptable %v value in configuration.toml", "NotificationSubscriptionMaxRESTRetries", "int"),
			InputConfKey:      "NotificationSubscriptionMaxRESTRetries",
			InputConfItemType: reflect.TypeOf(10),
		},
		{
			TestCaseName:      "Uint input",
			Output:            fmt.Sprintf("The \"%v\" application setting has not been set. Please set this to an acceptable %v value in configuration.toml", "NotificationSubscriptionMaxRESTRetries", "uint"),
			InputConfKey:      "NotificationSubscriptionMaxRESTRetries",
			InputConfItemType: reflect.TypeOf(uint(10)),
		},
	}
	return testTableGetGenericError, nil
}

// TestGetGenericError validates that the GetGenericError function
// returns an error message with a particular format
func TestGetGenericError(t *testing.T) {
	testTable, err := prepGetGenericErrorTest()
	require.NoError(t, err)
	// Test the function
	for _, testCase := range testTable {
		ct := testCase // pinning "current test" solves concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			output := GetGenericError(ct.InputConfKey, ct.InputConfItemType)
			assert.EqualError(t, output, ct.Output)
		})
	}
}

type testTableGetGenericParseErrorStruct struct {
	TestCaseName      string
	Output            string
	InputConfKey      string
	InputConfValue    string
	InputConfItemType reflect.Type
}

func prepGetGenericParseErrorTest() ([]testTableGetGenericParseErrorStruct, error) {
	testTimeDuration, err := time.ParseDuration("1s")
	if err != nil {
		return []testTableGetGenericParseErrorStruct{}, fmt.Errorf("Failed to set up time duration test value: %v", err.Error())
	}
	testTableGetGenericParseError := []testTableGetGenericParseErrorStruct{
		{
			TestCaseName:      "Time duration input",
			Output:            fmt.Sprintf("The \"%v\" application setting been set, but failed to parse the value \"%v\". Please set this to an acceptable %v value in configuration.toml", "AverageTemperatureMeasurementDuration", "-15s", "time.Duration"),
			InputConfKey:      "AverageTemperatureMeasurementDuration",
			InputConfValue:    "-15s",
			InputConfItemType: reflect.TypeOf(testTimeDuration),
		},
		{
			TestCaseName:      "String slice input",
			Output:            fmt.Sprintf("The \"%v\" application setting been set, but failed to parse the value \"%v\". Please set this to an acceptable %v value in configuration.toml", "NotificationEmailAddresses", "test@site.com,test2@site.com", "[]string"),
			InputConfKey:      "NotificationEmailAddresses",
			InputConfValue:    "test@site.com,test2@site.com",
			InputConfItemType: reflect.TypeOf([]string{}),
		},
		{
			TestCaseName:      "Integer input",
			Output:            fmt.Sprintf("The \"%v\" application setting been set, but failed to parse the value \"%v\". Please set this to an acceptable %v value in configuration.toml", "NotificationSubscriptionMaxRESTRetries", "10", "int"),
			InputConfKey:      "NotificationSubscriptionMaxRESTRetries",
			InputConfValue:    "10",
			InputConfItemType: reflect.TypeOf(10),
		},
		{
			TestCaseName:      "Uint input",
			Output:            fmt.Sprintf("The \"%v\" application setting been set, but failed to parse the value \"%v\". Please set this to an acceptable %v value in configuration.toml", "NotificationSubscriptionMaxRESTRetries", "10", "uint"),
			InputConfKey:      "NotificationSubscriptionMaxRESTRetries",
			InputConfValue:    "10",
			InputConfItemType: reflect.TypeOf(uint(10)),
		},
	}
	return testTableGetGenericParseError, nil
}

// TestGetGenericParseError validates that the GetGenericParseError function
// returns an error message with a particular format
func TestGetGenericParseError(t *testing.T) {
	testTable, err := prepGetGenericParseErrorTest()
	require.NoError(t, err)
	// Test the function
	for _, testCase := range testTable {
		ct := testCase // pinning "current test" solves concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			output := GetGenericParseError(ct.InputConfKey, ct.InputConfValue, ct.InputConfItemType)
			assert.EqualError(t, output, ct.Output)
		})
	}
}

type testTableMarshalSettingsStruct struct {
	TestCaseName    string
	Error           error
	InputMap        map[string]string
	InputStrict     bool
	InputInterface  interface{}
	OutputInterface interface{}
}

// These field names would have followed the format "fXxx" instead of "FXxx",
// but unfortunately the reflect package requires fields to be exported.
type allTypes struct {
	FBool         bool
	FByteSlice    []byte
	FFloat32      float32
	FFloat64      float64
	FInt          int
	FInt8         int8
	FInt16        int16
	FInt32        int32
	FInt64        int64
	FString       string
	FStringSlice  []string
	FTimeDuration time.Duration
	FUint         uint
	FUint64       uint64
	FUint8        uint8
	FUint32       uint32
}

type unexportedType struct {
	ExportedField   string
	unexportedField string
}

type unsupportedType struct {
	FBool chan bool
}

// For ease of copy/paste, using "const" on every line was chosen instead of
// using const ( ) notation. Gofmt forces var statements to have a line of
// separation between them.

// The following are names of the fields used in the allTypes struct,
// hence the "fXxx" naming convention (f stands for "field")
const fBool = "FBool"
const fByteSlice = "FByteSlice"
const fFloat32 = "FFloat32"
const fFloat64 = "FFloat64"
const fInt = "FInt"
const fInt8 = "FInt8"
const fInt16 = "FInt16"
const fInt32 = "FInt32"
const fInt64 = "FInt64"
const fString = "FString"
const fStringSlice = "FStringSlice"
const fTimeDuration = "FTimeDuration"
const fUint = "FUint"
const fUint64 = "FUint64"
const fUint8 = "FUint8"
const fUint32 = "FUint32"

// proper versions of all supported parseable types, that are to be used
// in tests, hence the "tXxx" prefix (it's a "test" value, hence "t" prefix)
const tBool = true

var tByteSlice = []byte("test bytes")

const tFloat32 = float32(1000.0)
const tFloat64 = float64(1001.0)
const tInt = int(42)
const tInt8 = int8(43)
const tInt16 = int16(44)
const tInt32 = int32(45)
const tInt64 = int64(46)
const tString = "test string"

var tStringSlice = []string{"test", "string"}

const tTimeDuration = time.Duration(1 * time.Second)
const tUint = uint(47)
const tUint64 = uint64(48)
const tUint8 = uint8(49)
const tUint32 = uint32(50)

// string versions of the above supported parseable types
const sBool = "true"
const sByteSlice = "test bytes"
const sFloat32 = "1000.0"
const sFloat64 = "1001.0"
const sInt = "42"
const sInt8 = "43"
const sInt16 = "44"
const sInt32 = "45"
const sInt64 = "46"
const sString = "test string"
const sStringSlice = "test, string"
const sTimeDuration = "1s"
const sUint = "47"
const sUint64 = "48"
const sUint8 = "49"
const sUint32 = "50"

func prepMarshalSettingsTest() []testTableMarshalSettingsStruct {
	testChanBool := make(chan bool)
	return []testTableMarshalSettingsStruct{
		{
			TestCaseName: "Success",
			Error:        nil,
			InputMap: map[string]string{
				fBool:         sBool,
				fByteSlice:    sByteSlice,
				fFloat32:      sFloat32,
				fFloat64:      sFloat64,
				fInt:          sInt,
				fInt8:         sInt8,
				fInt16:        sInt16,
				fInt32:        sInt32,
				fInt64:        sInt64,
				fString:       sString,
				fStringSlice:  sStringSlice,
				fTimeDuration: sTimeDuration,
				fUint:         sUint,
				fUint64:       sUint64,
				fUint8:        sUint8,
				fUint32:       sUint32,
			},
			InputStrict:    false,
			InputInterface: &allTypes{},
			OutputInterface: &allTypes{
				FBool:         tBool,
				FByteSlice:    tByteSlice,
				FFloat32:      tFloat32,
				FFloat64:      tFloat64,
				FInt:          tInt,
				FInt8:         tInt8,
				FInt16:        tInt16,
				FInt32:        tInt32,
				FInt64:        tInt64,
				FString:       tString,
				FStringSlice:  tStringSlice,
				FTimeDuration: tTimeDuration,
				FUint:         tUint,
				FUint64:       tUint64,
				FUint8:        tUint8,
				FUint32:       tUint32,
			},
		},
		{
			TestCaseName:    "Missing value in strict mode",
			Error:           GetGenericError("FBool", reflect.TypeOf(false)),
			InputMap:        map[string]string{},
			InputStrict:     true,
			InputInterface:  &allTypes{},
			OutputInterface: &allTypes{},
		},
		{
			TestCaseName:    "Passed an interface{} instead of a pointer to an interface{}",
			Error:           fmt.Errorf("input config must be a pointer to an interface"),
			InputMap:        map[string]string{},
			InputStrict:     false,
			InputInterface:  allTypes{},
			OutputInterface: allTypes{},
		},
		{
			TestCaseName: "Type that contains an unexported field",
			Error:        nil,
			InputMap: map[string]string{
				"ExportedField": "test value",
			},
			InputStrict:    false,
			InputInterface: &unexportedType{},
			OutputInterface: &unexportedType{
				ExportedField:   "test value",
				unexportedField: "",
			},
		},
		{
			TestCaseName: "Failure to parse a value",
			Error:        GetGenericParseError("FBool", "invalid boolean value", reflect.TypeOf(false)),
			InputMap: map[string]string{
				fBool: "invalid boolean value",
			},
			InputStrict:     false,
			InputInterface:  &allTypes{},
			OutputInterface: &allTypes{},
		},
		{
			TestCaseName: "Unsupported type",
			Error:        GetGenericParseError("FBool", "invalid boolean value", reflect.TypeOf(testChanBool)),
			InputMap: map[string]string{
				fBool: "invalid boolean value",
			},
			InputStrict: false,
			InputInterface: &unsupportedType{
				FBool: testChanBool,
			},
			OutputInterface: &unsupportedType{
				FBool: testChanBool,
			},
		},
	}
}

// TestMarshalSettings validates that the MarshalSettings
// function properly marshals a map[string]string into an arbitrary interface
// with support for converting a standard set of strings into Go types
func TestMarshalSettings(t *testing.T) {
	testTable := prepMarshalSettingsTest()
	for _, testCase := range testTable {
		ct := testCase // pinning the "current test" solves concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			assert := assert.New(t)
			err := MarshalSettings(ct.InputMap, ct.InputInterface, ct.InputStrict)
			assert.Equal(ct.Error, err)
			assert.Equal(ct.OutputInterface, ct.InputInterface)
		})
	}
}
