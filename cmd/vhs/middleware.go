package main

import (
	"log"
	"os"

	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/session"
)

func mustStartMiddleware(ctx *session.Context) *middleware.Middleware {
	if ctx.Config.Middleware == "" {
		return nil
	}

	m, err := middleware.New(ctx, ctx.Config.Middleware, os.Stderr)
	if err != nil {
		log.Fatalf("failed to create middleware: %v\n", err)
	}

	go func() {
		if err := m.Start(); err != nil {
			log.Fatalf("failed to start middleware: %v\n", err)
		}
	}()

	return m
}
