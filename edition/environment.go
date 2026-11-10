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

package edition

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// Environment is the machine and the setup a licensed Casdoor runs on, reported with the license so that the vendor
// knows what it supports
type Environment struct {
	Os             string   `json:"os"`
	Platform       string   `json:"platform"`
	Kernel         string   `json:"kernel"`
	Arch           string   `json:"arch"`
	Virtualization string   `json:"virtualization"`
	CpuModel       string   `json:"cpuModel"`
	CpuCores       int      `json:"cpuCores"`
	MemoryTotal    uint64   `json:"memoryTotal"`
	DiskTotal      uint64   `json:"diskTotal"`
	DiskFree       uint64   `json:"diskFree"`
	Deployment     string   `json:"deployment"`
	Supervisor     string   `json:"supervisor"`
	DbType         string   `json:"dbType"`
	DbVersion      string   `json:"dbVersion"`
	GoVersion      string   `json:"goVersion"`
	Origin         string   `json:"origin"`
	PrivateIps     []string `json:"privateIps"`
	StartTime      string   `json:"startTime"`
	TimeZone       string   `json:"timeZone"`
}

var startTime = time.Now()

func getSupervisor() string {
	switch {
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "":
		return "kubernetes"
	case os.Getenv("INVOCATION_ID") != "" || os.Getenv("NOTIFY_SOCKET") != "":
		return "systemd"
	case os.Getenv("SUPERVISOR_ENABLED") != "":
		return "supervisord"
	case os.Getenv("PM2_HOME") != "" || os.Getenv("pm_id") != "":
		return "pm2"
	}
	return ""
}

func getPrivateIps() []string {
	res := []string{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return res
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.IsLinkLocalUnicast() {
			continue
		}
		res = append(res, ipNet.IP.String())
	}
	sort.Strings(res)
	if len(res) > 8 {
		res = res[:8]
	}
	return res
}

func getDataDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

// collectEnvironment reads the environment, the parts that need the database or the upgrade package come from the reporter
func collectEnvironment() *Environment {
	env := &Environment{
		Os:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		CpuCores:   runtime.NumCPU(),
		Supervisor: getSupervisor(),
		DbType:     conf.GetConfigString("driverName"),
		GoVersion:  runtime.Version(),
		Origin:     conf.GetConfigString("origin"),
		PrivateIps: getPrivateIps(),
		StartTime:  startTime.Format(time.RFC3339),
	}
	env.TimeZone, _ = time.Now().Zone()

	if info, err := host.Info(); err == nil {
		env.Platform = strings.TrimSpace(info.Platform + " " + info.PlatformVersion)
		env.Kernel = info.KernelVersion
		if info.VirtualizationRole == "guest" {
			env.Virtualization = info.VirtualizationSystem
		}
	}
	if infos, err := cpu.Info(); err == nil && len(infos) > 0 {
		env.CpuModel = strings.TrimSpace(infos[0].ModelName)
		if env.CpuModel == "" {
			env.CpuModel = strings.TrimSpace(infos[0].VendorID + " " + infos[0].Family + " " + infos[0].Model)
		}
	}
	if vm, err := mem.VirtualMemory(); err == nil {
		env.MemoryTotal = vm.Total
	}
	if usage, err := disk.Usage(getDataDir()); err == nil {
		env.DiskTotal = usage.Total
		env.DiskFree = usage.Free
	}

	r := getReporter()
	if r != nil && r.Deployment != nil {
		env.Deployment = r.Deployment()
	}
	if r != nil && r.DbVersion != nil {
		env.DbVersion = r.DbVersion()
	}
	return env
}
