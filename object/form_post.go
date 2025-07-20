// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

const formPostTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Submit This Form</title>
    <meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
    <meta http-equiv="Pragma" content="no-cache" />
    <meta http-equiv="Expires" content="0" />
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
        .loading { font-size: 18px; color: #666; }
    </style>
</head>
<body onload="document.getElementById('form').submit()">
    <form id="form" method="post" action="{{.RedirectUri}}">
        {{range $key, $value := .Parameters}}
        <input type="hidden" name="{{$key}}" value="{{$value}}" />
        {{end}}
    </form>
</body>
</html>`

type FormPostResponse struct {
	RedirectUri string
	Parameters  map[string]string
}

func GenerateFormPostResponse(redirectUri string, parameters map[string]string) (string, error) {
	if redirectUri == "" {
		return "", fmt.Errorf("redirect URI cannot be empty")
	}

	parsedUri, err := url.Parse(redirectUri)
	if err != nil {
		return "", fmt.Errorf("invalid redirect URI: %s", err.Error())
	}

	if parsedUri.Scheme == "" || parsedUri.Host == "" {
		return "", fmt.Errorf("redirect URI must have scheme and host")
	}

	safeParameters := make(map[string]string)
	for key, value := range parameters {
		safeParameters[template.HTMLEscapeString(key)] = template.HTMLEscapeString(value)
	}

	responseData := FormPostResponse{
		RedirectUri: template.HTMLEscapeString(redirectUri),
		Parameters:  safeParameters,
	}

	tmpl, err := template.New("formpost").Parse(formPostTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse form post template: %s", err.Error())
	}

	var result strings.Builder
	err = tmpl.Execute(&result, responseData)
	if err != nil {
		return "", fmt.Errorf("failed to execute form post template: %s", err.Error())
	}

	return result.String(), nil
}
