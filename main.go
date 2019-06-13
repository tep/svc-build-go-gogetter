// Copyright 2019 Timothy E. Peoples
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package main

import (
	"context"
	"time"

	"github.com/kr/pretty"

	"toolman.org/base/log/v2"
	"toolman.org/base/toolman/v2"

	"toolman.org/svc/build/go/gogetter/internal/config"
	"toolman.org/svc/build/go/gogetter/internal/server"
	"toolman.org/svc/build/go/gogetter/internal/xlat"
)

func main() {
	cfg := config.New()

	toolman.Init(
		cfg.Flags(),
		toolman.StandardSignals(),
		toolman.LogFlushInterval(2*time.Second),
		toolman.LogDir(cfg.LogDir))

	ctx := context.Background()

	if err := run(ctx, cfg); err != nil {
		log.Exit(err)
	}
}

func run(ctx context.Context, cfg *config.Config) error {
	if err := cfg.Load(); err != nil {
		return err
	}

	c := *cfg
	c.Config = nil
	pretty.Println(c)

	x, err := xlat.New(cfg)
	if err != nil {
		return err
	}

	if err := x.Discover(ctx); err != nil {
		return err
	}

	x.Dump()

	s := server.New(cfg, x)

	toolman.RegisterShutdown(func() {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		s.Shutdown(ctx)
	})

	return s.ListenAndServe()
}
