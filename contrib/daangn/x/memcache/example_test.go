// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package memcache_test

import (
	"context"

	gomemcache "github.com/daangn/gomemcache/memcache"
	"github.com/daangn/x/memcache"

	memcachetrace "gopkg.in/daangn/dd-trace-go.v1/contrib/daangn/x/memcache"
	"gopkg.in/daangn/dd-trace-go.v1/ddtrace/tracer"
)

func Example() {
	span, ctx := tracer.StartSpanFromContext(context.Background(), "parent.request",
		tracer.ServiceName("web"),
		tracer.ResourceName("/home"),
	)
	defer span.Finish()

	client, _ := memcache.New(context.TODO(), &memcache.Options{
		CfgEp: "127.0.0.1:11211",
	})
	mc := memcachetrace.WrapClient(client)
	// you can use WithContext to set the parent span
	mc.WithContext(ctx).Set(&gomemcache.Item{Key: "my key", Value: []byte("my value")})
}
