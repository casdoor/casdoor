// Copyright 2024 The casbin Authors. All Rights Reserved.
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

package rule

import (
	"fmt"
	"net/http"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type CompoundRule struct{}

func (r *CompoundRule) checkRule(expressions []*object.Expression, req *http.Request) (*RuleResult, error) {
	operators := util.NewStack()
	res := true
	for _, expression := range expressions {
		isHit := true
		result, err := CheckRules([]string{expression.Value}, req)
		if err != nil {
			return nil, err
		}
		if result == nil || result.Action == "" {
			isHit = false
		}
		switch expression.Operator {
		case "and", "begin":
			res = res && isHit
		case "or":
			operators.Push(res)
			res = isHit
		default:
			return nil, fmt.Errorf("unknown operator: %s", expression.Operator)
		}
		if operators.Size() > 0 {
			last, ok := operators.Pop()
			for ok {
				res = last.(bool) || res
				last, ok = operators.Pop()
			}
		}
	}
	if res {
		return &RuleResult{}, nil
	}
	return nil, nil
}
