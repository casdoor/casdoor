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

package util

import xormadapter "github.com/casdoor/xorm-adapter/v3"

func CasbinToSlice(casbinRule xormadapter.CasbinRule) []string {
	s := []string{
		casbinRule.V0,
		casbinRule.V1,
		casbinRule.V2,
		casbinRule.V3,
		casbinRule.V4,
		casbinRule.V5,
	}
	// remove empty strings from end, for update model policy map
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != "" {
			s = s[:i+1]
			break
		}
	}
	return s
}

func safeReturn(policy []string, i int) string {
	if len(policy) > i {
		return policy[i]
	} else {
		return ""
	}
}

func MatrixToCasbinRules(Ptype string, policies [][]string) []*xormadapter.CasbinRule {
	res := []*xormadapter.CasbinRule{}

	for _, policy := range policies {
		line := xormadapter.CasbinRule{
			Ptype: Ptype,
			V0:    safeReturn(policy, 0),
			V1:    safeReturn(policy, 1),
			V2:    safeReturn(policy, 2),
			V3:    safeReturn(policy, 3),
			V4:    safeReturn(policy, 4),
			V5:    safeReturn(policy, 5),
		}
		res = append(res, &line)
	}

	return res
}
