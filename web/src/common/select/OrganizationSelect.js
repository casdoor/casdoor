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
import {Select} from "antd";
import i18next from "i18next";
import * as OrganizationBackend from "../../backend/OrganizationBackend";
import * as Setting from "../../Setting";

function OrganizationSelect(props) {
  const {onChange, initValue, style, onSelect} = props;
  const [organizations, setOrganizations] = React.useState([]);
  const [value, setValue] = React.useState(initValue);

  React.useEffect(() => {
    if (props.organizations === undefined) {
      getOrganizations();
    }
  }, []);

  const getOrganizations = () => {
    OrganizationBackend.getOrganizationNames("admin")
      .then((res) => {
        if (res.status === "ok") {
          setOrganizations(res.data);
          if (initValue === undefined) {
            setValue(organizations.length > 0 ? organizations[0] : "");
          }
        }
      });
  };

  const handleOnChange = (value) => {
    setValue(value);
    onChange?.(value);
  };

  return (
    <Select
      options={organizations.map((organization) => Setting.getOption(organization.name, organization.name))}
      virtual={false}
      placeholder={i18next.t("login:Please select an organization")}
      value={value}
      onChange={handleOnChange}
      filterOption={(input, option) => (option?.label ?? "").toLowerCase().includes(input.toLowerCase())}
      style={style}
      onSelect={onSelect}
    >
    </Select>
  );
}

export default OrganizationSelect;
