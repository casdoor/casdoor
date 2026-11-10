// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/util"
	"github.com/go-ldap/ldap/v3"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/xorm-io/core"
	"golang.org/x/crypto/ssh"
)

const (
	PamRotationSucceeded = "Succeeded"
	PamRotationFailed    = "Failed"
)

const pamRotationTimeout = 30 * time.Second

// a failed scheduled rotation is tried again after this long
const pamRotationRetry = time.Hour

const (
	pamLower   = "abcdefghijkmnopqrstuvwxyz"
	pamUpper   = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	pamDigits  = "23456789"
	pamSymbols = "!#%+-=?@^_~.,"
)

// pamLoginError is a failure to sign in to the target system, before anything was changed
type pamLoginError struct {
	err error
}

func (e *pamLoginError) Error() string {
	return e.err.Error()
}

// GeneratePamPassword makes a random password with at least one character of each class
func GeneratePamPassword(length int, charset string) (string, error) {
	if length <= 0 {
		length = 24
	}
	classes := []string{pamLower, pamUpper, pamDigits}
	if charset != "Alphanumeric" {
		classes = append(classes, pamSymbols)
	}
	all := strings.Join(classes, "")

	pick := func(set string) (byte, error) {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
		if err != nil {
			return 0, err
		}
		return set[n.Int64()], nil
	}

	res := make([]byte, length)
	for i := range res {
		set := all
		if i < len(classes) {
			set = classes[i]
		}
		c, err := pick(set)
		if err != nil {
			return "", err
		}
		res[i] = c
	}
	for i := len(res) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := n.Int64()
		res[i], res[j] = res[j], res[i]
	}
	return string(res), nil
}

type pamCredentials struct {
	// the current password of the account, empty when the admin credential is used
	old       string
	admin     string
	newSecret string
	// called after signing in, right before the password is changed on the target system
	beforeChange func() error
}

func pamAddress(account *PrivilegedAccount, defaultPort int) string {
	port := account.Port
	if port == 0 {
		port = defaultPort
	}
	return net.JoinHostPort(account.Host, strconv.Itoa(port))
}

var pamRotationLocks sync.Map

// RotatePrivilegedAccount sets a new random password of the account on the target system and keeps it in the vault,
// the result is stored on the account and recorded
func RotatePrivilegedAccount(account *PrivilegedAccount, trigger string) error {
	lock, _ := pamRotationLocks.LoadOrStore(account.GetId(), &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()

	fresh, err := getPrivilegedAccount(account.Owner, account.Name)
	if err != nil {
		return err
	}
	if fresh == nil {
		return fmt.Errorf("the privileged account: %s does not exist", account.GetId())
	}
	account = fresh

	err = rotatePrivilegedAccount(account)
	now := formatPamTime(time.Now())
	account.RotationTime = now
	if err == nil {
		account.LastRotatedTime = now
		account.RotationState = PamRotationSucceeded
		account.RotationMessage = ""
		account.NeedsRotation = false
	} else {
		account.RotationState = PamRotationFailed
		account.RotationMessage = err.Error()
		if len(account.RotationMessage) > 1000 {
			account.RotationMessage = account.RotationMessage[:1000]
		}
	}
	_, updateErr := ormer.Engine.ID(core.PK{account.Owner, account.Name}).
		Cols("rotation_time", "last_rotated_time", "rotation_state", "rotation_message", "needs_rotation").Update(account)
	if updateErr != nil {
		logs.Error("failed to save the rotation result of the privileged account: %s: %s", account.GetId(), updateErr.Error())
	}

	failure := ""
	if err != nil {
		failure = err.Error()
	}
	addPamRecord(account.Owner, "", "rotate-privileged-account", map[string]string{"account": account.GetId(), "type": account.Type, "trigger": trigger}, failure)
	return err
}

func rotatePrivilegedAccount(account *PrivilegedAccount) error {
	if err := edition.Require(edition.Enterprise, edition.FeaturePam); err != nil {
		return err
	}
	if !account.canRotate() {
		return fmt.Errorf("passwords of the type: %s cannot be rotated by Casdoor", account.Type)
	}
	if account.Host == "" {
		return errors.New("the host is empty")
	}
	if !IsFieldEncryptionKeyConfigured() {
		return errors.New("the configuration fieldEncryptionKey is not set")
	}

	newSecret, err := GeneratePamPassword(account.PasswordLength, account.PasswordCharset)
	if err != nil {
		return err
	}
	adminSecret, err := account.getAdminSecret()
	if err != nil {
		return err
	}
	secret, err := account.getSecret()
	if err != nil {
		return err
	}
	pendingSecret, err := account.getPendingSecret()
	if err != nil {
		return err
	}

	// a rotation that failed after the change may have set the pending password, so it is tried as well
	candidates := []string{secret}
	if pendingSecret != "" && pendingSecret != secret {
		candidates = append(candidates, pendingSecret)
	}
	useAdmin := account.AdminUsername != ""
	if useAdmin {
		candidates = []string{""}
	}

	changing := false
	for i, old := range candidates {
		credentials := &pamCredentials{old: old, admin: adminSecret, newSecret: newSecret}
		credentials.beforeChange = func() error {
			changing = true
			var encryptErr error
			if !useAdmin && old != secret {
				// the pending password turned out to be the current one
				account.Secret, encryptErr = encryptField(account.Owner, "privilegedAccount.secret", old)
				if encryptErr != nil {
					return encryptErr
				}
			}
			account.PendingSecret, encryptErr = encryptField(account.Owner, "privilegedAccount.secret", newSecret)
			if encryptErr != nil {
				return encryptErr
			}
			_, updateErr := ormer.Engine.ID(core.PK{account.Owner, account.Name}).Cols("secret", "pending_secret").Update(account)
			return updateErr
		}

		err = runPamConnector(account, credentials)
		var loginErr *pamLoginError
		if err != nil && errors.As(err, &loginErr) && !changing && i < len(candidates)-1 {
			continue
		}
		break
	}

	if err != nil {
		if changing {
			return fmt.Errorf("%s (the new password may or may not have been set, the attempted one is kept and tried next time)", err.Error())
		}
		return err
	}

	account.Secret, err = encryptField(account.Owner, "privilegedAccount.secret", newSecret)
	if err != nil {
		return err
	}
	account.PendingSecret = ""
	_, err = ormer.Engine.ID(core.PK{account.Owner, account.Name}).Cols("secret", "pending_secret", "host_key").Update(account)
	return err
}

func runPamConnector(account *PrivilegedAccount, credentials *pamCredentials) error {
	switch account.Type {
	case PamTypeLinux:
		return rotateSshPassword(account, credentials)
	case PamTypeMysql:
		return rotateMysqlPassword(account, credentials)
	case PamTypePostgresql:
		return rotatePostgresqlPassword(account, credentials)
	case PamTypeLdap, PamTypeAd:
		return rotateLdapPassword(account, credentials)
	}
	return fmt.Errorf("passwords of the type: %s cannot be rotated by Casdoor", account.Type)
}

func dialPamSsh(account *PrivilegedAccount, username string, password string) (*ssh.Client, error) {
	answer := func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		answers := make([]string, len(questions))
		for i := range answers {
			answers[i] = password
		}
		return answers, nil
	}
	config := &ssh.ClientConfig{
		User:    username,
		Auth:    []ssh.AuthMethod{ssh.Password(password), ssh.KeyboardInteractive(answer)},
		Timeout: pamRotationTimeout,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			fingerprint := ssh.FingerprintSHA256(key)
			if account.HostKey == "" {
				account.HostKey = fingerprint
				return nil
			}
			if account.HostKey != fingerprint {
				return fmt.Errorf("the host key of %s is %s, not the pinned %s", hostname, fingerprint, account.HostKey)
			}
			return nil
		},
	}
	client, err := ssh.Dial("tcp", pamAddress(account, 22), config)
	if err != nil {
		if strings.Contains(err.Error(), "unable to authenticate") {
			return nil, &pamLoginError{err: err}
		}
		return nil, err
	}
	return client, nil
}

func runPamSshCommand(client *ssh.Client, command string, stdin string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var output bytes.Buffer
	session.Stdin = strings.NewReader(stdin)
	session.Stdout = &output
	session.Stderr = &output
	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()
	select {
	case err = <-done:
	case <-time.After(pamRotationTimeout):
		return fmt.Errorf("%s timed out", command)
	}
	if err != nil {
		return fmt.Errorf("%s failed: %s: %s", command, err.Error(), strings.TrimSpace(output.String()))
	}
	return nil
}

// rotateSshPassword changes the password of a Linux or Unix account: by the account itself with passwd, or with chpasswd
// by the admin credential (with sudo unless it's root)
func rotateSshPassword(account *PrivilegedAccount, credentials *pamCredentials) error {
	username, password := account.Username, credentials.old
	if account.AdminUsername != "" {
		username, password = account.AdminUsername, credentials.admin
	}
	client, err := dialPamSsh(account, username, password)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = credentials.beforeChange(); err != nil {
		return err
	}
	switch {
	case account.AdminUsername == "root" || (account.AdminUsername == "" && account.Username == "root"):
		err = runPamSshCommand(client, "chpasswd", fmt.Sprintf("%s:%s\n", account.Username, credentials.newSecret))
	case account.AdminUsername != "":
		err = runPamSshCommand(client, "sudo -S -p '' chpasswd", fmt.Sprintf("%s\n%s:%s\n", credentials.admin, account.Username, credentials.newSecret))
	default:
		err = runPamSshCommand(client, "passwd", fmt.Sprintf("%s\n%s\n%s\n", credentials.old, credentials.newSecret, credentials.newSecret))
	}
	if err != nil {
		return err
	}

	verifier, err := dialPamSsh(account, account.Username, credentials.newSecret)
	if err != nil {
		if account.AdminUsername != "" {
			// the account itself may not be allowed to sign in over SSH
			return nil
		}
		return fmt.Errorf("the password was changed but signing in with it failed: %s", err.Error())
	}
	return verifier.Close()
}

func splitMysqlUser(username string) (string, string) {
	if index := strings.LastIndex(username, "@"); index > 0 {
		return username[:index], username[index+1:]
	}
	return username, "%"
}

func openPamMysql(account *PrivilegedAccount, username string, password string) (*sql.DB, error) {
	config := mysql.NewConfig()
	config.User = username
	config.Passwd = password
	config.Net = "tcp"
	config.Addr = pamAddress(account, 3306)
	config.InterpolateParams = true
	config.Timeout = pamRotationTimeout
	config.ReadTimeout = pamRotationTimeout
	config.WriteTimeout = pamRotationTimeout
	if account.SkipTlsVerify {
		config.TLSConfig = "preferred"
	} else {
		config.TLS = &tls.Config{ServerName: account.Host, MinVersion: tls.VersionTLS12}
	}
	connector, err := mysql.NewConnector(config)
	if err != nil {
		return nil, err
	}
	db := sql.OpenDB(connector)
	ctx, cancel := context.WithTimeout(context.Background(), pamRotationTimeout)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1045 {
			return nil, &pamLoginError{err: err}
		}
		return nil, err
	}
	return db, nil
}

// rotateMysqlPassword changes the password with ALTER USER, by the account itself or by the admin credential, the
// username may be given as "name@host" for the admin to change
func rotateMysqlPassword(account *PrivilegedAccount, credentials *pamCredentials) error {
	self := account.AdminUsername == ""
	username, password := account.AdminUsername, credentials.admin
	if self {
		username, _ = splitMysqlUser(account.Username)
		password = credentials.old
	}
	db, err := openPamMysql(account, username, password)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = credentials.beforeChange(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), pamRotationTimeout)
	defer cancel()
	if self {
		_, err = db.ExecContext(ctx, "ALTER USER CURRENT_USER() IDENTIFIED BY ? REPLACE ?", credentials.newSecret, credentials.old)
	} else {
		name, host := splitMysqlUser(account.Username)
		_, err = db.ExecContext(ctx, "ALTER USER ?@? IDENTIFIED BY ?", name, host, credentials.newSecret)
	}
	if err != nil {
		return err
	}

	if self {
		verifier, err := openPamMysql(account, username, credentials.newSecret)
		if err != nil {
			return fmt.Errorf("the password was changed but signing in with it failed: %s", err.Error())
		}
		return verifier.Close()
	}
	return nil
}

func quotePqValue(value string) string {
	return "'" + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), "'", `\'`) + "'"
}

func openPamPostgresql(account *PrivilegedAccount, username string, password string) (*sql.DB, error) {
	host, port, _ := net.SplitHostPort(pamAddress(account, 5432))
	database := account.Database
	if database == "" {
		database = "postgres"
	}
	sslMode := "verify-full"
	if account.SkipTlsVerify {
		sslMode = "prefer"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d", quotePqValue(host), port,
		quotePqValue(username), quotePqValue(password), quotePqValue(database), sslMode, int(pamRotationTimeout.Seconds()))
	connector, err := pq.NewConnector(dsn)
	if err != nil {
		return nil, err
	}
	db := sql.OpenDB(connector)
	ctx, cancel := context.WithTimeout(context.Background(), pamRotationTimeout)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "28P01" {
			return nil, &pamLoginError{err: err}
		}
		return nil, err
	}
	return db, nil
}

// rotatePostgresqlPassword changes the password with ALTER ROLE, by the account itself or by the admin credential
func rotatePostgresqlPassword(account *PrivilegedAccount, credentials *pamCredentials) error {
	self := account.AdminUsername == ""
	username, password := account.AdminUsername, credentials.admin
	if self {
		username, password = account.Username, credentials.old
	}
	db, err := openPamPostgresql(account, username, password)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = credentials.beforeChange(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), pamRotationTimeout)
	defer cancel()
	role := "CURRENT_USER"
	if !self {
		role = pq.QuoteIdentifier(account.Username)
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER ROLE %s WITH PASSWORD %s", role, pq.QuoteLiteral(credentials.newSecret)))
	if err != nil {
		return err
	}

	if self {
		verifier, err := openPamPostgresql(account, username, credentials.newSecret)
		if err != nil {
			return fmt.Errorf("the password was changed but signing in with it failed: %s", err.Error())
		}
		return verifier.Close()
	}
	return nil
}

func encodeAdPassword(password string) string {
	codes := utf16.Encode([]rune("\"" + password + "\""))
	buf := make([]byte, len(codes)*2)
	for i, code := range codes {
		binary.LittleEndian.PutUint16(buf[i*2:], code)
	}
	return string(buf)
}

func dialPamLdap(account *PrivilegedAccount, bindDn string, password string) (*ldap.Conn, error) {
	address := pamAddress(account, 636)
	tlsConfig := &tls.Config{ServerName: account.Host, InsecureSkipVerify: account.SkipTlsVerify, MinVersion: tls.VersionTLS12} // #nosec G402 -- set by the admin for an internal CA
	conn, err := ldap.DialURL("ldaps://"+address, ldap.DialWithTLSConfig(tlsConfig), ldap.DialWithDialer(&net.Dialer{Timeout: pamRotationTimeout}))
	if err != nil {
		return nil, err
	}
	conn.SetTimeout(pamRotationTimeout)
	if err = conn.Bind(bindDn, password); err != nil {
		conn.Close()
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return nil, &pamLoginError{err: err}
		}
		return nil, err
	}
	return conn, nil
}

// rotateLdapPassword changes the password over LDAPS: the unicodePwd attribute for Active Directory (a reset by the
// admin, or a delete and add of the old and new value by the account itself), the password modify extended operation
// otherwise. The username is the DN of the account.
func rotateLdapPassword(account *PrivilegedAccount, credentials *pamCredentials) error {
	self := account.AdminUsername == ""
	bindDn, password := account.AdminUsername, credentials.admin
	if self {
		bindDn, password = account.Username, credentials.old
	}
	conn, err := dialPamLdap(account, bindDn, password)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err = credentials.beforeChange(); err != nil {
		return err
	}
	if account.Type == PamTypeAd {
		request := ldap.NewModifyRequest(account.Username, nil)
		if self {
			request.Delete("unicodePwd", []string{encodeAdPassword(credentials.old)})
			request.Add("unicodePwd", []string{encodeAdPassword(credentials.newSecret)})
		} else {
			request.Replace("unicodePwd", []string{encodeAdPassword(credentials.newSecret)})
		}
		err = conn.Modify(request)
	} else {
		old := ""
		if self {
			old = credentials.old
		}
		_, err = conn.PasswordModify(ldap.NewPasswordModifyRequest(account.Username, old, credentials.newSecret))
	}
	if err != nil {
		return err
	}

	verifier, err := dialPamLdap(account, account.Username, credentials.newSecret)
	if err != nil {
		return fmt.Errorf("the password was changed but signing in with it failed: %s", err.Error())
	}
	verifier.Close()
	return nil
}

// isPamRotationDue tells whether the account's password is to be rotated now: after a check-in, or on its schedule,
// and a failed attempt waits a while before the next one
func isPamRotationDue(account *PrivilegedAccount, now time.Time) bool {
	if !account.IsEnabled || !account.canRotate() || account.Host == "" {
		return false
	}
	if attempt, ok := parsePamTime(account.RotationTime); ok && account.RotationState == PamRotationFailed && now.Sub(attempt) < pamRotationRetry {
		return false
	}
	if account.NeedsRotation {
		return true
	}
	if account.RotateIntervalDays <= 0 {
		return false
	}
	last, ok := parsePamTime(account.LastRotatedTime)
	if !ok {
		last, ok = parsePamTime(account.CreatedTime)
	}
	return !ok || now.Sub(last) >= time.Duration(account.RotateIntervalDays)*24*time.Hour
}

// RotateDuePrivilegedAccounts rotates the passwords that are due, but not while someone has the account checked out
func RotateDuePrivilegedAccounts() error {
	accounts := []*PrivilegedAccount{}
	err := ormer.Engine.Where("needs_rotation = ? or rotate_interval_days > ?", true, 0).Find(&accounts)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, account := range accounts {
		if !isPamRotationDue(account, now) {
			continue
		}
		count, err := ormer.Engine.Where("owner = ? and account = ?", account.Owner, account.Name).In("state", PamStateApproved, PamStateActive).Count(&PrivilegedRequest{})
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		trigger := "schedule"
		if account.NeedsRotation {
			trigger = "check-in"
		}
		if err = RotatePrivilegedAccount(account, trigger); err != nil {
			logs.Warn("failed to rotate the password of the privileged account: %s: %s", account.GetId(), err.Error())
		}
	}
	return nil
}

var pamJobRunning sync.Mutex

func runPamJobs() {
	if !pamJobRunning.TryLock() {
		return
	}
	defer pamJobRunning.Unlock()
	defer func() {
		if r := recover(); r != nil {
			logs.Error("the privileged access jobs panicked: %v", r)
		}
	}()

	if err := ExpirePrivilegedRequests(); err != nil {
		logs.Error("failed to expire the privileged requests: %s", err.Error())
	}
	if err := RotateDuePrivilegedAccounts(); err != nil {
		logs.Error("failed to rotate the privileged accounts: %s", err.Error())
	}
}

// InitPrivilegedAccessJobs expires the privileged access requests and rotates the due passwords every minute
func InitPrivilegedAccessJobs() {
	util.SafeGoroutine(func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			runPamJobs()
		}
	})
}
