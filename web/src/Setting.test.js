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

import {getProviderLogoURL, OtherProviderInfo} from "./Setting";

describe("ID Verification Provider", () => {
  test("OtherProviderInfo contains ID Verification category", () => {
    expect(OtherProviderInfo["ID Verification"]).toBeDefined();
  });

  test("OtherProviderInfo has Jumio provider", () => {
    expect(OtherProviderInfo["ID Verification"]["Jumio"]).toBeDefined();
    expect(OtherProviderInfo["ID Verification"]["Jumio"].logo).toBeDefined();
    expect(OtherProviderInfo["ID Verification"]["Jumio"].url).toBe("https://www.jumio.com/");
  });

  test("OtherProviderInfo has Alibaba Cloud provider", () => {
    expect(OtherProviderInfo["ID Verification"]["Alibaba Cloud"]).toBeDefined();
    expect(OtherProviderInfo["ID Verification"]["Alibaba Cloud"].logo).toBeDefined();
    expect(OtherProviderInfo["ID Verification"]["Alibaba Cloud"].url).toBe("https://www.aliyun.com/product/idverification");
  });

  test("getProviderLogoURL returns correct URL for Jumio provider", () => {
    const provider = {
      category: "ID Verification",
      type: "Jumio",
    };
    const logoUrl = getProviderLogoURL(provider);
    expect(logoUrl).toBeDefined();
    expect(logoUrl).toContain("social_default.png");
  });

  test("getProviderLogoURL returns correct URL for Alibaba Cloud provider", () => {
    const provider = {
      category: "ID Verification",
      type: "Alibaba Cloud",
    };
    const logoUrl = getProviderLogoURL(provider);
    expect(logoUrl).toBeDefined();
    expect(logoUrl).toContain("social_aliyun.png");
  });

  test("getProviderLogoURL does not crash for ID Verification providers", () => {
    const providers = [
      {category: "ID Verification", type: "Jumio"},
      {category: "ID Verification", type: "Alibaba Cloud"},
    ];
    
    providers.forEach(provider => {
      expect(() => getProviderLogoURL(provider)).not.toThrow();
    });
  });
});
