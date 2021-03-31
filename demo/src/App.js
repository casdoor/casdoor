import React, { useState } from "react";
import "./App.css";
import FormComponent from "./Components/Form";
import Navbar from "./Components/Navbar";
import { Layout } from "antd";
import HomePage from "./Pages/HomePage";

const { Header, Content } = Layout;

function App() {
  const [logged, setLogged] = useState(false);

  return (
    <div>
      <Header style={{ backgroundColor: "white" }}>
        <Navbar logged={logged}/>
      </Header>
      <Content>
        {logged===true ? <HomePage /> : <FormComponent/>}
      </Content>
    </div>
  );
}

export default App;
