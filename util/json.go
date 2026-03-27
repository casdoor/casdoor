// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

func StructToJson(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func StructToJsonFormatted(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(data)
}

func JsonToStruct(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

func ConvertJsonValue(value interface{}) (interface{}, error) {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		return convertMapToAnonymousStruct(typedValue)
	case []interface{}:
		result := make([]interface{}, len(typedValue))
		for i, item := range typedValue {
			convertedItem, err := ConvertJsonValue(item)
			if err != nil {
				return nil, err
			}
			result[i] = convertedItem
		}
		return result, nil
	default:
		return value, nil
	}
}

func TryJsonToAnonymousStruct(j string) (interface{}, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(j), &data); err != nil {
		return nil, err
	}

	switch data.(type) {
	case map[string]interface{}, []interface{}:
		return ConvertJsonValue(data)
	default:
		return nil, fmt.Errorf("JSON value is not an object or array")
	}
}

func convertMapToAnonymousStruct(data map[string]interface{}) (interface{}, error) {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fields := make([]reflect.StructField, 0, len(keys))
	values := make([]interface{}, 0, len(keys))
	usedNames := map[string]int{}

	for _, key := range keys {
		convertedValue, err := ConvertJsonValue(data[key])
		if err != nil {
			return nil, err
		}

		fieldName := jsonKeyToExportedFieldName(key)
		if fieldName == "" {
			fieldName = "Field"
		}
		if count := usedNames[fieldName]; count > 0 {
			usedNames[fieldName] = count + 1
			fieldName = fmt.Sprintf("%s%d", fieldName, count+1)
		} else {
			usedNames[fieldName] = 1
		}

		fieldType := reflect.TypeOf(convertedValue)
		if fieldType == nil {
			fieldType = reflect.TypeOf((*interface{})(nil)).Elem()
		}

		fields = append(fields, reflect.StructField{
			Name: fieldName,
			Type: fieldType,
			Tag:  reflect.StructTag(fmt.Sprintf(`json:%q`, key)),
		})
		values = append(values, convertedValue)
	}

	structType, err := safeStructOf(fields)
	if err != nil {
		return nil, err
	}

	structValue := reflect.New(structType).Elem()
	for i, value := range values {
		field := structValue.Field(i)
		if value == nil {
			field.Set(reflect.Zero(field.Type()))
			continue
		}

		fieldValue := reflect.ValueOf(value)
		if fieldValue.Type().AssignableTo(field.Type()) {
			field.Set(fieldValue)
			continue
		}
		if fieldValue.Type().ConvertibleTo(field.Type()) {
			field.Set(fieldValue.Convert(field.Type()))
			continue
		}

		return nil, fmt.Errorf("cannot assign JSON field %q to %s", keys[i], field.Type())
	}

	return structValue.Addr().Interface(), nil
}

func safeStructOf(fields []reflect.StructField) (_ reflect.Type, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to convert JSON object to anonymous struct: %v", r)
		}
	}()

	return reflect.StructOf(fields), nil
}

func jsonKeyToExportedFieldName(key string) string {
	parts := strings.FieldsFunc(key, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if len(parts) == 0 {
		return ""
	}

	for i, part := range parts {
		parts[i] = upperFirstRune(part)
	}

	fieldName := strings.Join(parts, "")
	if fieldName == "" {
		return ""
	}

	firstRune := []rune(fieldName)[0]
	if unicode.IsDigit(firstRune) {
		fieldName = "Field" + fieldName
	}

	return fieldName
}

func upperFirstRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return ""
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
