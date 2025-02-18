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

import {Card, Row, Spin} from "antd";
import i18next from "i18next";
import React from "react";
import * as Setting from "../Setting";
import SingleCard from "./SingleCard";

const GridCards = (props) => {
  const items = props.items;

  if (items === null || items === undefined) {
    return (
      <div style={{display: "flex", justifyContent: "center", alignItems: "center", marginTop: "10%"}}>
        <Spin size="large" tip={i18next.t("login:Loading")} style={{paddingTop: "10%"}} />
      </div>
    );
  }

  return (
    Setting.isMobile() ? (
      <Card styles={{body: {padding: 0}}}>
        {items.map(item => <SingleCard key={item.link} logo={item.logo} link={item.link} title={item.name} desc={item.description} isSingle={items.length === 1} />)}
      </Card>
    ) : (
      <div style={{margin: "0 15px"}}>
        <Row>
          {items.map(item => <SingleCard logo={item.logo} link={item.link} title={item.name} desc={item.description} time={item.createdTime} isSingle={items.length === 1} key={item.name} />)}
        </Row>
      </div>
    )
  );
};

export default GridCards;
