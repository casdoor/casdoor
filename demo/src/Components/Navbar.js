import { Menu } from "antd";
import React, { useState } from "react";

const Navbar = () => {
  const [current, setCurrent] = useState("login");
  const handleClick = (e) => {
    console.log("click ", e);
    setCurrent(e.key );
  };

  return (
    <Menu onClick={handleClick} selectedKeys={current} mode="horizontal" style={{margin: 0}}>
      <Menu.Item key="login">Login</Menu.Item>
      <Menu.Item key="signup">Register</Menu.Item>
    </Menu>
  );
};

export default Navbar;
