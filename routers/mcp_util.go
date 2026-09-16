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

package routers

import (
	"encoding/json"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// applicationStub is a lightweight struct for extracting owner/name from application data
type applicationStub struct {
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	Organization string `json:"organization"`
}

// userStub is a lightweight struct for extracting owner/name from user data
type userStub struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

// getMcpObjects returns every object an MCP tool call acts on, so that the tool
// arguments are authorized the same way as the query params and body of the REST APIs.
func getMcpObjects(ctx *context.Context) ([]Object, error) {
	body := ctx.Input.RequestBody
	if len(body) == 0 {
		return nil, nil
	}

	// Parse MCP request to determine tool name
	type mcpRequest struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params,omitempty"`
	}

	type mcpCallToolParams struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments,omitempty"`
	}

	type getApplicationsArgs struct {
		Owner string `json:"owner"`
	}

	type getApplicationArgs struct {
		Id string `json:"id"`
	}

	type addApplicationArgs struct {
		Application applicationStub `json:"application"`
	}

	type updateApplicationArgs struct {
		Id          string          `json:"id"`
		Application applicationStub `json:"application"`
	}

	type deleteApplicationArgs struct {
		Application applicationStub `json:"application"`
	}

	type getUsersArgs struct {
		Owner string `json:"owner"`
	}

	type getUserArgs struct {
		Id    string `json:"id"`
		Owner string `json:"owner"`
	}

	type addUserArgs struct {
		User userStub `json:"user"`
	}

	type updateUserArgs struct {
		Id   string   `json:"id"`
		User userStub `json:"user"`
	}

	type deleteUserArgs struct {
		User userStub `json:"user"`
	}

	var mcpReq mcpRequest
	err := json.Unmarshal(body, &mcpReq)
	if err != nil {
		return nil, nil
	}

	// Only extract object for tool calls
	if mcpReq.Method != "tools/call" {
		return nil, nil
	}

	var params mcpCallToolParams
	err = json.Unmarshal(mcpReq.Params, &params)
	if err != nil {
		return nil, nil
	}

	// Extract owner/id from arguments based on tool
	switch params.Name {
	case "get_applications":
		var args getApplicationsArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return []Object{{Owner: args.Owner}}, nil
		}
	case "get_application":
		var args getApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			obj, err := getMcpApplicationObject(args.Id)
			return []Object{obj}, err
		}
	case "update_application":
		var args updateApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			obj, err := getMcpApplicationObject(args.Id)
			if err != nil {
				return nil, err
			}
			return appendMcpObject([]Object{obj}, extractOwnerNameFromAppStub(args.Application)), nil
		}
	case "add_application":
		var args addApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return []Object{extractOwnerNameFromAppStub(args.Application)}, nil
		}
	case "delete_application":
		var args deleteApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return []Object{extractOwnerNameFromAppStub(args.Application)}, nil
		}
	case "get_users":
		var args getUsersArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return []Object{{Owner: args.Owner}}, nil
		}
	case "get_user":
		var args getUserArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			if args.Id != "" {
				owner, name, err := util.GetOwnerAndNameFromIdWithError(args.Id)
				return []Object{{Owner: owner, Name: name}}, err
			}
			return []Object{{Owner: args.Owner}}, nil
		}
	case "add_user":
		var args addUserArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return []Object{{Owner: args.User.Owner, Name: args.User.Name}}, nil
		}
	case "update_user":
		var args updateUserArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			owner, name, err := util.GetOwnerAndNameFromIdWithError(args.Id)
			if err != nil {
				return nil, err
			}
			return appendMcpObject([]Object{{Owner: owner, Name: name}}, Object{Owner: args.User.Owner, Name: args.User.Name}), nil
		}
	case "delete_user":
		var args deleteUserArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return []Object{{Owner: args.User.Owner, Name: args.User.Name}}, nil
		}
	}

	return nil, nil
}

// getMcpApplicationObject resolves an application id to its organization, which is
// the object the REST application APIs are authorized against.
func getMcpApplicationObject(id string) (Object, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return Object{}, err
	}

	application, err := object.GetApplication(id)
	if err != nil {
		return Object{}, err
	}
	if application != nil {
		owner = application.Organization
	}

	return Object{Owner: owner, Name: name}, nil
}

func appendMcpObject(objects []Object, obj Object) []Object {
	if obj.Owner != "" && obj != objects[0] {
		objects = append(objects, obj)
	}
	return objects
}

// extractOwnerNameFromAppStub extracts owner and name from application stub
// Prioritizes organization field over owner field for consistency
func extractOwnerNameFromAppStub(app applicationStub) Object {
	// Try organization field first (used in application APIs)
	if app.Organization != "" {
		return Object{Owner: app.Organization, Name: app.Name}
	}
	// Fall back to owner field
	return Object{Owner: app.Owner, Name: app.Name}
}
