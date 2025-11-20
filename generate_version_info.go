// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

package main

import (
	"fmt"
	"os"

	"github.com/casdoor/casdoor/util"
)

func main() {
	versionInfo, err := util.GetVersionInfo()
	if err != nil {
		// If git operations fail, use default values
		fmt.Fprintf(os.Stderr, "Warning: Failed to get version info from git: %v\n", err)
		fmt.Fprintf(os.Stderr, "Using default version info\n")
		versionInfo = &util.VersionInfo{
			Version:      "unknown",
			CommitId:     "unknown",
			CommitOffset: 0,
		}
	}
	
	// Output in the format expected by GetVersionInfoFromFile()
	// Format: {version commitId commitOffset}
	fmt.Printf("{%s %s %d}\n", versionInfo.Version, versionInfo.CommitId, versionInfo.CommitOffset)
}
