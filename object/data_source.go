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
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/go-sql-driver/mysql"
)

const dataSourceFieldChars = "?#/()'\"\\=;& \t\r\n"

func checkDataSourceFields(fields map[string]string) error {
	for name, value := range fields {
		if strings.ContainsAny(value, dataSourceFieldChars) {
			return fmt.Errorf("the %s: %s of the database connection contains characters that are not allowed", name, value)
		}
	}
	return nil
}

// isTrustedDbHost reports whether the deployment operator has listed the host in the
// trustedDbHosts config, as "host" (any port) or "host:port".
func isTrustedDbHost(host string, port int) bool {
	host = strings.Trim(host, "[]")
	if host == "" {
		return false
	}

	for _, entry := range strings.Split(conf.GetConfigString("trustedDbHosts"), ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		entryHost, entryPort := entry, port
		if h, p, err := net.SplitHostPort(entry); err == nil {
			entryPort, err = strconv.Atoi(p)
			if err != nil {
				continue
			}
			entryHost = h
		}

		if entryPort == port && strings.EqualFold(strings.Trim(entryHost, "[]"), host) {
			return true
		}
	}
	return false
}

func getOwnDbAddress() (host string, port int, ok bool) {
	dataSourceName := conf.GetConfigDataSourceName()
	switch conf.GetConfigString("driverName") {
	case "mysql":
		cfg, err := mysql.ParseDSN(dataSourceName)
		if err != nil {
			return "", 0, false
		}
		if cfg.Net == "unix" {
			return "localhost", 3306, true
		}
		return splitDbHostPort(cfg.Addr, 3306)
	case "postgres":
		host = strings.Trim(util.GetValueFromDataSourceName("host", dataSourceName), `'"`)
		if host == "" || strings.HasPrefix(host, "/") {
			host = "localhost"
		}
		port = 5432
		if value := strings.Trim(util.GetValueFromDataSourceName("port", dataSourceName), `'"`); value != "" {
			port, _ = strconv.Atoi(value)
		}
		return host, port, true
	case "mssql":
		urlObj, err := url.Parse(dataSourceName)
		if err != nil || urlObj.Hostname() == "" {
			return "", 0, false
		}
		return splitDbHostPort(urlObj.Host, 1433)
	}
	return "", 0, false
}

func splitDbHostPort(address string, defaultPort int) (string, int, bool) {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return address, defaultPort, address != ""
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, false
	}
	return host, port, true
}

func isSameDbServer(host1 string, port1 int, host2 string, port2 int) bool {
	if port1 != port2 {
		return false
	}
	if strings.EqualFold(host1, host2) {
		return true
	}

	ips1 := lookupDbHostIps(host1)
	ips2 := lookupDbHostIps(host2)
	localIps := getLocalIps()
	isLocal := func(ips []net.IP) bool {
		for _, ip := range ips {
			if ip.IsLoopback() || ip.IsUnspecified() || localIps[ip.String()] {
				return true
			}
		}
		return false
	}
	if isLocal(ips1) && isLocal(ips2) {
		return true
	}

	for _, ip1 := range ips1 {
		for _, ip2 := range ips2 {
			if ip1.Equal(ip2) {
				return true
			}
		}
	}
	return false
}

func lookupDbHostIps(host string) []net.IP {
	host = strings.Trim(host, "[]")
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil
	}
	return ips
}

func getLocalIps() map[string]bool {
	res := map[string]bool{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return res
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			res[ipNet.IP.String()] = true
		}
	}
	return res
}

func quotePostgresDataSourceValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return fmt.Sprintf("'%s'", value)
}

func getMssqlDataSourceName(user string, password string, host string, port int, database string) string {
	dataSourceUrl := url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, strconv.Itoa(port)),
		RawQuery: url.Values{"database": {database}}.Encode(),
	}
	return dataSourceUrl.String()
}
