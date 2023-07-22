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

import {Button, Result} from "antd";
import i18next from "i18next";
import React, {useEffect} from "react";

export const NotFoundResult = ({onUpdateApplication}) => {
  useEffect(() => {
    if (onUpdateApplication !== undefined) {
      onUpdateApplication(null);
    }
  }, [onUpdateApplication]);

  return (
    <Result
      style={{margin: "0 auto"}}
      status="404"
      title="404 NOT FOUND"
      subTitle={i18next.t("general:Sorry, the page you visited does not exist.")}
      extra={<Button type="primary" href={"/"}>{i18next.t("general:Back Home")}</Button>}
    />
  );
};

export default NotFoundResult;
