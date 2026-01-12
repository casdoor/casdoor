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

// BuiltInThemes contains predefined theme configurations
// This is migrated from backend object/theme.go to reduce API calls

export const builtInThemes = [
  {
    name: "default",
    displayName: "Default",
    description: "Casdoor's default theme with purple primary color and centered layout",
    themeData: {
      themeType: "default",
      colorPrimary: "#5734d3",
      borderRadius: 6,
      isCompact: false,
      isEnabled: true,
    },
    formOffset: 2,
    formBackgroundUrl: "https://cdn.casbin.org/img/casdoor-login-bg.png",
    formBackgroundUrlMobile: "",
    formCss: `<style>
  .login-panel {
    padding: 40px 70px 0 70px;
    border-radius: 10px;
    background-color: #ffffff;
    box-shadow: 0 0 30px 20px rgba(0, 0, 0, 0.20);
  }
</style>`,
    formCssMobile: "",
  },
  {
    name: "dark",
    displayName: "Dark",
    description: "Dark theme for better viewing in low-light environments",
    themeData: {
      themeType: "dark",
      colorPrimary: "#5734d3",
      borderRadius: 2,
      isCompact: false,
      isEnabled: true,
    },
    formOffset: 2,
    formBackgroundUrl: "https://cdn.casbin.org/img/casdoor-login-bg-dark.png",
    formBackgroundUrlMobile: "",
    formCss: `<style>
  .login-panel-dark {
    padding: 40px 70px 0 70px;
    border-radius: 10px;
    background-color: #1f1f1f;
    box-shadow: 0 0 30px 20px rgba(255, 255, 255, 0.15);
  }
</style>`,
    formCssMobile: "",
  },
  {
    name: "lark",
    displayName: "Document",
    description: "Professional document-style theme with green accents and side panel",
    themeData: {
      themeType: "lark",
      colorPrimary: "#00b96b",
      borderRadius: 4,
      isCompact: false,
      isEnabled: true,
    },
    formOffset: 4,
    formBackgroundUrl: "",
    formBackgroundUrlMobile: "",
    formCss: `<style>
  .login-panel {
    padding: 40px 70px 0 70px;
    border-radius: 8px;
    background-color: #ffffff;
    box-shadow: 0 2px 16px rgba(0, 0, 0, 0.12);
  }
</style>`,
    formCssMobile: "",
  },
  {
    name: "comic",
    displayName: "Blossom",
    description: "Playful and friendly theme with pink accents and rounded corners",
    themeData: {
      themeType: "comic",
      colorPrimary: "#eb2f96",
      borderRadius: 16,
      isCompact: false,
      isEnabled: true,
    },
    formOffset: 2,
    formBackgroundUrl: "https://cdn.casbin.org/img/casdoor-login-bg-blossom.png",
    formBackgroundUrlMobile: "",
    formCss: `<style>
  .login-panel {
    padding: 50px 80px 0 80px;
    border-radius: 20px;
    background: linear-gradient(135deg, #ffffff 0%, #fff5f8 100%);
    box-shadow: 0 8px 32px rgba(235, 47, 150, 0.15);
  }
</style>`,
    formCssMobile: "",
  },
  {
    name: "ocean",
    displayName: "Ocean",
    description: "Cool and calm ocean-inspired blue theme with gradient background",
    themeData: {
      themeType: "default",
      colorPrimary: "#1890ff",
      borderRadius: 8,
      isCompact: false,
      isEnabled: true,
    },
    formOffset: 2,
    formBackgroundUrl: "https://cdn.casbin.org/img/casdoor-login-bg-ocean.png",
    formBackgroundUrlMobile: "",
    formCss: `<style>
  .login-panel {
    padding: 45px 75px 0 75px;
    border-radius: 12px;
    background: linear-gradient(135deg, #ffffff 0%, #e6f7ff 100%);
    box-shadow: 0 4px 24px rgba(24, 144, 255, 0.20);
  }
</style>`,
    formCssMobile: "",
  },
  {
    name: "sunset",
    displayName: "Sunset",
    description: "Warm sunset-inspired orange theme with vibrant colors",
    themeData: {
      themeType: "default",
      colorPrimary: "#ff7a45",
      borderRadius: 10,
      isCompact: false,
      isEnabled: true,
    },
    formOffset: 2,
    formBackgroundUrl: "https://cdn.casbin.org/img/casdoor-login-bg-sunset.png",
    formBackgroundUrlMobile: "",
    formCss: `<style>
  .login-panel {
    padding: 45px 75px 0 75px;
    border-radius: 14px;
    background: linear-gradient(135deg, #ffffff 0%, #fff7e6 100%);
    box-shadow: 0 6px 28px rgba(255, 122, 69, 0.18);
  }
</style>`,
    formCssMobile: "",
  },
  {
    name: "forest",
    displayName: "Forest",
    description: "Natural forest-inspired green theme with organic feel",
    themeData: {
      themeType: "default",
      colorPrimary: "#52c41a",
      borderRadius: 6,
      isCompact: false,
      isEnabled: true,
    },
    formOffset: 2,
    formBackgroundUrl: "https://cdn.casbin.org/img/casdoor-login-bg-forest.png",
    formBackgroundUrlMobile: "",
    formCss: `<style>
  .login-panel {
    padding: 40px 70px 0 70px;
    border-radius: 10px;
    background: linear-gradient(135deg, #ffffff 0%, #f6ffed 100%);
    box-shadow: 0 4px 24px rgba(82, 196, 26, 0.15);
  }
</style>`,
    formCssMobile: "",
  },
  {
    name: "corporate",
    displayName: "Corporate",
    description: "Professional corporate blue with minimal design and left-aligned form",
    themeData: {
      themeType: "default",
      colorPrimary: "#0050b3",
      borderRadius: 2,
      isCompact: true,
      isEnabled: true,
    },
    formOffset: 0,
    formBackgroundUrl: "",
    formBackgroundUrlMobile: "",
    formCss: `<style>
  .login-panel {
    padding: 35px 60px 0 60px;
    border-radius: 4px;
    background-color: #ffffff;
    box-shadow: 0 1px 8px rgba(0, 0, 0, 0.12);
    border: 1px solid #e8e8e8;
  }
</style>`,
    formCssMobile: "",
  },
];

// Get a specific built-in theme by name
export function getBuiltInThemeByName(name) {
  return builtInThemes.find((theme) => theme.name === name) || null;
}
