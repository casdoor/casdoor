import React, { useState } from "react";
import Cropper from "react-cropper";
import "cropperjs/dist/cropper.css";
import * as Setting from "./Setting";
import * as UserBackend from "./backend/UserBackend";
import {DownOutlined, LogoutOutlined, SettingOutlined} from '@ant-design/icons';
import {Input, Button, Row, Col} from 'antd';

export const CropperDiv: React.FC = () => {
  const [image, setImage] = useState("");
  const [cropData, setCropData] = useState("#");
  const [cropper, setCropper] = useState<any>();

  const onChange = (e: any) => {
    e.preventDefault();
    let files;
    if (e.dataTransfer) {
      files = e.dataTransfer.files;
    } else if (e.target) {
      files = e.target.files;
    }
    const reader = new FileReader();
    reader.onload = () => {
      setImage(reader.result as any);
    };
    reader.readAsDataURL(files[0]);
  };

  const getCropData = () => {
    if (typeof cropper !== "undefined") {
      if (cropper.getCroppedCanvas() === null) {
        Setting.showMessage("error", "You haven't select a picture.");
        return;
      }
      setCropData(cropper.getCroppedCanvas().toDataURL());
    }
  };

  const uploadAvatar = () => {
    let canvas = cropper.getCroppedCanvas();
    if (canvas === null) {
      Setting.showMessage("error", "You must select a picture first!");
      return;
    }
    UserBackend.uploadAvatar(canvas.toDataURL());
  }

  
    return (<Col>
        <Row>
        <Col style={{margin: "20px auto 100px auto" }}>
          <Row style={{width: "100%"}}>
            <Input type="file" onChange={onChange} style={{width: "60%"}}/>
            <Button onClick={uploadAvatar} style={{marginLeft: "30px"}}>Upload Avatar</Button>
          </Row>
          <Cropper
            style={{ height: 400, width: "60%" }}
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
        </Row>
      </Col>
    )
};

export default CropperDiv;
