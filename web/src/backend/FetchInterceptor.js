import {Modal} from "antd";
import {ExclamationCircleFilled} from "@ant-design/icons";
import * as Conf from "../Conf";

const {confirm} = Modal;
const {fetch: originalFetch} = window;

/**
 * When modify data, prompt it's read-only and ask whether to go writable site
 */
const demoModePrompt = async(url, option) => {
  if (option.method === "POST") {
    confirm({
      title: "This is a read-only demo site!",
      icon: <ExclamationCircleFilled />,
      content: "Go Writable site demo?",
      onOk() {
        window.open("https://demo.casdoor.com", "_blank");
      },
      onCancel() {},
    });
  }
  return option;
};

const requsetInterceptors = [];
const responseInterceptors = [];

// when it's in DemoMode, demoModePrompt() should run before fetch
if (Conf.IsDemoMode) {
  requsetInterceptors.push(demoModePrompt);
}

/**
 * rewrite fetch to support interceptors
 */
window.fetch = async(url, option = {}) => {
  for (const fn of requsetInterceptors) {
    fn(url, option);
  }

  const response = await originalFetch(url, option);
  responseInterceptors.forEach(fn => (response) => fn(response));
  return response;
};
