// Copyright 2021 The casbin Authors. All Rights Reserved.
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

package xlsx

import (
	"github.com/casdoor/casdoor/util"
	"github.com/tealeg/xlsx"
)

func ReadXlsxFile(fileId string) [][]string {
	path := util.GetUploadXlsxPath(fileId)
	file, err := xlsx.OpenFile(path)
	if err != nil {
		panic(err)
	}

	res := [][]string{}
	for _, sheet := range file.Sheets {
		for _, row := range sheet.Rows {
			line := []string{}
			for _, cell := range row.Cells {
				text := cell.String()
				line = append(line, text)
			}
			res = append(res, line)
		}
		break
	}

	return res
}
