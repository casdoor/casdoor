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
import i18next from "i18next";
import {Tag} from "antd";

const AppListPage = (props) => {
  const [applications, setApplications] = React.useState(null);
  const [selectedTags, setSelectedTags] = React.useState([]);
  const [allTags, setAllTags] = React.useState([]);

  const sort = (applications) => {
    return [...applications].sort((a, b) => a.order - b.order);
  };

  const extractTags = (applications) => {
    const tagsSet = new Set();
    applications.forEach(application => {
      if (application.tags && Array.isArray(application.tags)) {
        application.tags.forEach(tag => tagsSet.add(tag));
      }
    });
    return Array.from(tagsSet);
  };

  React.useEffect(() => {
    if (props.account === null) {
      return;
    }
    ApplicationBackend.getApplicationsByOrganization("admin", props.account.owner)
      .then((res) => {
        const applications = res.data || [];
        const sortedApps = sort(applications);
        setApplications(sortedApps);
        setAllTags(extractTags(sortedApps));
      });
  }, [props.account]);

  const handleTagChange = (tag, checked) => {
    setSelectedTags(prev =>
      checked
        ? [...prev, tag]
        : prev.filter(t => t !== tag)
    );
  };

  const filterByTags = (applications) => {
    if (selectedTags.length === 0) {return applications;}

    return applications.filter(application => {
      if (!application.tags || !Array.isArray(application.tags)) {return false;}

      return selectedTags.every(tag => application.tags.includes(tag));
    });
  };

  const generateTagColor = (tag) => {
    const colors = [
      "#ff4d4f", "#f5222d", "#ff7a45", "#fa541c",
      "#ffa940", "#fa8c16", "#ffc53d", "#faad14",
      "#ffec3d", "#fadb14", "#bae637", "#a0d911",
      "#73d13d", "#52c41a", "#36cfc9", "#13c2c2",
      "#40a9ff", "#1890ff", "#f759ab", "#eb2f96",
    ];
    let hash = 5381;
    for (let i = 0; i < tag.length; i++) {
      hash = (hash * 33) ^ tag.charCodeAt(i);
    }
    return colors[Math.abs(hash) % colors.length];
  };

  const getItems = () => {
    if (applications === null) {
      return null;
    }

    const filteredApps = filterByTags(applications);

    return filteredApps.map(application => {
      let homepageUrl = application.homepageUrl;
      if (homepageUrl === "<custom-url>") {
        homepageUrl = props.account.homepage;
      }

      const tagObjects = application.tags ? application.tags.map(tag => ({
        name: tag,
        color: generateTagColor(tag),
      })) : [];

      return {
        link: homepageUrl,
        name: application.displayName,
        description: application.description,
        logo: application.logo,
        createdTime: "",
        tags: tagObjects,
      };
    });
  };

  const TagFilterArea = () => {
    return (
      <div style={{marginBottom: "20px", display: "flex", flexWrap: "wrap", gap: "8px"}}>
        <span style={{marginRight: "8px", fontWeight: "bold"}}>{i18next.t("organization:Tags")}</span>
        {allTags.map(tag => (
          <Tag.CheckableTag
            key={tag}
            checked={selectedTags.includes(tag)}
            onChange={(checked) => handleTagChange(tag, checked)}
            style={{backgroundColor: selectedTags.includes(tag) ? generateTagColor(tag) : "white", borderColor: generateTagColor(tag)}}
          >
            {tag}
          </Tag.CheckableTag>
        ))}

        {selectedTags.length > 0 && (
          <button
            onClick={() => setSelectedTags([])}
            style={{marginLeft: "10px", padding: "2px 8px", background: "#ffffff", border: "2px solid #ddd", borderRadius: "4px", cursor: "pointer"}}
          >
            {i18next.t("forget:Reset")}
          </button>
        )}
      </div>
    );
  };

  return (
    <div style={{padding: "20px"}}>
      {allTags.length > 0 && TagFilterArea()}
      <div style={{display: "flex", justifyContent: "center", flexDirection: "column", alignItems: "center"}}>
        <GridCards items={getItems()} />
      </div>
    </div>
  );
};

export default AppListPage;
