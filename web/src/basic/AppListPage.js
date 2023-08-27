// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
import * as ApplicationBackend from "../backend/ApplicationBackend";
import GridCards from "./GridCards";

const AppListPage = (props) => {
  const [applications, setApplications] = React.useState(null);

  React.useEffect(() => {
    if (props.account === null) {
      return;
    }
    ApplicationBackend.getApplicationsByOrganization("admin", props.account.owner)
      .then((res) => {
        setApplications(res.data || []);
      });
  }, [props.account]);

  const getItems = () => {
    if (applications === null) {
      return null;
    }

    return applications.map(application => {
      let homepageUrl = application.homepageUrl;
      if (homepageUrl === "<custom-url>") {
        homepageUrl = props.account.homepage;
      }

      return {
        link: homepageUrl,
        name: application.displayName,
        description: application.description,
        logo: application.logo,
        createdTime: "",
      };
    });

  };

  return (
    <div style={{display: "flex", justifyContent: "center", flexDirection: "column", alignItems: "center"}}>
      <GridCards items={getItems()} />
    </div>
  );
};

export default AppListPage;
