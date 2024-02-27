// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

package object

import (
	"context"
	"database/sql/driver"
	"fmt"
	"net"
	"time"

	mssql "github.com/denisenkom/go-mssqldb"

	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
)

type ViaSSHDialer struct {
	Client       *ssh.Client
	Context      *context.Context
	DatabaseType string
}

func (v *ViaSSHDialer) MysqlDial(ctx context.Context, addr string) (net.Conn, error) {
	return v.Client.Dial("tcp", addr)
}

func (v *ViaSSHDialer) Open(s string) (_ driver.Conn, err error) {
	if v.DatabaseType == "mssql" {
		c, err := mssql.NewConnector(s)
		if err != nil {
			return nil, err
		}
		c.Dialer = v
		return c.Connect(context.Background())
	} else if v.DatabaseType == "postgres" {
		return pq.DialOpen(v, s)
	}
	return nil, nil
}

func (v *ViaSSHDialer) Dial(network, address string) (net.Conn, error) {
	return v.Client.Dial(network, address)
}

func (v *ViaSSHDialer) DialContext(ctx context.Context, network string, addr string) (net.Conn, error) {
	return v.Client.DialContext(ctx, network, addr)
}

func (v *ViaSSHDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return v.Client.Dial(network, address)
}

func DialWithPassword(SshUser string, SshPassword string, SshHost string, SshPort int) (*ssh.Client, error) {
	address := fmt.Sprintf("%s:%d", SshHost, SshPort)
	config := &ssh.ClientConfig{
		User: SshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(SshPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return ssh.Dial("tcp", address, config)
}

func DialWithCert(SshUser string, CertId string, SshHost string, SshPort int) (*ssh.Client, error) {
	address := fmt.Sprintf("%s:%d", SshHost, SshPort)
	config := &ssh.ClientConfig{
		User:            SshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	cert, err := GetCert(CertId)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey([]byte(cert.PrivateKey))
	if err != nil {
		return nil, err
	}
	config.Auth = []ssh.AuthMethod{
		ssh.PublicKeys(signer),
	}
	return ssh.Dial("tcp", address, config)
}

func DialWithPrivateKey(SshUser string, PrivateKey []byte, SshHost string, SshPort int) (*ssh.Client, error) {
	address := fmt.Sprintf("%s:%d", SshHost, SshPort)
	config := &ssh.ClientConfig{
		User:            SshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	signer, err := ssh.ParsePrivateKey(PrivateKey)
	if err != nil {
		return nil, err
	}
	config.Auth = []ssh.AuthMethod{
		ssh.PublicKeys(signer),
	}
	return ssh.Dial("tcp", address, config)
}
