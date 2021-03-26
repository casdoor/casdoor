import './App.css';
import FormComponent from './Components/Form';
import Navbar from './Components/Navbar';
import { Layout } from "antd";

const {Header, Content} = Layout
function App() {
  return (
    <>
      <Header>
        <Navbar/>
      </Header>
      <Content><FormComponent/></Content>
    </>
  );
}

export default App;
