import React, {useState} from "react";
import github_logo from '../assets/github_logo.png'
import google_logo from '../assets/google_logo.png'
import facebook_logo from '../assets/facebook_logo.png'
import { Form } from "antd";


const providers = [
    {
        img : github_logo,
        displayName : "Github",
        url : "/"
    },
    {
        img : google_logo,
        displayName : "Google",
        url : "/"
    },
    {
        img : facebook_logo,
        displayName : "Facebook",
        url : "/"
    }
]

function ProviderLogin() {
    const [p] = useState(providers)
  return (
    <Form.Item>
      {p.map((provider) => {
        return (
          <a href={"/signup"}>
            <img
              width={40}
              height={40}
              src={provider.img}
              alt={provider.displayName}
              style={{ margin: "3px" }}
            />
          </a>
        );
      })}
    </Form.Item>
  );
}

export default ProviderLogin;
