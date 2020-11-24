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
import { Button, Popconfirm, Table } from 'antd';
import { getFormattedDate, showMessage } from '../setting';
import { useNavigate } from 'react-router-dom';
import moment from 'moment';
import * as userBackend from '../backend/user-backend';
import tw from 'twin.macro';

function UserTable() {
  const [users, setUsers] = useState<Array<any>>([]);
  const navigate = useNavigate();

  useEffect(() => {
    userBackend.getUsers('admin').then((res) => {
      setUsers(res.data);
    });
  }, []);

  function newUser() {
    return {
      owner: 'admin', // this.props.account.username,
      name: `user_${users.length}`,
      createdTime: moment().format(),
      password: '123456',
      passwordType: 'plain',
      displayName: `New User - ${users.length}`,
      email: 'user@example.com',
      phone: '1-12345678',
    };
  }

  function addUser() {
    const value = newUser();
    userBackend
      .addUser(value)
      .then((res) => {
        showMessage('success', `User added successfully`);
        setUsers([...users, value]);
      })
      .catch((error) => {
        showMessage('error', `User failed to add: ${error}`);
      });
  }

  function deleteUser(i: number) {
    userBackend
      .deleteUser(users[i])
      .then((res) => {
        showMessage('success', `User deleted successfully`);
        setUsers(users.filter((n) => n != users[i]));
      })
      .catch((error) => {
        showMessage('error', `User failed to delete: ${error}`);
      });
  }

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      width: '120px',
      sorter: (a: any, b: any) => a.name.localeCompare(b.name),
      render: (text: string) => {
        return <a href={`/users/${text}`}>{text}</a>;
      },
    },
    {
      title: 'Created Time',
      dataIndex: 'createdTime',
      key: 'createdTime',
      width: '160px',
      sorter: (a: any, b: any) => a.createdTime.localeCompare(b.createdTime),
      render: (text: string) => {
        return getFormattedDate(text);
      },
    },
    {
      title: 'PasswordType',
      dataIndex: 'passwordType',
      key: 'passwordType',
      width: '150px',
      sorter: (a: any, b: any) => a.passwordType.localeCompare(b.passwordType),
    },
    {
      title: 'Password',
      dataIndex: 'password',
      key: 'password',
      width: '150px',
      sorter: (a: any, b: any) => a.password.localeCompare(b.password),
    },
    {
      title: 'Display Name',
      dataIndex: 'displayName',
      key: 'displayName',
      // width: '100px',
      sorter: (a: any, b: any) => a.displayName.localeCompare(b.displayName),
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
      width: '150px',
      sorter: (a: any, b: any) => a.email.localeCompare(b.email),
    },
    {
      title: 'Phone',
      dataIndex: 'phone',
      key: 'phone',
      width: '120px',
      sorter: (a: any, b: any) => a.phone.localeCompare(b.phone),
    },
    {
      title: 'Action',
      dataIndex: '',
      key: 'op',
      width: '170px',
      render: (text: string, record: any, index: number) => {
        return (
          <div>
            <Button
              style={{
                marginTop: '10px',
                marginBottom: '10px',
                marginRight: '10px',
              }}
              type="primary"
              onClick={() => {
                navigate(`/users/${record.name}`);
              }}
            >
              Edit
            </Button>
            <Popconfirm title={`Sure to delete user: ${record.name} ?`} onConfirm={() => deleteUser(index)}>
              <Button style={{ marginBottom: '10px' }} type="dashed">
                Delete
              </Button>
            </Popconfirm>
          </div>
        );
      },
    },
  ];
  return (
    <div>
      <div css={tw`flex items-center justify-between`}>
        <h1 css={tw`m-0 text-2xl`}>Users</h1>
        <Button type="primary" size="middle" onClick={addUser}>
          Add
        </Button>
      </div>
      <div css={tw`py-4`}>
        <Table
          columns={columns}
          dataSource={users}
          rowKey="name"
          size="middle"
          pagination={{ pageSize: 100 }}
          loading={users === null || users === undefined}
        />
      </div>
    </div>
  );
}

function UserListPage() {
  return (
    <>
      <UserTable />
    </>
  );
}

export default UserListPage;
