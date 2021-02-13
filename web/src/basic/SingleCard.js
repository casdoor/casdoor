// Copyright 2021 The casbin Authors. All Rights Reserved.
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
import {Card, Col, Row} from "antd";
import * as Setting from "../Setting";

const { Meta } = Card;

class SingleCard extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  renderCardMobile(logo, link, title, desc, time, isSingle) {
    const gridStyle = {
      width: '100vw',
      textAlign: 'center',
      cursor: 'pointer',
    };

    return (
      <Card.Grid style={gridStyle} onClick={() => Setting.goToLink(link)}>
        <img src={logo} alt="logo" height={60} style={{marginBottom: '20px'}}/>
        <Meta
          title={title}
          description={desc}
        />
      </Card.Grid>
    )
  }

  renderCard(logo, link, title, desc, time, isSingle) {
    return (
      <Col style={{paddingLeft: "20px", paddingRight: "20px", paddingBottom: "20px", marginBottom: "20px"}} span={6}>
        <Card
          hoverable
          cover={
            <img alt="logo" src={logo} width={"100%"} height={"100%"} />
          }
          onClick={() => Setting.goToLink(link)}
          style={isSingle ? {width: "320px"} : null}
        >
          <Meta title={title} description={desc} />
          <br/>
          <br/>
          <Meta title={""} description={Setting.getFormattedDateShort(time)} />
        </Card>
      </Col>
    )
  }

  render() {
    if (Setting.isMobile()) {
      return this.renderCardMobile(this.props.logo, this.props.link, this.props.title, this.props.desc, this.props.time, this.props.isSingle);
    } else {
      return this.renderCard(this.props.logo, this.props.link, this.props.title, this.props.desc, this.props.time, this.props.isSingle);
    }
  }
}

export default SingleCard;
