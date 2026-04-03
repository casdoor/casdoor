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

package controllers

import (
	"fmt"
	"io"
	"strings"

	"github.com/casdoor/casdoor/object"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// AddOtlpEntry
// @Title AddTrace
// @Tag OTLP API
// @Description receive otlp trace protobuf
// @Success 200 {object} string
// @router /api/v1/traces [post]
func (c *ApiController) AddTrace() {
	if !strings.HasPrefix(c.Ctx.Input.Header("Content-Type"), "application/x-protobuf") {
		c.Ctx.Output.SetStatus(415)
		c.Ctx.Output.Body([]byte("unsupported content type"))
		return
	}

	body, err := io.ReadAll(c.Ctx.Request.Body)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte("read body failed"))
		return
	}

	var req coltracepb.ExportTraceServiceRequest

	if err := proto.Unmarshal(body, &req); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte(fmt.Sprintf("bad protobuf: %v", err)))
		return
	}

	opts := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}

	message, err := opts.Marshal(&req)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte(fmt.Sprintf("marshal trace failed: %v", err)))
		return
	}

	entry := object.NewTraceEntry(message)

	if _, err := object.AddEntry(entry); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte(fmt.Sprintf("save trace failed: %v", err)))
		return
	}

	resp := &coltracepb.ExportTraceServiceResponse{}
	respBytes, _ := proto.Marshal(resp)

	c.Ctx.Output.Header("Content-Type", "application/x-protobuf")
	c.Ctx.Output.SetStatus(200)
	c.Ctx.Output.Body(respBytes)
}
