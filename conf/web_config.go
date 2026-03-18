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

package conf

type WebConfig struct {
	ShowGithubCorner bool   `json:"showGithubCorner"`
	ForceLanguage    string `json:"forceLanguage"`
	DefaultLanguage  string `json:"defaultLanguage"`
	IsDemoMode       bool   `json:"isDemoMode"`
	StaticBaseUrl    string `json:"staticBaseUrl"`
	AiAssistantUrl   string `json:"aiAssistantUrl"`
}

func GetWebConfig() *WebConfig {
	config := &WebConfig{}

	config.ShowGithubCorner = GetConfigBool("showGithubCorner")
	config.ForceLanguage = GetLanguage(GetConfigString("forceLanguage"))
	config.DefaultLanguage = GetLanguage(GetConfigString("defaultLanguage"))
	config.IsDemoMode = IsDemoMode()
	config.StaticBaseUrl = GetConfigString("staticBaseUrl")
	config.AiAssistantUrl = GetConfigString("aiAssistantUrl")

	return config
}
