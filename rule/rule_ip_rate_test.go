// Copyright 2023 The casbin Authors. All Rights Reserved.
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
	"net/http"
	"testing"

	"github.com/casdoor/casdoor/object"
)

func TestIpRateRule_checkRule(t *testing.T) {
	type fields struct {
		ruleName string
	}
	type args struct {
		args []struct {
			expressions []*object.Expression
			req         *http.Request
		}
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []bool
		want1   []string
		want2   []string
		wantErr []bool
	}{
		{
			name: "Test 1",
			fields: fields{
				ruleName: "rule1",
			},
			args: args{
				args: []struct {
					expressions []*object.Expression
					req         *http.Request
				}{
					{
						expressions: []*object.Expression{
							{
								Operator: "1",
								Value:    "1",
							},
						},
						req: &http.Request{
							RemoteAddr: "127.0.0.1",
						},
					},
					{
						expressions: []*object.Expression{
							{
								Operator: "1",
								Value:    "1",
							},
						},
						req: &http.Request{
							RemoteAddr: "127.0.0.1",
						},
					},
				},
			},
			want:    []bool{false, true},
			want1:   []string{"", "Block"},
			want2:   []string{"", "Rate limit exceeded"},
			wantErr: []bool{false, false},
		},
		{
			name: "Test 2",
			fields: fields{
				ruleName: "rule2",
			},
			args: args{
				args: []struct {
					expressions []*object.Expression
					req         *http.Request
				}{
					{
						expressions: []*object.Expression{
							{
								Operator: "1",
								Value:    "1",
							},
						},
						req: &http.Request{
							RemoteAddr: "127.0.0.1",
						},
					},
					{
						expressions: []*object.Expression{
							{
								Operator: "10",
								Value:    "1",
							},
						},
						req: &http.Request{
							RemoteAddr: "127.0.0.1",
						},
					},
				},
			},
			want:    []bool{false, false},
			want1:   []string{"", ""},
			want2:   []string{"", ""},
			wantErr: []bool{false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &IpRateRule{
				ruleName: tt.fields.ruleName,
			}
			for i, arg := range tt.args.args {
				result, err := r.checkRule(arg.expressions, arg.req)
				if (err != nil) != tt.wantErr[i] {
					t.Errorf("checkRule() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				got := result != nil
				got1 := ""
				got2 := ""
				if result != nil {
					got1 = result.Action
					got2 = result.Reason
				}
				if got != tt.want[i] {
					t.Errorf("checkRule() got = %v, want %v", got, tt.want[i])
				}
				if got1 != tt.want1[i] {
					t.Errorf("checkRule() got1 = %v, want %v", got1, tt.want1[i])
				}
				if got2 != tt.want2[i] {
					t.Errorf("checkRule() got2 = %v, want %v", got2, tt.want2[i])
				}
			}
		})
	}
}

func TestIpRateRule_checkRuleBlackListKeepsOtherIps(t *testing.T) {
	ruleName := "admin/test-ip-rate-blacklist"
	blockedIp := "1.1.1.1"
	otherIp := "2.2.2.2"
	r := &IpRateRule{
		ruleName: ruleName,
	}

	// a rate limit of 0 always denies, so both IPs are put on the black list
	// without depending on how long the test takes to run
	denyAll := []*object.Expression{
		{
			Operator: "0",
			Value:    "3600",
		},
	}

	for _, ip := range []string{blockedIp, otherIp} {
		result, err := r.checkRule(denyAll, &http.Request{RemoteAddr: ip})
		if err != nil {
			t.Fatalf("checkRule() error = %v", err)
		}
		if result == nil || result.Action != "Block" {
			t.Fatalf("checkRule() for the IP: %s should be blocked, got: %+v", ip, result)
		}
	}

	if n := len(blackList[ruleName]); n != 2 {
		t.Errorf("black list of %s has %d entries after blocking %s and %s, want 2", ruleName, n, blockedIp, otherIp)
	}

	// a rate limit that allows the request again: only the black list entry can
	// keep the IP blocked while its block time is not over yet
	allowAll := []*object.Expression{
		{
			Operator: "1000",
			Value:    "3600",
		},
	}

	result, err := r.checkRule(allowAll, &http.Request{RemoteAddr: blockedIp})
	if err != nil {
		t.Fatalf("checkRule() error = %v", err)
	}
	if result == nil || result.Action != "Block" {
		t.Errorf("checkRule() %s is still within its block time, it should stay blocked, got: %+v", blockedIp, result)
	}
}
