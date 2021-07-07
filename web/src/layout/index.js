import React from "react";
import { Component } from "react";
import { Layout, Menu, Breadcrumb, Avatar } from "antd";
import * as Setting from "../Setting.js";
import CustomGithubCorner from "../CustomGithubCorner";
import { Link, withRouter } from "react-router-dom";
import i18next from "i18next";
import { Nav } from "../component/breadcrumb";
import SizeContext from "antd/lib/config-provider/SizeContext";
import * as conf from "../common/Conf.js";
const { Header, Footer, Sider, Content } = Layout;

const SubMenu = Menu.SubMenu;

export class BasicSider extends Component {
  state = {
    collapsed: false,
  };

  onCollapse = (collapsed) => {
    this.setState({ collapsed });
  };

  render() {
    const { collapsed } = this.state;
    return (
      <Sider
        breakpoint="sm"
        collapsible
        collapsed={collapsed}
        onCollapse={this.onCollapse}
        theme="light"
      >
        <div className="brand">
          <div className="siderLogo">
            <img alt="logo" src={conf.logoPath} height="40px" />
            {!collapsed && <span>{conf.siteName}</span>}
          </div>
        </div>
        <div>
          <Menu
            theme="light"
            mode="inline"
            selectedKeys={this.props.path}
            defaultSelectedKeys={["0"]}
          >
            {this.props.children}s
          </Menu>
        </div>
      </Sider>
    );
  }
}

export class BasicHeader extends Component {
  render() {
    return (
      <Header style={{ background: "#fff", textAlign: "right", padding: 0 }}>
        <div>{this.props.children}</div>
        <CustomGithubCorner />
      </Header>
    );
  }
}

export class BasicContent extends Component {
  render() {
    return (
      <Content style={{ margin: "0px 15px" }}>
        <Nav />
        <div style={{ padding: 24, background: "#fff", minHeight: 360 }}>
          {this.props.children}
        </div>
      </Content>
    );
  }
}
