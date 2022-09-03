// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

const path = require("path");
const {generateTheme} = require("antd-theme-generator");

const options = {
  stylesDir: path.join(__dirname, "../src/assets"),
  antDir: path.join(__dirname, "../node_modules/antd"),
  varFile: path.join(__dirname, "../src/assets/style/variables.less"),
  mainLessFile: path.join(__dirname, "../src/index.less"),
  themeVariables: [
    "@primary-color",
    "@layout-body-background",
    "@layout-header-background",
    "@body-background",
    "@component-background",
    "@heading-color",
    "@text-color",
    "@text-color-inverse",
    "@text-color-secondary",
    "@shadow-color",
    "@border-color-split",
    "@background-color-light",
    "@background-color-base",
    "@checkbox-check-color",
    "@disabled-color",
    "@menu-dark-color",
    "@menu-dark-highlight-color",
    "@menu-dark-arrow-color",
    "@btn-primary-color",
    "@table-selected-row-bg",
  ],
  outputFilePath: path.join(__dirname, "../public/color.less"),
};

/* eslint-disable */
generateTheme(options)
  .then(less => {
    console.log("Theme generated successfully");
  })
  .catch(error => {
    console.log("Error", error);
  });
/* eslint-disable */
