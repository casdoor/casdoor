// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package scan

import (
	"fmt"

	"github.com/casdoor/casdoor/object"
)

type ScanProvider interface {
	Scan(target string, command string) (string, error)
	ParseResult(rawResult string) (string, error)
	GetResultSummary(result string) string
}

func GetScanProviderFromProvider(provider *object.Provider) (ScanProvider, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider is nil")
	}

	switch {
	case provider.Category == ScanProviderCategory && provider.Type == McpScanProviderType && provider.SubType == IntranetScanProviderSubType:
		return NewIntranetServerProvider(), nil
	}

	return nil, fmt.Errorf("scan provider type: %s (sub type: %s) is not supported", provider.Type, provider.SubType)
}
