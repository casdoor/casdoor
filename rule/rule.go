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

type Rule interface {
	checkRule(expressions []*object.Expression, req *http.Request) (*RuleResult, error)
}

type RuleResult struct {
	Action     string
	StatusCode int
	Reason     string
}

func CheckRules(ruleIds []string, r *http.Request) (*RuleResult, error) {
	rules, err := object.GetRulesByRuleIds(ruleIds)
	if err != nil {
		return nil, err
	}
	for i, rule := range rules {
		var ruleObj Rule
		switch rule.Type {
		case "User-Agent":
			ruleObj = &UaRule{}
		case "IP":
			ruleObj = &IpRule{}
		case "WAF":
			ruleObj = &WafRule{}
		case "IP Rate Limiting":
			ruleObj = &IpRateRule{
				ruleName: rule.GetId(),
			}
		case "Compound":
			ruleObj = &CompoundRule{}
		default:
			return nil, fmt.Errorf("unknown rule type: %s for rule: %s", rule.Type, rule.GetId())
		}

		result, err := ruleObj.checkRule(rule.Expressions, r)
		if err != nil {
			return nil, err
		}

		if result != nil {
			// Use rule's action if no action specified by the rule check
			if result.Action == "" {
				result.Action = rule.Action
			}

			// Determine status code
			if result.StatusCode == 0 {
				if rule.StatusCode != 0 {
					result.StatusCode = rule.StatusCode
				} else {
					// Set default status codes if not specified
					switch result.Action {
					case "Block":
						result.StatusCode = 403
					case "Drop":
						result.StatusCode = 400
					case "Allow":
						result.StatusCode = 200
					case "CAPTCHA":
						result.StatusCode = 302
					default:
						return nil, fmt.Errorf("unknown rule action: %s for rule: %s", result.Action, rule.GetId())
					}
				}
			}

			// Update reason if rule has custom reason
			if result.Action == "Block" || result.Action == "Drop" {
				if rule.IsVerbose {
					// Add verbose debug info with rule name and triggered expression
					result.Reason = util.GenerateVerboseReason(rule.GetId(), result.Reason, rule.Reason)
				} else if rule.Reason != "" {
					result.Reason = rule.Reason
				} else if result.Reason != "" {
					result.Reason = fmt.Sprintf("hit rule %s: %s", ruleIds[i], result.Reason)
				}
			}

			return result, nil
		}
	}

	// Default action if no rule matched
	return &RuleResult{
		Action:     "Allow",
		StatusCode: 200,
	}, nil
}
