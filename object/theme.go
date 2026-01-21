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
	Name                    string     `json:"name"`
	DisplayName             string     `json:"displayName"`
	Description             string     `json:"description"`
	ThemeData               *ThemeData `json:"themeData"`
	FormOffset              int        `json:"formOffset"`
	FormBackgroundUrl       string     `json:"formBackgroundUrl"`
	FormBackgroundUrlMobile string     `json:"formBackgroundUrlMobile"`
	FormCss                 string     `json:"formCss"`
	FormCssMobile           string     `json:"formCssMobile"`
}

// GetBuiltInThemes returns a list of predefined beautiful theme configurations
func GetBuiltInThemes() []*BuiltInTheme {
	return []*BuiltInTheme{
		{
			Name:        "default",
			DisplayName: "Default",
			Description: "Casdoor's default theme with purple primary color and centered layout",
			ThemeData: &ThemeData{
				ThemeType:    "default",
				ColorPrimary: "#5734d3",
				BorderRadius: 6,
				IsCompact:    false,
				IsEnabled:    true,
			},
			FormOffset:              2,
			FormBackgroundUrl:       "https://cdn.casbin.org/img/casdoor-login-bg.png",
			FormBackgroundUrlMobile: "",
			FormCss: `<style>
  .login-panel {
    padding: 40px 70px 0 70px;
    border-radius: 10px;
    background-color: #ffffff;
    box-shadow: 0 0 30px 20px rgba(0, 0, 0, 0.20);
  }
</style>`,
			FormCssMobile: "",
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
			FormOffset:              2,
			FormBackgroundUrl:       "https://cdn.casbin.org/img/casdoor-login-bg-dark.png",
			FormBackgroundUrlMobile: "",
			FormCss: `<style>
  .login-panel-dark {
    padding: 40px 70px 0 70px;
    border-radius: 10px;
    background-color: #1f1f1f;
    box-shadow: 0 0 30px 20px rgba(255, 255, 255, 0.15);
  }
</style>`,
			FormCssMobile: "",
		},
		{
			Name:        "lark",
			DisplayName: "Document",
			Description: "Professional document-style theme with green accents and side panel",
			ThemeData: &ThemeData{
				ThemeType:    "lark",
				ColorPrimary: "#00b96b",
				BorderRadius: 4,
				IsCompact:    false,
				IsEnabled:    true,
			},
			FormOffset:              4,
			FormBackgroundUrl:       "",
			FormBackgroundUrlMobile: "",
			FormCss: `<style>
  .login-panel {
    padding: 40px 70px 0 70px;
    border-radius: 8px;
    background-color: #ffffff;
    box-shadow: 0 2px 16px rgba(0, 0, 0, 0.12);
  }
</style>`,
			FormCssMobile: "",
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
			FormOffset:              2,
			FormBackgroundUrl:       "https://cdn.casbin.org/img/casdoor-login-bg-blossom.png",
			FormBackgroundUrlMobile: "",
			FormCss: `<style>
  .login-panel {
    padding: 50px 80px 0 80px;
    border-radius: 20px;
    background: linear-gradient(135deg, #ffffff 0%, #fff5f8 100%);
    box-shadow: 0 8px 32px rgba(235, 47, 150, 0.15);
  }
</style>`,
			FormCssMobile: "",
		},
		{
			Name:        "ocean",
			DisplayName: "Ocean",
			Description: "Cool and calm ocean-inspired blue theme with gradient background",
			ThemeData: &ThemeData{
				ThemeType:    "default",
				ColorPrimary: "#1890ff",
				BorderRadius: 8,
				IsCompact:    false,
				IsEnabled:    true,
			},
			FormOffset:              2,
			FormBackgroundUrl:       "https://cdn.casbin.org/img/casdoor-login-bg-ocean.png",
			FormBackgroundUrlMobile: "",
			FormCss: `<style>
  .login-panel {
    padding: 45px 75px 0 75px;
    border-radius: 12px;
    background: linear-gradient(135deg, #ffffff 0%, #e6f7ff 100%);
    box-shadow: 0 4px 24px rgba(24, 144, 255, 0.20);
  }
</style>`,
			FormCssMobile: "",
		},
		{
			Name:        "sunset",
			DisplayName: "Sunset",
			Description: "Warm sunset-inspired orange theme with vibrant colors",
			ThemeData: &ThemeData{
				ThemeType:    "default",
				ColorPrimary: "#ff7a45",
				BorderRadius: 10,
				IsCompact:    false,
				IsEnabled:    true,
			},
			FormOffset:              2,
			FormBackgroundUrl:       "https://cdn.casbin.org/img/casdoor-login-bg-sunset.png",
			FormBackgroundUrlMobile: "",
			FormCss: `<style>
  .login-panel {
    padding: 45px 75px 0 75px;
    border-radius: 14px;
    background: linear-gradient(135deg, #ffffff 0%, #fff7e6 100%);
    box-shadow: 0 6px 28px rgba(255, 122, 69, 0.18);
  }
</style>`,
			FormCssMobile: "",
		},
		{
			Name:        "forest",
			DisplayName: "Forest",
			Description: "Natural forest-inspired green theme with organic feel",
			ThemeData: &ThemeData{
				ThemeType:    "default",
				ColorPrimary: "#52c41a",
				BorderRadius: 6,
				IsCompact:    false,
				IsEnabled:    true,
			},
			FormOffset:              2,
			FormBackgroundUrl:       "https://cdn.casbin.org/img/casdoor-login-bg-forest.png",
			FormBackgroundUrlMobile: "",
			FormCss: `<style>
  .login-panel {
    padding: 40px 70px 0 70px;
    border-radius: 10px;
    background: linear-gradient(135deg, #ffffff 0%, #f6ffed 100%);
    box-shadow: 0 4px 24px rgba(82, 196, 26, 0.15);
  }
</style>`,
			FormCssMobile: "",
		},
		{
			Name:        "corporate",
			DisplayName: "Corporate",
			Description: "Professional corporate blue with minimal design and left-aligned form",
			ThemeData: &ThemeData{
				ThemeType:    "default",
				ColorPrimary: "#0050b3",
				BorderRadius: 2,
				IsCompact:    true,
				IsEnabled:    true,
			},
			FormOffset:              0,
			FormBackgroundUrl:       "",
			FormBackgroundUrlMobile: "",
			FormCss: `<style>
  .login-panel {
    padding: 35px 60px 0 60px;
    border-radius: 4px;
    background-color: #ffffff;
    box-shadow: 0 1px 8px rgba(0, 0, 0, 0.12);
    border: 1px solid #e8e8e8;
  }
</style>`,
			FormCssMobile: "",
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
