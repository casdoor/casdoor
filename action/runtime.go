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

package action

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Abi is the calling convention between Casdoor and an action module, the "abi" of its action.json
const Abi = 1

const (
	callTimeout     = 60 * time.Second
	fetchTimeout    = 30 * time.Second
	maxResponseSize = 10 << 20
	// 256 MB
	memoryLimitPages = 4096
)

type Config struct {
	ClientId      string `json:"clientId"`
	ClientSecret  string `json:"clientSecret"`
	ClientId2     string `json:"clientId2"`
	ClientSecret2 string `json:"clientSecret2"`
	Host          string `json:"host"`
	Metadata      string `json:"metadata"`
}

type fetchRequest struct {
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    []byte            `json:"body,omitempty"`
}

type fetchResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    []byte            `json:"body,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type callState struct {
	name   string
	hosts  []string
	result []byte
}

type callStateKey struct{}

var (
	runtimeMutex   sync.Mutex
	runtime        wazero.Runtime
	compiledByHash = map[[32]byte]wazero.CompiledModule{}
)

var fetchClient = &http.Client{
	Timeout: fetchTimeout,
	Transport: &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
			// the allowed host names could still resolve to this machine or its network
			Control: func(network, address string, c syscall.RawConn) error {
				host, _, err := net.SplitHostPort(address)
				if err != nil {
					return err
				}
				ip := net.ParseIP(host)
				if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
					return fmt.Errorf("actions cannot connect to %s", host)
				}
				return nil
			},
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		state, ok := req.Context().Value(callStateKey{}).(*callState)
		if !ok {
			return fmt.Errorf("no action is running")
		}
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		return checkFetchUrl(req.URL, state.hosts)
	},
}

func getCallState(ctx context.Context) *callState {
	return ctx.Value(callStateKey{}).(*callState)
}

func readMemory(m api.Module, ptr uint32, size uint32) []byte {
	data, ok := m.Memory().Read(ptr, size)
	if !ok {
		panic(fmt.Errorf("the action passed memory out of range"))
	}
	return bytes.Clone(data)
}

func checkFetchUrl(u *url.URL, hosts []string) error {
	if u.Scheme != "https" {
		return fmt.Errorf("actions can only send https requests, not %s", u.String())
	}
	if u.Port() != "" && u.Port() != "443" {
		return fmt.Errorf("actions can only send requests to port 443, not %s", u.Port())
	}
	for _, host := range hosts {
		if strings.EqualFold(u.Hostname(), host) {
			return nil
		}
	}
	return fmt.Errorf("the action is not allowed to send requests to %s", u.Hostname())
}

func fetch(ctx context.Context, state *callState, input []byte) *fetchResponse {
	req := fetchRequest{}
	err := json.Unmarshal(input, &req)
	if err != nil {
		return &fetchResponse{Error: err.Error()}
	}

	u, err := url.Parse(req.Url)
	if err != nil {
		return &fetchResponse{Error: err.Error()}
	}
	err = checkFetchUrl(u, state.hosts)
	if err != nil {
		return &fetchResponse{Error: err.Error()}
	}

	method := strings.ToUpper(req.Method)
	if method == "" {
		method = http.MethodGet
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(req.Body))
	if err != nil {
		return &fetchResponse{Error: err.Error()}
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := fetchClient.Do(httpReq)
	if err != nil {
		return &fetchResponse{Error: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return &fetchResponse{Error: err.Error()}
	}
	if len(body) > maxResponseSize {
		return &fetchResponse{Error: fmt.Sprintf("the response of %s is larger than 10 MB", u.Hostname())}
	}

	headers := map[string]string{}
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}
	return &fetchResponse{Status: resp.StatusCode, Headers: headers, Body: body}
}

func getRuntime(ctx context.Context) (wazero.Runtime, error) {
	if runtime != nil {
		return runtime, nil
	}

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithMemoryLimitPages(memoryLimitPages).WithCloseOnContextDone(true))
	_, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		r.Close(ctx)
		return nil, err
	}

	_, err = r.NewHostModuleBuilder("casdoor").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, ptr uint32, size uint32) uint32 {
		state := getCallState(ctx)
		output, err := json.Marshal(fetch(ctx, state, readMemory(m, ptr, size)))
		if err != nil {
			output, _ = json.Marshal(&fetchResponse{Error: err.Error()})
		}
		state.result = output
		return uint32(len(output))
	}).Export("fetch").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, ptr uint32) {
		state := getCallState(ctx)
		if !m.Memory().Write(ptr, state.result) {
			panic(fmt.Errorf("the action passed memory out of range"))
		}
		state.result = nil
	}).Export("read_result").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, ptr uint32, size uint32) {
		logs.Info("action %s: %s", getCallState(ctx).name, readMemory(m, ptr, size))
	}).Export("log").
		Instantiate(ctx)
	if err != nil {
		r.Close(ctx)
		return nil, err
	}

	runtime = r
	return runtime, nil
}

func compile(ctx context.Context, module []byte) (wazero.Runtime, wazero.CompiledModule, error) {
	runtimeMutex.Lock()
	defer runtimeMutex.Unlock()

	r, err := getRuntime(ctx)
	if err != nil {
		return nil, nil, err
	}

	hash := sha256.Sum256(module)
	if compiled, ok := compiledByHash[hash]; ok {
		return r, compiled, nil
	}

	compiled, err := r.CompileModule(ctx, module)
	if err != nil {
		return nil, nil, err
	}
	for _, name := range []string{"casdoor_alloc", "casdoor_free", "casdoor_call", "_initialize"} {
		if _, ok := compiled.ExportedFunctions()[name]; !ok {
			compiled.Close(ctx)
			return nil, nil, fmt.Errorf("the module does not export %s, build it with the Casdoor action SDK and -buildmode=c-shared", name)
		}
	}

	compiledByHash[hash] = compiled
	return r, compiled, nil
}

// Check compiles a module, so that a broken one is refused when it is installed rather than when it is first used.
func Check(module []byte) error {
	_, _, err := compile(context.Background(), module)
	return err
}

type logWriter struct {
	name string
}

func (w logWriter) Write(p []byte) (int, error) {
	logs.Warning("action %s: %s", w.name, strings.TrimSpace(string(p)))
	return len(p), nil
}

// Call runs method in a new instance of the module, which can only send requests to hosts.
func Call(name string, module []byte, hosts []string, method string, config *Config, params interface{}, out interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, callStateKey{}, &callState{name: name, hosts: hosts})

	r, compiled, err := compile(ctx, module)
	if err != nil {
		return err
	}

	moduleConfig := wazero.NewModuleConfig().
		WithName("").
		WithStartFunctions("_initialize").
		WithSysWalltime().
		WithSysNanotime().
		WithRandSource(rand.Reader).
		WithStdout(logWriter{name: name}).
		WithStderr(logWriter{name: name})
	m, err := r.InstantiateModule(ctx, compiled, moduleConfig)
	if err != nil {
		return fmt.Errorf("failed to start the action %s: %w", name, err)
	}
	defer m.Close(ctx)

	paramsData, err := json.Marshal(params)
	if err != nil {
		return err
	}
	input, err := json.Marshal(map[string]interface{}{
		"abi":    Abi,
		"method": method,
		"config": config,
		"params": json.RawMessage(paramsData),
	})
	if err != nil {
		return err
	}

	res, err := m.ExportedFunction("casdoor_alloc").Call(ctx, uint64(len(input)))
	if err != nil {
		return fmt.Errorf("the action %s failed: %w", name, err)
	}
	inputPtr := uint32(res[0])
	if !m.Memory().Write(inputPtr, input) {
		return fmt.Errorf("the action %s returned memory out of range", name)
	}

	res, err = m.ExportedFunction("casdoor_call").Call(ctx, uint64(inputPtr), uint64(len(input)))
	if err != nil {
		return fmt.Errorf("the action %s failed: %w", name, err)
	}
	outputPtr, outputSize := uint32(res[0]>>32), uint32(res[0])
	output, ok := m.Memory().Read(outputPtr, outputSize)
	if !ok {
		return fmt.Errorf("the action %s returned memory out of range", name)
	}

	result := struct {
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error"`
	}{}
	err = json.Unmarshal(output, &result)
	if err != nil {
		return fmt.Errorf("the action %s returned an invalid result: %w", name, err)
	}
	if result.Error != "" {
		return fmt.Errorf("%s", result.Error)
	}
	if out == nil || len(result.Result) == 0 {
		return nil
	}
	return json.Unmarshal(result.Result, out)
}
