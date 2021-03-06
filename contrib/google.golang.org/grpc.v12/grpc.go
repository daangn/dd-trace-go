// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

//go:generate protoc -I . fixtures_test.proto --go_out=. --go-grpc_out=.

// Package grpc provides functions to trace the google.golang.org/grpc package v1.2.
package grpc // import "gopkg.in/daangn/dd-trace-go.v1/contrib/google.golang.org/grpc.v12"

import (
	"math"
	"net"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"gopkg.in/daangn/dd-trace-go.v1/contrib/google.golang.org/internal/grpcutil"
	"gopkg.in/daangn/dd-trace-go.v1/ddtrace"
	"gopkg.in/daangn/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/daangn/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/daangn/dd-trace-go.v1/internal/globalconfig"
)

// UnaryServerInterceptor will trace requests to the given grpc server.
func UnaryServerInterceptor(opts ...InterceptorOption) grpc.UnaryServerInterceptor {
	cfg := new(interceptorConfig)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	if cfg.serviceName == "" {
		cfg.serviceName = "grpc.server"
		if svc := globalconfig.ServiceName(); svc != "" {
			cfg.serviceName = svc
		}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		span, ctx := startSpanFromContext(ctx, info.FullMethod, cfg.serviceName, cfg.analyticsRate)
		resp, err := handler(ctx, req)
		span.Finish(tracer.WithError(err))
		return resp, err
	}
}

func startSpanFromContext(ctx context.Context, method, service string, rate float64) (ddtrace.Span, context.Context) {
	opts := []ddtrace.StartSpanOption{
		tracer.ServiceName(service),
		tracer.ResourceName(method),
		tracer.Tag(tagMethod, method),
		tracer.SpanType(ext.AppTypeRPC),
		tracer.Measured(),
	}
	if !math.IsNaN(rate) {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, rate))
	}
	md, _ := metadata.FromIncomingContext(ctx) // nil is ok
	if sctx, err := tracer.Extract(grpcutil.MDCarrier(md)); err == nil {
		opts = append(opts, tracer.ChildOf(sctx))
	}
	return tracer.StartSpanFromContext(ctx, "grpc.server", opts...)
}

// UnaryClientInterceptor will add tracing to a grpc client.
func UnaryClientInterceptor(opts ...InterceptorOption) grpc.UnaryClientInterceptor {
	cfg := new(interceptorConfig)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	if cfg.serviceName == "" {
		cfg.serviceName = "grpc.client"
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var (
			span ddtrace.Span
			p    peer.Peer
		)
		spanopts := []ddtrace.StartSpanOption{
			tracer.Tag(tagMethod, method),
			tracer.SpanType(ext.AppTypeRPC),
		}
		if !math.IsNaN(cfg.analyticsRate) {
			spanopts = append(spanopts, tracer.Tag(ext.EventSampleRate, cfg.analyticsRate))
		}
		span, ctx = tracer.StartSpanFromContext(ctx, "grpc.client", spanopts...)
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		_ = tracer.Inject(span.Context(), grpcutil.MDCarrier(md))
		ctx = metadata.NewOutgoingContext(ctx, md)
		opts = append(opts, grpc.Peer(&p))
		err := invoker(ctx, method, req, reply, cc, opts...)
		if p.Addr != nil {
			addr := p.Addr.String()
			host, port, err := net.SplitHostPort(addr)
			if err == nil {
				if host != "" {
					span.SetTag(ext.TargetHost, host)
				}
				span.SetTag(ext.TargetPort, port)
			}
		}
		span.SetTag(tagCode, status.Code(err).String())
		span.Finish(tracer.WithError(err))
		return err
	}
}
