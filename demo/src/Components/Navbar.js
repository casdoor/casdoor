import { Menu, Avatar, Badge } from "antd";
import React, { useState } from "react";
import { UserOutlined } from '@ant-design/icons';

const Render = (props)=>{
  const [select, setSelect] = useState("login");

  const handleClick = (e) => {
    console.log("click ", e);
    setSelect(e.key );
  };

  if(props.logged === true){
    return (<Menu onClick={handleClick} selectedKeys={select} mode="horizontal">
      <Menu.Item key="avatar"><Avatar size={"medium"} icon={<UserOutlined />} />{"  User Name "}</Menu.Item>
    </Menu>)
  }
  return (
    <Menu onClick={handleClick} selectedKeys={select} mode="horizontal">
      <Menu.Item key="login">Login</Menu.Item>
      <Menu.Item key="signup">Register</Menu.Item>
    </Menu>
  )
}

const Navbar = (props) => {
  

  return (
    <div>
      <Render logged={props.logged}/>
    </div>
  );
};

export default Navbar;
