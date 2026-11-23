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

/** The signals of adaptive risk with their default scores, as the server has them. */
export const RiskSignals: {name: string; score: number; labelKey: string}[] = [
  {name: "new-device", score: 20, labelKey: "enterprise:New device"},
  {name: "new-location", score: 30, labelKey: "enterprise:New location"},
  {name: "new-country", score: 40, labelKey: "organization:New country"},
  {name: "impossible-travel", score: 70, labelKey: "organization:Impossible travel"},
  {name: "ip-failures", score: 40, labelKey: "organization:Many accounts failed from the IP"},
  {name: "user-failures", score: 30, labelKey: "organization:Repeated wrong passwords"},
  {name: "suspicious-client", score: 20, labelKey: "organization:Script or headless browser"},
  {name: "untrusted-device", score: 40, labelKey: "organization:Untrusted device"},
  {name: "blocked-country", score: 100, labelKey: "organization:Blocked country"},
];
