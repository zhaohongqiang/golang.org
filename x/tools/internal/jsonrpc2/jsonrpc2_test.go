// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonrpc2_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"path"
	"reflect"
	"testing"

	"golang.org/x/tools/internal/jsonrpc2"
)

var logRPC = flag.Bool("logrpc", false, "Enable jsonrpc2 communication logging")

type callTest struct {
	method string
	params interface{}
	expect interface{}
}

var callTests = []callTest{
	{"no_args", nil, true},
	{"one_string", "fish", "got:fish"},
	{"one_number", 10, "got:10"},
	{"join", []string{"a", "b", "c"}, "a/b/c"},
	//TODO: expand the test cases
}

func (test *callTest) newResults() interface{} {
	switch e := test.expect.(type) {
	case []interface{}:
		var r []interface{}
		for _, v := range e {
			r = append(r, reflect.New(reflect.TypeOf(v)).Interface())
		}
		return r
	case nil:
		return nil
	default:
		return reflect.New(reflect.TypeOf(test.expect)).Interface()
	}
}

func (test *callTest) verifyResults(t *testing.T, results interface{}) {
	if results == nil {
		return
	}
	val := reflect.Indirect(reflect.ValueOf(results)).Interface()
	if !reflect.DeepEqual(val, test.expect) {
		t.Errorf("%v:Results are incorrect, got %+v expect %+v", test.method, val, test.expect)
	}
}

func TestPlainCall(t *testing.T) {
	ctx := context.Background()
	a, b := prepare(ctx, t, false)
	for _, test := range callTests {
		results := test.newResults()
		if err := a.Call(ctx, test.method, test.params, results); err != nil {
			t.Fatalf("%v:Call failed: %v", test.method, err)
		}
		test.verifyResults(t, results)
		if err := b.Call(ctx, test.method, test.params, results); err != nil {
			t.Fatalf("%v:Call failed: %v", test.method, err)
		}
		test.verifyResults(t, results)
	}
}

func TestHeaderCall(t *testing.T) {
	ctx := context.Background()
	a, b := prepare(ctx, t, true)
	for _, test := range callTests {
		results := test.newResults()
		if err := a.Call(ctx, test.method, test.params, results); err != nil {
			t.Fatalf("%v:Call failed: %v", test.method, err)
		}
		test.verifyResults(t, results)
		if err := b.Call(ctx, test.method, test.params, results); err != nil {
			t.Fatalf("%v:Call failed: %v", test.method, err)
		}
		test.verifyResults(t, results)
	}
}

<<<<<<< HEAD
func prepare(ctx context.Context, t *testing.T, withHeaders bool) (*testHandler, *testHandler) {
	a := &testHandler{t: t}
	b := &testHandler{t: t}
	a.reader, b.writer = io.Pipe()
	b.reader, a.writer = io.Pipe()
	for _, h := range []*testHandler{a, b} {
		h := h
		if withHeaders {
			h.stream = jsonrpc2.NewHeaderStream(h.reader, h.writer)
		} else {
			h.stream = jsonrpc2.NewStream(h.reader, h.writer)
		}
		args := []interface{}{jsonrpc2.Handler(handle)}
		if *logRPC {
			args = append(args, jsonrpc2.Log)
		}
		h.Conn = jsonrpc2.NewConn(ctx, h.stream, args...)
		go func() {
			defer func() {
				h.reader.Close()
				h.writer.Close()
			}()
			if err := h.Conn.Wait(ctx); err != nil {
				t.Fatalf("Stream failed: %v", err)
			}
		}()
	}
	return a, b
}

type testHandler struct {
	t      *testing.T
	reader *io.PipeReader
	writer *io.PipeWriter
	stream jsonrpc2.Stream
	*jsonrpc2.Conn
=======
func prepare(ctx context.Context, t *testing.T, withHeaders bool) (*jsonrpc2.Conn, *jsonrpc2.Conn) {
	aR, bW := io.Pipe()
	bR, aW := io.Pipe()
	a := run(ctx, t, withHeaders, aR, aW)
	b := run(ctx, t, withHeaders, bR, bW)
	return a, b
}

func run(ctx context.Context, t *testing.T, withHeaders bool, r io.ReadCloser, w io.WriteCloser) *jsonrpc2.Conn {
	var stream jsonrpc2.Stream
	if withHeaders {
		stream = jsonrpc2.NewHeaderStream(r, w)
	} else {
		stream = jsonrpc2.NewStream(r, w)
	}
	conn := jsonrpc2.NewConn(stream)
	conn.Handler = handle
	if *logRPC {
		conn.Logger = jsonrpc2.Log
	}
	go func() {
		defer func() {
			r.Close()
			w.Close()
		}()
		if err := conn.Run(ctx); err != nil {
			t.Fatalf("Stream failed: %v", err)
		}
	}()
	return conn
>>>>>>> bd25a1f6d07d2d464980e6a8576c1ed59bb3950a
}

func handle(ctx context.Context, c *jsonrpc2.Conn, r *jsonrpc2.Request) {
	switch r.Method {
	case "no_args":
		if r.Params != nil {
			c.Reply(ctx, r, nil, jsonrpc2.NewErrorf(jsonrpc2.CodeInvalidParams, "Expected no params"))
			return
		}
		c.Reply(ctx, r, true, nil)
	case "one_string":
		var v string
		if err := json.Unmarshal(*r.Params, &v); err != nil {
			c.Reply(ctx, r, nil, jsonrpc2.NewErrorf(jsonrpc2.CodeParseError, "%v", err.Error()))
			return
		}
		c.Reply(ctx, r, "got:"+v, nil)
	case "one_number":
		var v int
		if err := json.Unmarshal(*r.Params, &v); err != nil {
			c.Reply(ctx, r, nil, jsonrpc2.NewErrorf(jsonrpc2.CodeParseError, "%v", err.Error()))
			return
		}
		c.Reply(ctx, r, fmt.Sprintf("got:%d", v), nil)
	case "join":
		var v []string
		if err := json.Unmarshal(*r.Params, &v); err != nil {
			c.Reply(ctx, r, nil, jsonrpc2.NewErrorf(jsonrpc2.CodeParseError, "%v", err.Error()))
			return
		}
		c.Reply(ctx, r, path.Join(v...), nil)
	default:
		c.Reply(ctx, r, nil, jsonrpc2.NewErrorf(jsonrpc2.CodeMethodNotFound, "method %q not found", r.Method))
	}
}
