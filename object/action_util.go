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

package object

import (
	"encoding/json"
	"fmt"

	"github.com/casdoor/casdoor/action"
)

// getInstalledAction loads an action installed in the organization, checking that it is of the kind the caller runs
func getInstalledAction(owner string, name string, kind string) (*action.Module, *actionFile, error) {
	module, hosts, err := getActionModule(owner, name)
	if err != nil {
		return nil, nil, err
	}

	integration, err := getIntegration(owner, name)
	if err != nil {
		return nil, nil, err
	}
	bundle := &IntegrationBundle{}
	err = json.Unmarshal([]byte(integration.Bundle), bundle)
	if err != nil {
		return nil, nil, err
	}
	file, err := getActionFile(bundle)
	if err != nil {
		return nil, nil, err
	}
	if file.Kind != kind {
		return nil, nil, fmt.Errorf("the action: %s is of kind %s, not %s", name, file.Kind, kind)
	}
	return &action.Module{Name: name, Module: module, Hosts: hosts}, file, nil
}
