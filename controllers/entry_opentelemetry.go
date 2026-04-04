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
	"github.com/casdoor/casdoor/object"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// @Title AddOtlpTrace
// @Tag OTLP API
// @Description receive otlp trace protobuf
// @Success 200 {object} string
// @router /api/v1/traces [post]
func (c *ApiController) AddOtlpTrace() {
	body := readProtobufBody(c.Ctx)
	if body == nil {
		return
	}
	providerName := resolveOpenClawProvider(c.Ctx)
	if providerName == "" {
		return
	}

	var req coltracepb.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		responseOtlpError(c.Ctx, 400, "bad protobuf: %v", err)
		return
	}

	message, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(&req)
	if err != nil {
		responseOtlpError(c.Ctx, 500, "marshal trace failed: %v", err)
		return
	}

	if _, err := object.AddEntry(object.NewTraceEntry(providerName, message)); err != nil {
		responseOtlpError(c.Ctx, 500, "save trace failed: %v", err)
		return
	}

	resp, _ := proto.Marshal(&coltracepb.ExportTraceServiceResponse{})
	c.Ctx.Output.Header("Content-Type", "application/x-protobuf")
	c.Ctx.Output.SetStatus(200)
	c.Ctx.Output.Body(resp)
}

// @Title AddOtlpMetrics
// @Tag OTLP API
// @Description receive otlp metrics protobuf
// @Success 200 {object} string
// @router /api/v1/metrics [post]
func (c *ApiController) AddOtlpMetrics() {
	body := readProtobufBody(c.Ctx)
	if body == nil {
		return
	}
	providerName := resolveOpenClawProvider(c.Ctx)
	if providerName == "" {
		return
	}

	var req colmetricspb.ExportMetricsServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		responseOtlpError(c.Ctx, 400, "bad protobuf: %v", err)
		return
	}

	message, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(&req)
	if err != nil {
		responseOtlpError(c.Ctx, 500, "marshal metrics failed: %v", err)
		return
	}

	if _, err := object.AddEntry(object.NewOtlpEntry(providerName, "metrics", message)); err != nil {
		responseOtlpError(c.Ctx, 500, "save metrics failed: %v", err)
		return
	}

	resp, _ := proto.Marshal(&colmetricspb.ExportMetricsServiceResponse{})
	c.Ctx.Output.Header("Content-Type", "application/x-protobuf")
	c.Ctx.Output.SetStatus(200)
	c.Ctx.Output.Body(resp)
}

// @Title AddOtlpLogs
// @Tag OTLP API
// @Description receive otlp logs protobuf
// @Success 200 {object} string
// @router /api/v1/logs [post]
func (c *ApiController) AddOtlpLogs() {
	body := readProtobufBody(c.Ctx)
	if body == nil {
		return
	}
	providerName := resolveOpenClawProvider(c.Ctx)
	if providerName == "" {
		return
	}

	var req collogspb.ExportLogsServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		responseOtlpError(c.Ctx, 400, "bad protobuf: %v", err)
		return
	}

	message, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(&req)
	if err != nil {
		responseOtlpError(c.Ctx, 500, "marshal logs failed: %v", err)
		return
	}

	if _, err := object.AddEntry(object.NewOtlpEntry(providerName, "log", message)); err != nil {
		responseOtlpError(c.Ctx, 500, "save logs failed: %v", err)
		return
	}

	resp, _ := proto.Marshal(&collogspb.ExportLogsServiceResponse{})
	c.Ctx.Output.Header("Content-Type", "application/x-protobuf")
	c.Ctx.Output.SetStatus(200)
	c.Ctx.Output.Body(resp)
}
