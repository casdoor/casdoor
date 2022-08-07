// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

import React, {useState} from "react";
import Cropper from "react-cropper";
import "cropperjs/dist/cropper.css";
import * as Setting from "./Setting";
import {Button, Col, Modal, Row} from "antd";
import i18next from "i18next";
import * as ResourceBackend from "./backend/ResourceBackend";

export const CropperDiv = (props) => {
  const [image, setImage] = useState("");
  const [cropper, setCropper] = useState();
  const [visible, setVisible] = React.useState(false);
  const [confirmLoading, setConfirmLoading] = React.useState(false);
  const {title} = props;
  const {user} = props;
  const {buttonText} = props;
  let uploadButton;

  const onChange = (e) => {
    e.preventDefault();
    let files;
    if (e.dataTransfer) {
      files = e.dataTransfer.files;
    } else if (e.target) {
      files = e.target.files;
    }
    const reader = new FileReader();
    reader.onload = () => {
      setImage(reader.result);
    };
    if (!(files[0] instanceof Blob)) {
      return;
    }
    reader.readAsDataURL(files[0]);
  };

  const uploadAvatar = () => {
    cropper.getCroppedCanvas().toBlob(blob => {
      if (blob === null) {
        Setting.showMessage("error", "You must select a picture first!");
        return false;
      }
      // Setting.showMessage("success", "uploading...");
      const extension = image.substring(image.indexOf("/") + 1, image.indexOf(";base64"));
      const fullFilePath = `avatar/${user.owner}/${user.name}.${extension}`;
      ResourceBackend.uploadResource(user.owner, user.name, "avatar", "CropperDiv", fullFilePath, blob)
        .then((res) => {
          if (res.status === "ok") {
            window.location.href = "/account";
          } else {
            Setting.showMessage("error", res.msg);
          }
        });
      return true;
    });
  };

  const showModal = () => {
    setVisible(true);
  };

  const handleOk = () => {
    setConfirmLoading(true);
    if (!uploadAvatar()) {
      setConfirmLoading(false);
    }
  };

  const handleCancel = () => {
    console.log("Clicked cancel button");
    setVisible(false);
  };

  const selectFile = () => {
    uploadButton.click();
  };

  return (
    <div>
      <Button type="default" onClick={showModal}>
        {buttonText}
      </Button>
      <Modal
        maskClosable={false}
        title={title}
        visible={visible}
        okText={i18next.t("user:Upload a photo")}
        confirmLoading={confirmLoading}
        onCancel={handleCancel}
        width={600}
        footer={
          [<Button key="placeholderKey" block type="primary" onClick={handleOk}>{i18next.t("user:Set new profile picture")}</Button>]
        }
      >
        <Col style={{margin: "0px auto 40px auto", width: 1000, height: 300}}>
          <Row style={{width: "100%", marginBottom: "20px"}}>
            <input style={{display: "none"}} ref={input => uploadButton = input} type="file" accept="image/*" onChange={onChange} />
            <Button block onClick={selectFile}>{i18next.t("user:Select a photo...")}</Button>
          </Row>
          <Cropper
            style={{height: "100%"}}
            initialAspectRatio={1}
            preview=".img-preview"
            src={image}
            viewMode={1}
            guides={true}
            minCropBoxHeight={10}
            minCropBoxWidth={10}
            background={false}
            responsive={true}
            autoCropArea={1}
            checkOrientation={false}
            onInitialized={(instance) => {
              setCropper(instance);
            }}
          />
        </Col>
      </Modal>
    </div>
  );
};

export default CropperDiv;
