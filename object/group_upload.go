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

package object

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/casdoor/casdoor/xlsx"
)

func StringArrayToGroup(stringArray [][]string) ([]*Group, error) {
	fieldNames := stringArray[0]
	excelMap := []map[string]string{}
	groupFieldMap := map[string]int{}

	reflectedGroup := reflect.TypeOf(Group{})
	for i := 0; i < reflectedGroup.NumField(); i++ {
		groupFieldMap[strings.ToLower(reflectedGroup.Field(i).Name)] = i
	}

	for idx, field := range stringArray {
		if idx == 0 {
			continue
		}

		isEmptyRow := true
		for _, val := range field {
			if strings.TrimSpace(val) != "" {
				isEmptyRow = false
				break
			}
		}
		if isEmptyRow {
			continue
		}

		tempMap := map[string]string{}
		for idx, val := range field {
			if idx >= len(fieldNames) {
				continue
			}
			tempMap[fieldNames[idx]] = val
		}
		excelMap = append(excelMap, tempMap)
	}

	groups := []*Group{}

	for _, g := range excelMap {
		group := Group{}
		reflectedGroup := reflect.ValueOf(&group).Elem()
		for k, v := range g {
			if v == "" || v == "null" || v == "[]" || v == "{}" {
				continue
			}

			fName := strings.ToLower(strings.ReplaceAll(k, "_", ""))
			fieldIdx, ok := groupFieldMap[fName]
			if !ok {
				continue
			}

			fv := reflectedGroup.Field(fieldIdx)
			if !fv.IsValid() {
				continue
			}

			switch fv.Kind() {
			case reflect.String:
				fv.SetString(v)
			case reflect.Bool:
				fv.SetBool(v == "1")
			case reflect.Int:
				intVal, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				fv.SetInt(int64(intVal))
			default:
				switch fv.Type() {
				case reflect.TypeOf([]string{}):
					var strSlice []string
					if err := json.Unmarshal([]byte(v), &strSlice); err != nil {
						return nil, err
					}
					fv.Set(reflect.ValueOf(strSlice))
				}
			}
		}
		groups = append(groups, &group)
	}

	return groups, nil
}

func getGroupMap(owner string) (map[string]*Group, error) {
	m := map[string]*Group{}

	groups, err := GetGroups(owner)
	if err != nil {
		return m, err
	}

	for _, group := range groups {
		m[group.GetId()] = group
	}

	return m, nil
}

func UploadGroups(owner string, path string) (bool, error) {
	table := xlsx.ReadXlsxFile(path)

	oldGroupMap, err := getGroupMap(owner)
	if err != nil {
		return false, err
	}

	transGroups, err := StringArrayToStruct[Group](table)
	if err != nil {
		return false, err
	}

	newGroups := []*Group{}
	for _, group := range transGroups {
		if _, ok := oldGroupMap[group.GetId()]; !ok {
			newGroups = append(newGroups, group)
		}
	}

	if len(newGroups) == 0 {
		return false, nil
	}

	return AddGroupsInBatch(newGroups)
}
