// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package utilities

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// GetGenericError helps keep the ProcessApplicationSettings slim by enabling
// re-use of a common error format that should be returned to the user when
// the user fails to specify a configuration item.
func GetGenericError(confKey string, confItemType reflect.Type) error {
	return fmt.Errorf("The \"%v\" application setting has not been set. Please set this to an acceptable %v value in configuration.toml", confKey, confItemType)
}

// GetGenericParseError helps keep the ProcessApplicationSettings slim by enabling
// re-use of a common error format that should be returned to the user when
// the user fails to specify a parseable configuration item.
func GetGenericParseError(confKey string, confValue interface{}, confItemType reflect.Type) error {
	return fmt.Errorf("The \"%v\" application setting been set, but failed to parse the value \"%v\". Please set this to an acceptable %v value in configuration.toml", confKey, confValue, confItemType)
}

// MarshalSettings converts a map[string]string into a properly
// marshaled config interface{}. This is useful for converting an EdgeX
// Application Settings or Device Service Driver config map into a proper
// struct. Since values in these maps are strings, some assumptions
// are made about these values when marshaling them into their proper values.
// For example, a config interface that contains a field with type []string
// will assume that the corresponding field name's key in the map is a
// CSV string.
// This function does not cover every possible type in Go,
// so there are bound to be some issues converting. It contains the most common
// types.
// If the strict boolean is set to true, this function will error out when
// a value in the input map is not specified.
// Note that the config parameter must be a pointer to an interface.
// All fields in the config interface should be exported. The reflect module
// does not have the ability to modify unexported fields. Any unexported fields
// will simply be ignored.
func MarshalSettings(settings map[string]string, config interface{}, strict bool) (err error) {
	// because an interface{} can represent a &struct or a regular struct,
	// enforce the operation of this function on pointers only
	if reflect.TypeOf(config).Kind() != reflect.Ptr {
		return fmt.Errorf("input config must be a pointer to an interface")
	}

	val := reflect.ValueOf(config).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		valueField := val.Field(i)

		// a value can be reported as invalid if it has the zero value
		if !valueField.IsValid() {
			return fmt.Errorf("input config field %v is not a valid field according to reflect", typeField.Name)
		}

		// unexported struct fields cannot be changed, so we skip this field
		if !valueField.CanSet() {
			continue
		}

		// see if the settings map contains a value for this field
		val, ok := settings[typeField.Name]
		if !ok && strict {
			// quit if it does not contain the field, and if strict is true
			return GetGenericError(typeField.Name, typeField.Type)
		}

		var err error
		// parse string values into go types according to this field's type
		switch valueField.Interface().(type) {
		case bool:
			newValue := false
			newValue, err = strconv.ParseBool(val)
			valueField.SetBool(newValue)
		case []byte:
			parsedVal := []byte(val)
			valueField.Set(reflect.ValueOf(parsedVal))
		case float32:
			parsedVal := float64(0.0)
			parsedVal, err = strconv.ParseFloat(val, 32)
			valueField.SetFloat(parsedVal)
		case float64:
			parsedVal := float64(0.0)
			parsedVal, err = strconv.ParseFloat(val, 64)
			valueField.SetFloat(parsedVal)
		case int8:
			parsedVal := int64(0.0)
			parsedVal, err = strconv.ParseInt(val, 10, 8)
			valueField.SetInt(parsedVal)
		case int16:
			parsedVal := int64(0.0)
			parsedVal, err = strconv.ParseInt(val, 10, 16)
			valueField.SetInt(parsedVal)
		case int32:
			parsedVal := int64(0.0)
			parsedVal, err = strconv.ParseInt(val, 10, 32)
			valueField.SetInt(parsedVal)
		case int64:
			parsedVal := int64(0.0)
			parsedVal, err = strconv.ParseInt(val, 10, 64)
			valueField.SetInt(parsedVal)
		case int:
			parsedVal := int64(0.0)
			parsedVal, err = strconv.ParseInt(val, 10, strconv.IntSize)
			valueField.SetInt(int64(parsedVal))
		case string:
			valueField.SetString(val)
		case []string:
			valueField.Set(reflect.ValueOf(DeleteEmptyAndTrim(strings.Split(val, ","))))
		case time.Duration:
			var parsedVal time.Duration
			parsedVal, err = time.ParseDuration(val)
			valueField.Set(reflect.ValueOf(parsedVal))
		case uint64:
			parsedVal := uint64(0)
			parsedVal, err = strconv.ParseUint(val, 10, 64)
			valueField.SetUint(parsedVal)
		case uint:
			parsedVal := uint64(0)
			parsedVal, err = strconv.ParseUint(val, 10, strconv.IntSize)
			valueField.SetUint(parsedVal)
		case uint8:
			parsedVal := uint64(0)
			parsedVal, err = strconv.ParseUint(val, 10, 8)
			valueField.SetUint(parsedVal)
		case uint16:
			parsedVal := uint64(0)
			parsedVal, err = strconv.ParseUint(val, 10, 16)
			valueField.SetUint(parsedVal)
		case uint32:
			parsedVal := uint64(0)
			parsedVal, err = strconv.ParseUint(val, 10, 32)
			valueField.SetUint(parsedVal)
		default:
			// in the default case, we do not know how to treat the value
			return GetGenericParseError(typeField.Name, val, typeField.Type)
		}
		if err != nil {
			return GetGenericParseError(typeField.Name, val, typeField.Type)
		}
	}
	return nil
}
