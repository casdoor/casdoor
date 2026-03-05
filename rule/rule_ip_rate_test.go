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
