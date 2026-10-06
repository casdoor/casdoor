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
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"net"
	"strings"
	"time"

	mssql "github.com/microsoft/go-mssqldb"

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

func getHostKeyCallback(hostKey string, observedHostKey *string) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		*observedHostKey = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
		if strings.TrimSpace(hostKey) == "" || isHostKeyMatched(hostKey, key) {
			return nil
		}
		return fmt.Errorf("the SSH host key of %s does not match the saved one, got: %s (%s); clear \"SSH host key\" in the syncer only if the server's key was changed on purpose", hostname, *observedHostKey, ssh.FingerprintSHA256(key))
	}
}

func isHostKeyMatched(hostKey string, key ssh.PublicKey) bool {
	hostKey = strings.TrimSpace(hostKey)
	if strings.HasPrefix(hostKey, "SHA256:") {
		return hostKey == ssh.FingerprintSHA256(key)
	}

	savedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(hostKey))
	if err != nil {
		_, _, savedKey, _, _, err = ssh.ParseKnownHosts([]byte(hostKey))
		if err != nil {
			return false
		}
	}
	return bytes.Equal(savedKey.Marshal(), key.Marshal())
}

func dialSsh(sshUser string, authMethod ssh.AuthMethod, sshHost string, sshPort int, hostKey string) (*ssh.Client, string, error) {
	observedHostKey := ""
	config := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: getHostKeyCallback(hostKey, &observedHostKey),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshHost, sshPort), config)
	if err != nil {
		return nil, "", err
	}
	return client, observedHostKey, nil
}

func DialWithPassword(SshUser string, SshPassword string, SshHost string, SshPort int, HostKey string) (*ssh.Client, string, error) {
	return dialSsh(SshUser, ssh.Password(SshPassword), SshHost, SshPort, HostKey)
}

func DialWithCert(SshUser string, CertId string, SshHost string, SshPort int, HostKey string) (*ssh.Client, string, error) {
	cert, err := GetCert(CertId)
	if err != nil {
		return nil, "", err
	}
	if cert == nil {
		return nil, "", fmt.Errorf("the cert: %s is not found", CertId)
	}

	return DialWithPrivateKey(SshUser, []byte(cert.PrivateKey), SshHost, SshPort, HostKey)
}

func DialWithPrivateKey(SshUser string, PrivateKey []byte, SshHost string, SshPort int, HostKey string) (*ssh.Client, string, error) {
	signer, err := ssh.ParsePrivateKey(PrivateKey)
	if err != nil {
		return nil, "", err
	}

	return dialSsh(SshUser, ssh.PublicKeys(signer), SshHost, SshPort, HostKey)
}
