// Copyright 2020 The casbin Authors. All Rights Reserved.
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

import React, { useEffect, useState } from 'react';
import { Button, Card, Col, Input, Row } from 'antd';
import * as userBackend from '../backend/user-backend';
import { useNavigate, useParams } from 'react-router-dom';
import { showMessage } from '../setting';

function UserEditPage() {
  const [userName, setUserName] = useState(useParams().userName);
  const [user, setUser] = useState<userBackend.User | undefined>(undefined);
  const navigate = useNavigate();

  useEffect(() => {
    userBackend.getUser('admin', userName).then((res) => {
      setUser(res.data);
    });
  }, []);

  function submitUserEdit() {
    if (user == undefined) {
      return;
    }
    userBackend
      .updateUser(user.owner, userName, user)
      .then((res) => {
        if (res.data) {
          showMessage('success', `Successfully saved`);
          setUserName(user.name);
          navigate(`/users/${user.name}`);
        } else {
          showMessage('error', `failed to save: server side failure`);
          updateUserField('name', userName);
        }
      })
      .catch((error) => {
        showMessage('error', `failed to save: ${error}`);
      });
  }

  function updateUserField(key: keyof userBackend.User, value: string) {
    if (user) {
      let newUser = {
        ...user,
      };
      newUser[key] = value;
      setUser(newUser);
    }
  }

  function renderUser() {
    if (!user) {
      return;
    }

    return (
      <Card
        size="small"
        title={
          <div>
            Edit User&nbsp;&nbsp;&nbsp;&nbsp;
            <Button type="primary" onClick={submitUserEdit}>
              Save
            </Button>
          </div>
        }
        style={{ marginLeft: '5px' }}
        type="inner"
      >
        <Row style={{ marginTop: '10px' }}>
          <Col style={{ marginTop: '5px' }} span={2}>
            Name:
          </Col>
          <Col span={22}>
            <Input
              value={user.name}
              onChange={(e) => {
                console.warn(e.target.value);
                updateUserField('name', e.target.value);
              }}
            />
          </Col>
        </Row>
        <Row style={{ marginTop: '20px' }}>
          <Col style={{ marginTop: '5px' }} span={2}>
            Password Type:
          </Col>
          <Col span={22}>
            <Input
              value={user.passwordType}
              onChange={(e) => {
                updateUserField('passwordType', e.target.value);
              }}
            />
          </Col>
        </Row>
        <Row style={{ marginTop: '20px' }}>
          <Col style={{ marginTop: '5px' }} span={2}>
            Password:
          </Col>
          <Col span={22}>
            <Input
              value={user.password}
              onChange={(e) => {
                updateUserField('password', e.target.value);
              }}
            />
          </Col>
        </Row>
        <Row style={{ marginTop: '20px' }}>
          <Col style={{ marginTop: '5px' }} span={2}>
            Display Name:
          </Col>
          <Col span={22}>
            <Input
              value={user.displayName}
              onChange={(e) => {
                updateUserField('displayName', e.target.value);
              }}
            />
          </Col>
        </Row>
        <Row style={{ marginTop: '20px' }}>
          <Col style={{ marginTop: '5px' }} span={2}>
            Email:
          </Col>
          <Col span={22}>
            <Input
              value={user.email}
              onChange={(e) => {
                updateUserField('email', e.target.value);
              }}
            />
          </Col>
        </Row>
        <Row style={{ marginTop: '20px' }}>
          <Col style={{ marginTop: '5px' }} span={2}>
            Phone:
          </Col>
          <Col span={22}>
            <Input
              value={user.phone}
              onChange={(e) => {
                updateUserField('phone', e.target.value);
              }}
            />
          </Col>
        </Row>
      </Card>
    );
  }

  return (
    <div>
      <Row style={{ width: '100%' }}>
        <Col span={1} />
        <Col span={22}>{renderUser()}</Col>
        <Col span={1} />
      </Row>
      <Row style={{ margin: 10 }}>
        <Col span={2} />
        <Col span={18}>
          <Button type="primary" size="large" onClick={submitUserEdit}>
            Save
          </Button>
        </Col>
      </Row>
    </div>
  );
}

export default UserEditPage;
