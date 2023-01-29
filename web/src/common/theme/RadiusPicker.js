// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import {InputNumber, Slider, Space} from "antd";

export default function RadiusPicker({value, onChange}) {
  return (
    <Space size="large">
      <InputNumber
        value={value}
        onChange={onChange}
        style={{width: 120}}
        min={0}
        formatter={(val) => `${val}px`}
        parser={(str) => (str ? parseFloat(str) : str)}
      />

      <Slider
        tooltip={{open: false}}
        style={{width: 128}}
        min={0}
        value={value}
        max={20}
        onChange={onChange}
      />
    </Space>
  );
}
