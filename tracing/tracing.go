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

// Package tracing exports the traces of Casdoor's own requests to an OpenTelemetry collector (SkyWalking, Jaeger,
// Alibaba Cloud ARMS and others take OTLP over HTTP). It needs a commercial edition license and the
// "otelEndpoint" config item, e.g. http://localhost:4318.
package tracing

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/edition"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var (
	enabled    bool
	tracer     trace.Tracer = noop.NewTracerProvider().Tracer("casdoor")
	propagator              = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
)

func Enabled() bool {
	return enabled
}

// Init starts the exporter when "otelEndpoint" is set and the license allows it
func Init() {
	endpoint := strings.TrimSpace(conf.GetConfigString("otelEndpoint"))
	if endpoint == "" {
		return
	}
	if err := edition.Require(edition.Commercial, "OpenTelemetry tracing"); err != nil {
		logs.Warn("tracing is off: %s", err.Error())
		return
	}

	options := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(strings.TrimRight(endpoint, "/") + "/v1/traces")}
	if headers := conf.GetConfigString("otelHeaders"); headers != "" {
		m := map[string]string{}
		for _, pair := range strings.Split(headers, ",") {
			if k, v, ok := strings.Cut(pair, "="); ok {
				m[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
		options = append(options, otlptracehttp.WithHeaders(m))
	}
	exporter, err := otlptracehttp.New(context.Background(), options...)
	if err != nil {
		logs.Warn("tracing is off: %s", err.Error())
		return
	}

	serviceName := conf.GetConfigString("otelServiceName")
	if serviceName == "" {
		serviceName = "casdoor"
	}
	res, err := resource.Merge(resource.Default(), resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(serviceName)))
	if err != nil {
		res = resource.Default()
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagator)
	tracer = provider.Tracer("casdoor")
	enabled = true
	logs.Info("tracing to %s as %s", endpoint, serviceName)
}

// StartRequest starts the server span of a request, continuing the trace of an upstream gateway or application
// that sends a W3C traceparent header
func StartRequest(r *http.Request, route string) (context.Context, trace.Span) {
	ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	return tracer.Start(ctx, r.Method+" "+route,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.HTTPRequestMethodKey.String(r.Method),
			semconv.URLPath(r.URL.Path),
			semconv.ServerAddress(r.Host),
			semconv.UserAgentOriginal(r.UserAgent()),
		),
	)
}

// EndRequest ends the server span with the response status
func EndRequest(span trace.Span, status int, attributes ...attribute.KeyValue) {
	span.SetAttributes(semconv.HTTPResponseStatusCode(status))
	span.SetAttributes(attributes...)
	if status >= 500 {
		span.SetStatus(codes.Error, http.StatusText(status))
	}
	span.End()
}

// Start starts a child span, e.g. a step of the sign-in
func Start(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	return tracer.Start(ctx, name, trace.WithAttributes(attributes...))
}

// End ends a span, marking it failed when err is not nil
func End(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
