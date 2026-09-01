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

import React from "react";
import {Badge, Button, InputNumber} from "antd";
import {MinusOutlined, PlusOutlined, ShoppingCartOutlined} from "@ant-design/icons";

export class QuantityStepper extends React.Component {
  render() {
    const {value, onIncrease, onDecrease, onChange, min = 1, max, disabled} = this.props;

    const parsedValue = (value === null || value === undefined || value === "") ? NaN : Number(value);
    const normalizedValue = Number.isFinite(parsedValue) ? parsedValue : min;

    return (
      <div style={{display: "inline-flex", alignItems: "center", border: "1px solid #d9d9d9", borderRadius: "6px", height: "36px", ...this.props.style}}>
        <Button
          type="text"
          size="small"
          icon={<MinusOutlined />}
          disabled={disabled || normalizedValue <= min}
          onClick={onDecrease}
          style={{borderRadius: "6px 0 0 6px", height: "100%", width: "calc(100% / 3)"}}
        />

        <InputNumber
          min={min}
          max={max}
          value={normalizedValue}
          onChange={onChange}
          controls={false}
          disabled={disabled}
          style={{
            width: "calc(100% / 3)",
            height: "100%",
            textAlign: "center",
            border: "none",
            boxShadow: "none",
            pointerEvents: onChange ? "auto" : "none",
            display: "flex",
            alignItems: "center",
          }}
        />

        <Button
          type="text"
          size="small"
          icon={<PlusOutlined />}
          disabled={disabled || (max !== undefined && normalizedValue >= max)}
          onClick={onIncrease}
          style={{borderRadius: "0 6px 6px 0", height: "100%", width: "calc(100% / 3)"}}
        />
      </div>
    );
  }
}

export class FloatingCartButton extends React.Component {
  render() {
    const {itemCount, onClick} = this.props;

    return (
      <div
        style={{
          position: "fixed",
          bottom: "50px",
          right: "50px",
          zIndex: 1000,
          cursor: "pointer",
        }}
        onClick={onClick}
      >
        <Badge count={itemCount} offset={[-5, 5]} size="default">
          <Button
            type="primary"
            shape="circle"
            icon={<ShoppingCartOutlined style={{fontSize: "24px"}} />}
            size="large"
            style={{width: "60px", height: "60px", boxShadow: "0 4px 8px rgba(0,0,0,0.15)"}}
          />
        </Badge>
      </div>
    );
  }
}
