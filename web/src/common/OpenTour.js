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

import React from "react";
import * as TourConfig from "../TourConfig";
import * as Setting from "../Setting";
import {Tooltip} from "antd";
import "../App.less";
import {QuestionOutlined} from "@ant-design/icons";

class OpenTour extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      isTourVisible: props.isTourVisible ?? TourConfig.getTourVisible(),
    };
  }

  render() {
    return (
      <Tooltip title="You can open the tour.">
        <div className="select-box" style={{display: Setting.isMobile() ? "none" : null, ...this.props.style}} >
          <QuestionOutlined style={{fontSize: "24px", color: "#4d4d4d"}} onClick={() => TourConfig.setIsTourVisible(true)} />
        </div>
      </Tooltip>
    );
  }
}

export default OpenTour;
