// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

import {getFullServerUrl} from "./Setting";

describe("getFullServerUrl", () => {
  const originalLocation = window.location;

  beforeEach(() => {
    delete window.location;
  });

  afterEach(() => {
    window.location = originalLocation;
  });

  test("should redirect from development server port 7001 to backend port 8000", () => {
    window.location = {origin: "http://localhost:7001"};
    expect(getFullServerUrl()).toBe("http://localhost:8000");
  });

  test("should use window.location.origin when running on port 8000", () => {
    window.location = {origin: "http://localhost:8000"};
    expect(getFullServerUrl()).toBe("http://localhost:8000");
  });

  test("should use window.location.origin when running on custom port 6000", () => {
    window.location = {origin: "http://localhost:6000"};
    expect(getFullServerUrl()).toBe("http://localhost:6000");
  });

  test("should use window.location.origin when running on production domain", () => {
    window.location = {origin: "https://example.com"};
    expect(getFullServerUrl()).toBe("https://example.com");
  });

  test("should use window.location.origin when running on production domain with custom port", () => {
    window.location = {origin: "https://example.com:9000"};
    expect(getFullServerUrl()).toBe("https://example.com:9000");
  });
});
