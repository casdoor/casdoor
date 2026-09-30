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

package audit

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	syslogQueueSize     = 10000
	syslogDialTimeout   = 5 * time.Second
	syslogWriteTimeout  = 5 * time.Second
	syslogRetryInterval = 10 * time.Second
	syslogUdpMaxLength  = 60000
	syslogMsgIdMaxLen   = 32
)

var syslogNetworks = map[string]string{
	"UDP": "udp",
	"TCP": "tcp",
	"TLS": "tls",
}

var syslogFacilities = map[string]int{
	"kern": 0, "user": 1, "mail": 2, "daemon": 3, "auth": 4, "syslog": 5, "lpr": 6, "news": 7,
	"uucp": 8, "cron": 9, "authpriv": 10, "ftp": 11, "ntp": 12, "security": 13, "console": 14, "solaris-cron": 15,
	"local0": 16, "local1": 17, "local2": 18, "local3": 19, "local4": 20, "local5": 21, "local6": 22, "local7": 23,
}

var syslogFormats = map[string]bool{
	"RFC 5424": true,
	"RFC 3164": true,
}

type syslogSettings struct {
	network  string
	address  string
	host     string
	facility int
	format   string
}

func newSyslogSettings(config *Config) (*syslogSettings, error) {
	host := strings.TrimSpace(config.Host)
	if host == "" {
		return nil, fmt.Errorf("the syslog server host should not be empty")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return nil, fmt.Errorf("the syslog server port: %d is invalid", config.Port)
	}

	protocol := config.Protocol
	if protocol == "" {
		protocol = "UDP"
	}
	network, ok := syslogNetworks[protocol]
	if !ok {
		return nil, fmt.Errorf("the syslog protocol: %s is not supported", config.Protocol)
	}

	facilityName := config.Facility
	if facilityName == "" {
		facilityName = "auth"
	}
	facility, ok := syslogFacilities[facilityName]
	if !ok {
		return nil, fmt.Errorf("the syslog facility: %s is not supported", config.Facility)
	}

	format := config.Format
	if format == "" {
		format = "RFC 5424"
	}
	if !syslogFormats[format] {
		return nil, fmt.Errorf("the syslog format: %s is not supported", config.Format)
	}

	return &syslogSettings{
		network:  network,
		address:  net.JoinHostPort(host, strconv.Itoa(config.Port)),
		host:     host,
		facility: facility,
		format:   format,
	}, nil
}

type SyslogProvider struct {
	settings  *syslogSettings
	hostname  string
	procId    string
	queue     chan string
	done      chan struct{}
	closeOnce sync.Once
	conn      net.Conn
}

func NewSyslogProvider(config *Config) (*SyslogProvider, error) {
	settings, err := newSyslogSettings(config)
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "-"
	}

	p := &SyslogProvider{
		settings: settings,
		hostname: hostname,
		procId:   strconv.Itoa(os.Getpid()),
		queue:    make(chan string, syslogQueueSize),
		done:     make(chan struct{}),
	}
	go p.run()
	return p, nil
}

func (p *SyslogProvider) Send(event *Event) error {
	message := p.formatMessage(event)
	select {
	case p.queue <- message:
		return nil
	default:
		return fmt.Errorf("the syslog queue for %s is full, the audit event is dropped", p.settings.address)
	}
}

func (p *SyslogProvider) Close() error {
	p.closeOnce.Do(func() {
		close(p.done)
	})
	return nil
}

func (p *SyslogProvider) run() {
	defer p.closeConn()

	for {
		select {
		case <-p.done:
			return
		case message := <-p.queue:
			for !p.write(message) {
				select {
				case <-p.done:
					return
				case <-time.After(syslogRetryInterval):
				}
			}
		}
	}
}

func (p *SyslogProvider) write(message string) bool {
	if p.conn == nil {
		conn, err := p.dial()
		if err != nil {
			fmt.Printf("SyslogProvider: failed to connect to %s: %v\n", p.settings.address, err)
			return false
		}
		p.conn = conn
	}

	_ = p.conn.SetWriteDeadline(time.Now().Add(syslogWriteTimeout))
	_, err := p.conn.Write([]byte(message))
	if err != nil {
		fmt.Printf("SyslogProvider: failed to write to %s: %v\n", p.settings.address, err)
		p.closeConn()
		return false
	}
	return true
}

func (p *SyslogProvider) dial() (net.Conn, error) {
	dialer := &net.Dialer{Timeout: syslogDialTimeout}
	if p.settings.network == "tls" {
		return tls.DialWithDialer(dialer, "tcp", p.settings.address, &tls.Config{
			ServerName: p.settings.host,
			MinVersion: tls.VersionTLS12,
		})
	}
	return dialer.Dial(p.settings.network, p.settings.address)
}

func (p *SyslogProvider) closeConn() {
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
}

func (p *SyslogProvider) formatMessage(event *Event) string {
	eventTime := event.Time
	if eventTime.IsZero() {
		eventTime = time.Now()
	}

	priority := p.settings.facility*8 + event.Severity
	message := strings.NewReplacer("\r", " ", "\n", " ").Replace(event.Message)

	var line string
	if p.settings.format == "RFC 3164" {
		line = fmt.Sprintf("<%d>%s %s casdoor[%s]: %s", priority, eventTime.Format(time.Stamp), p.hostname, p.procId, message)
	} else {
		line = fmt.Sprintf("<%d>1 %s %s casdoor %s %s - %s", priority, eventTime.Format(time.RFC3339), p.hostname, p.procId, getSyslogMsgId(event.Action), message)
	}

	// TCP and TLS streams are framed by a trailing LF, each UDP datagram carries exactly one message
	if p.settings.network == "udp" {
		if len(line) > syslogUdpMaxLength {
			line = strings.ToValidUTF8(line[:syslogUdpMaxLength], "")
		}
		return line
	}
	return line + "\n"
}

func getSyslogMsgId(action string) string {
	msgId := strings.Map(func(r rune) rune {
		if r < 33 || r > 126 {
			return '_'
		}
		return r
	}, action)
	if len(msgId) > syslogMsgIdMaxLen {
		msgId = msgId[:syslogMsgIdMaxLen]
	}
	if msgId == "" {
		return "-"
	}
	return msgId
}
