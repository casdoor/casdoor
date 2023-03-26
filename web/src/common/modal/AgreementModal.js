// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import {Checkbox, Form, Modal} from "antd";
import i18next from "i18next";
import React, {useEffect, useState} from "react";

export const AgreementModal = (props) => {
  const {open, onOk, onCancel, application} = props;
  const [doc, setDoc] = useState("");

  useEffect(() => {
    getTermsOfUseContent(application.termsOfUseUrl).then((data) => {
      setDoc(data);
    });
  }, []);

  return (

    <Modal
      title={i18next.t("signup:Terms of Use")}
      open={open}
      width={"55vw"}
      closable={false}
      okText={i18next.t("signup:Accept")}
      cancelText={i18next.t("signup:Decline")}
      onOk={onOk}
      onCancel={onCancel}
    >
      <iframe title={"terms"} style={{border: 0, width: "100%", height: "60vh"}} srcDoc={doc} />
    </Modal>
  );
};

function getTermsOfUseContent(url) {
  return fetch(url, {
    method: "GET",
  }).then(r => r.text());
}

export function isAgreementRequired(application) {
  if (application) {
    const agreementItem = application.signupItems.find(item => item.name === "Agreement");
    if (!agreementItem || agreementItem.rule === "None" || !agreementItem.rule) {
      return false;
    }
    if (agreementItem.required) {
      return true;
    }
  }
  return false;
}

function initDefaultValue(application) {
  const agreementItem = application.signupItems.find(item => item.name === "Agreement");

  return isAgreementRequired(application) && agreementItem.rule === "Default True";
}

export function renderAgreementFormItem(application, required, layout, ths) {
  return (<React.Fragment>
    <Form.Item
      name="agreement"
      key="agreement"
      valuePropName="checked"
      rules={[
        {
          required: required,
        },
        () => ({
          validator: (_, value) => {
            if (!required) {
              return Promise.resolve();
            }

            if (!value) {
              return Promise.reject(i18next.t("signup:Please accept the agreement!"));
            } else {
              return Promise.resolve();
            }
          },
        }),
      ]
      }
      {...layout}
      initialValue={initDefaultValue(application)}
    >
      <Checkbox style={{float: "left"}}>
        {i18next.t("signup:Accept")}&nbsp;
        <a onClick={() => {
          ths.setState({
            isTermsOfUseVisible: true,
          });
        }}
        >
          {i18next.t("signup:Terms of Use")}
        </a>
      </Checkbox>
    </Form.Item>
    <AgreementModal application={application} layout={layout} open={ths.state.isTermsOfUseVisible}
      onOk={() => {
        ths.form.current.setFieldsValue({agreement: true});
        ths.setState({
          isTermsOfUseVisible: false,
        });
      }}
      onCancel={() => {
        ths.form.current.setFieldsValue({agreement: false});
        ths.setState({
          isTermsOfUseVisible: false,
        });
      }} />
  </React.Fragment>
  );
}
