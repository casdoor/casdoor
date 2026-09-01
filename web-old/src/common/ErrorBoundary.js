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

// This component is rendered when the app is already broken, so it deliberately
// uses plain elements and inline styles instead of antd or i18next, to make sure
// it can still be shown when one of them is the thing that failed.
class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
    };
  }

  static getDerivedStateFromError(error) {
    return {error: error};
  }

  componentDidCatch(error, errorInfo) {
    // eslint-disable-next-line no-console
    console.error(error, errorInfo);
  }

  render() {
    if (this.state.error === null) {
      return this.props.children;
    }

    return (
      <div style={{maxWidth: "560px", margin: "15vh auto 0 auto", padding: "0 24px", textAlign: "center", lineHeight: 1.6, color: "#262626"}}>
        <div style={{fontSize: "20px", fontWeight: 600}}>
          Something went wrong
        </div>
        <div style={{marginTop: "12px", fontSize: "14px", color: "#595959"}}>
          Casdoor could not render this page. Please reload the page, and try a recent version of Chrome, Edge, Firefox or Safari if the problem persists.
        </div>
        <div style={{marginTop: "12px", padding: "8px 12px", fontSize: "12px", fontFamily: "Consolas, Menlo, monospace", color: "#8c8c8c", background: "#fafafa", border: "1px solid #f0f0f0", borderRadius: "5px", wordBreak: "break-all", textAlign: "left"}}>
          {String(this.state.error && this.state.error.message ? this.state.error.message : this.state.error)}
        </div>
        <a href={window.location.pathname + window.location.search} style={{display: "inline-block", marginTop: "16px", fontSize: "14px"}}>
          Reload
        </a>
      </div>
    );
  }
}

export default ErrorBoundary;
