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
import {Alert, Button, Card, Descriptions} from "antd";
import {ArrowLeftOutlined} from "@ant-design/icons";
import i18next from "i18next";
import * as EntryBackend from "./backend/EntryBackend";
import * as Setting from "./Setting";
import Loading from "./common/Loading";
import {Editor} from "./common/Editor";

class OpenClawSessionTranscriptPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      owner: props.match.params.organizationName,
      entryName: props.match.params.entryName,
      loading: false,
      error: "",
      transcript: null,
    };
  }

  componentDidMount() {
    this.loadTranscript();
  }

  componentDidUpdate(prevProps) {
    const owner = this.props.match.params.organizationName;
    const entryName = this.props.match.params.entryName;
    if (owner !== prevProps.match.params.organizationName || entryName !== prevProps.match.params.entryName) {
      this.setState({
        owner,
        entryName,
        transcript: null,
      }, () => this.loadTranscript());
    }
  }

  loadTranscript() {
    this.setState({loading: true, error: ""});
    EntryBackend.getOpenClawSessionTranscript(this.state.owner, this.state.entryName)
      .then((res) => {
        if (res.status === "ok" && res.data) {
          this.setState({
            loading: false,
            error: "",
            transcript: res.data,
          });
        } else {
          this.setState({
            loading: false,
            error: `${i18next.t("general:Failed to load")}: ${res.msg}`,
            transcript: null,
          });
        }
      })
      .catch((error) => {
        this.setState({
          loading: false,
          error: `${i18next.t("general:Failed to load")}: ${error?.message || String(error)}`,
          transcript: null,
        });
      });
  }

  getEntryPath() {
    return `/entries/${this.state.owner}/${encodeURIComponent(this.state.entryName)}`;
  }

  renderContent() {
    if (this.state.loading) {
      return <Loading type="page" tip={i18next.t("login:Loading")} />;
    }

    if (this.state.error) {
      return <Alert type="warning" showIcon message={this.state.error} />;
    }

    const transcript = this.state.transcript;
    if (!transcript) {
      return null;
    }

    return (
      <div style={{display: "grid", gap: 12}}>
        <Descriptions bordered size="small" column={Setting.isMobile() ? 1 : 3}>
          <Descriptions.Item label={i18next.t("resource:File name")}>
            {transcript.fileName || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("resource:File size")}>
            {Setting.getFriendlyFileSize(transcript.fileSize || 0)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Loaded size")}>
            {Setting.getFriendlyFileSize(transcript.loadedSize || 0)}
          </Descriptions.Item>
        </Descriptions>
        {transcript.truncated ? (
          <Alert
            type="warning"
            showIcon
            message={i18next.t("entry:Transcript truncated")}
          />
        ) : null}
        <Editor
          value={transcript.content || ""}
          readOnly
          fillWidth
          dark
          height="calc(100vh - 300px)"
          minHeight="360px"
        />
      </div>
    );
  }

  render() {
    return (
      <Card
        size="small"
        title={i18next.t("entry:Raw JSONL")}
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => this.props.history.push(this.getEntryPath())}
          >
            {i18next.t("entry:Back to session")}
          </Button>
        }
        style={Setting.isMobile() ? {margin: "5px"} : {}}
      >
        {this.renderContent()}
      </Card>
    );
  }
}

export default OpenClawSessionTranscriptPage;
