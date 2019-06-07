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
	toolman.Init(cfg.Flags(), toolman.StandardSignals(), toolman.LogFlushInterval(2*time.Second), toolman.LogToStderr())
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

	return s.ListenAndServe()
}
