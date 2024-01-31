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
	"reflect"
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

func TryJsonToAnonymousStruct(j string) (interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(j), &data); err != nil {
		return nil, err
	}

	// Create a slice of StructFields
	fields := make([]reflect.StructField, 0, len(data))
	for k, v := range data {
		fields = append(fields, reflect.StructField{
			Name: k,
			Type: reflect.TypeOf(v),
		})
	}

	// Create the struct type
	t := reflect.StructOf(fields)

	// Unmarshal again, this time to the new struct type
	val := reflect.New(t)
	i := val.Interface()
	if err := json.Unmarshal([]byte(j), &i); err != nil {
		return nil, err
	}
	return i, nil
}
