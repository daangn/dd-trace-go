// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// Package memcache provides functions to trace the daangn/x/memcache package (https://github.com/daangn/x/memcache).
//
// `WrapClient` will wrap a memcache `Client` and return a new struct with all
// the same methods, so should be seamless for existing applications. It also
// has an additional `WithContext` method which can be used to connect a span
// to an existing trace.
package memcache // import "gopkg.in/daangn/dd-trace-go.v1/contrib/daangn/x/memcache"

import (
	"context"
	"math"

	gomemcache "github.com/daangn/gomemcache/memcache"
	"github.com/daangn/x/memcache"

	"gopkg.in/daangn/dd-trace-go.v1/ddtrace"
	"gopkg.in/daangn/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/daangn/dd-trace-go.v1/ddtrace/tracer"
)

// WrapClient wraps a memcache.Client so that all requests are traced using the
// default tracer with the service name "memcached".
func WrapClient(client *memcache.Client, opts ...ClientOption) *Client {
	cfg := new(clientConfig)
	defaults(cfg)
	for _, opt := range opts {
		opt(cfg)
	}
	return &Client{
		Client:  client,
		cfg:     cfg,
		context: context.Background(),
	}
}

// A Client is used to trace requests to the memcached server.
type Client struct {
	*memcache.Client
	cfg     *clientConfig
	context context.Context
}

// WithContext creates a copy of the Client with the given context.
func (c *Client) WithContext(ctx context.Context) *Client {
	return &Client{
		Client:  c.Client,
		cfg:     c.cfg,
		context: ctx,
	}
}

// startSpan starts a span from the context set with WithContext.
func (c *Client) startSpan(resourceName string) ddtrace.Span {
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeMemcached),
		tracer.ServiceName(c.cfg.serviceName),
		tracer.ResourceName(resourceName),
	}
	if !math.IsNaN(c.cfg.analyticsRate) {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, c.cfg.analyticsRate))
	}
	span, _ := tracer.StartSpanFromContext(c.context, operationName, opts...)
	return span
}

// wrapped methods:

// Add invokes and traces Client.Add.
func (c *Client) Add(item *gomemcache.Item) error {
	span := c.startSpan("Add")
	err := c.Client.Add(c.context, item)
	span.Finish(tracer.WithError(err))
	return err
}

// CompareAndSwap invokes and traces Client.CompareAndSwap.
func (c *Client) CompareAndSwap(item *gomemcache.Item) error {
	span := c.startSpan("CompareAndSwap")
	err := c.Client.CompareAndSwap(c.context, item)
	span.Finish(tracer.WithError(err))
	return err
}

// Decrement invokes and traces Client.Decrement.
func (c *Client) Decrement(key string, delta uint64) (newValue uint64, err error) {
	span := c.startSpan("Decrement")
	newValue, err = c.Client.Decrement(c.context, key, delta)
	span.Finish(tracer.WithError(err))
	return newValue, err
}

// Delete invokes and traces Client.Delete.
func (c *Client) Delete(key string) error {
	span := c.startSpan("Delete")
	err := c.Client.Delete(c.context, key)
	span.Finish(tracer.WithError(err))
	return err
}

// DeleteAll invokes and traces Client.DeleteAll.
func (c *Client) DeleteAll() error {
	span := c.startSpan("DeleteAll")
	err := c.Client.DeleteAll(c.context)
	span.Finish(tracer.WithError(err))
	return err
}

// FlushAll invokes and traces Client.FlushAll.
func (c *Client) FlushAll() error {
	span := c.startSpan("FlushAll")
	err := c.Client.FlushAll()
	span.Finish(tracer.WithError(err))
	return err
}

// Get invokes and traces Client.Get.
func (c *Client) Get(key string) (item *gomemcache.Item, err error) {
	span := c.startSpan("Get")
	item, err = c.Client.Get(c.context, key)
	span.Finish(tracer.WithError(err))
	return item, err
}

// GetMulti invokes and traces Client.GetMulti.
func (c *Client) GetMulti(keys []string) (map[string]*gomemcache.Item, error) {
	span := c.startSpan("GetMulti")
	items, err := c.Client.GetMulti(c.context, keys)
	span.Finish(tracer.WithError(err))
	return items, err
}

// Increment invokes and traces Client.Increment.
func (c *Client) Increment(key string, delta uint64) (newValue uint64, err error) {
	span := c.startSpan("Increment")
	newValue, err = c.Client.Increment(c.context, key, delta)
	span.Finish(tracer.WithError(err))
	return newValue, err
}

// Replace invokes and traces Client.Replace.
func (c *Client) Replace(item *gomemcache.Item) error {
	span := c.startSpan("Replace")
	err := c.Client.Replace(c.context, item)
	span.Finish(tracer.WithError(err))
	return err
}

// Set invokes and traces Client.Set.
func (c *Client) Set(item *gomemcache.Item) error {
	span := c.startSpan("Set")
	err := c.Client.Set(c.context, item)
	span.Finish(tracer.WithError(err))
	return err
}

// Touch invokes and traces Client.Touch.
func (c *Client) Touch(key string, seconds int32) error {
	span := c.startSpan("Touch")
	err := c.Client.Touch(c.context, key, seconds)
	span.Finish(tracer.WithError(err))
	return err
}
