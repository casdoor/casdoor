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

// BuiltInTheme represents a predefined theme configuration
type BuiltInTheme struct {
	Name        string     `json:"name"`
	DisplayName string     `json:"displayName"`
	Description string     `json:"description"`
	ThemeData   *ThemeData `json:"themeData"`
}

// GetBuiltInThemes returns a list of predefined beautiful theme configurations
func GetBuiltInThemes() []*BuiltInTheme {
	return []*BuiltInTheme{
		{
			Name:        "default",
			DisplayName: "Default",
			Description: "Casdoor's default theme with purple primary color",
			ThemeData: &ThemeData{
				ThemeType:    "default",
				ColorPrimary: "#5734d3",
				BorderRadius: 6,
				IsCompact:    false,
				IsEnabled:    true,
			},
		},
		{
			Name:        "dark",
			DisplayName: "Dark",
			Description: "Dark theme for better viewing in low-light environments",
			ThemeData: &ThemeData{
				ThemeType:    "dark",
				ColorPrimary: "#5734d3",
				BorderRadius: 2,
				IsCompact:    false,
				IsEnabled:    true,
			},
		},
		{
			Name:        "lark",
			DisplayName: "Document",
			Description: "Professional document-style theme with green accents",
			ThemeData: &ThemeData{
				ThemeType:    "lark",
				ColorPrimary: "#00b96b",
				BorderRadius: 4,
				IsCompact:    false,
				IsEnabled:    true,
			},
		},
		{
			Name:        "comic",
			DisplayName: "Blossom",
			Description: "Playful and friendly theme with pink accents and rounded corners",
			ThemeData: &ThemeData{
				ThemeType:    "comic",
				ColorPrimary: "#eb2f96",
				BorderRadius: 16,
				IsCompact:    false,
				IsEnabled:    true,
			},
		},
		{
			Name:        "ocean",
			DisplayName: "Ocean",
			Description: "Cool and calm ocean-inspired blue theme",
			ThemeData: &ThemeData{
				ThemeType:    "default",
				ColorPrimary: "#1890ff",
				BorderRadius: 8,
				IsCompact:    false,
				IsEnabled:    true,
			},
		},
		{
			Name:        "sunset",
			DisplayName: "Sunset",
			Description: "Warm sunset-inspired orange theme",
			ThemeData: &ThemeData{
				ThemeType:    "default",
				ColorPrimary: "#ff7a45",
				BorderRadius: 10,
				IsCompact:    false,
				IsEnabled:    true,
			},
		},
		{
			Name:        "forest",
			DisplayName: "Forest",
			Description: "Natural forest-inspired green theme",
			ThemeData: &ThemeData{
				ThemeType:    "default",
				ColorPrimary: "#52c41a",
				BorderRadius: 6,
				IsCompact:    false,
				IsEnabled:    true,
			},
		},
		{
			Name:        "corporate",
			DisplayName: "Corporate",
			Description: "Professional corporate blue with minimal design",
			ThemeData: &ThemeData{
				ThemeType:    "default",
				ColorPrimary: "#0050b3",
				BorderRadius: 2,
				IsCompact:    true,
				IsEnabled:    true,
			},
		},
	}
}

// GetBuiltInTheme returns a specific built-in theme by name
func GetBuiltInTheme(name string) *BuiltInTheme {
	themes := GetBuiltInThemes()
	for _, theme := range themes {
		if theme.Name == name {
			return theme
		}
	}
	return nil
}
