import React from "react";
import { Link, withRouter } from "react-router-dom";
import { Breadcrumb } from "antd";
import {
  HomeOutlined,
  UserOutlined,
  ContainerOutlined,
  BankOutlined,
  AppstoreOutlined,
  ControlOutlined,
  ApiOutlined,
} from "@ant-design/icons";

const breadcrumbNameMap = {
  "/": "Home",
  "/account": "Account",
  "/organizations": "organizations",
  "/users": "Users",
  "/providers": "Providers",
  "/applications": "Applications",
  "/tokens": "Tokens",
  ":organizationName": ":organizationName",
};

export const Nav = withRouter((props) => {
  const { location } = props;
  const pathSnippets = location.pathname.split("/").filter((i) => i);
  const extraBreadcrumbItems = pathSnippets.map((_, index) => {
    const url = `/${pathSnippets.slice(0, index + 1).join("/")}`;
    return (
      <Breadcrumb.Item key={url}>
        <Link to={url}></Link>
        {breadcrumbNameMap[url]}
      </Breadcrumb.Item>
    );
  });
  const breadcrumbItems = [
    <Breadcrumb.Item key="home">
      <Link to="/">Home</Link>
    </Breadcrumb.Item>,
  ].concat(extraBreadcrumbItems);
  return (
    <div className="breadcrumb">
      <Breadcrumb>{breadcrumbItems}</Breadcrumb>
    </div>
  );
});

export default Nav;
