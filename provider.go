package server

import (
	"context"
	"net/http"
	"regexp"
)

// Provider provides services to a module
type Provider interface {
	Logger
	Router

	// Plugins returns a list of registered plugin names
	Plugins() []string

	// GetPlugin returns a named plugin or nil if not available
	GetPlugin(context.Context, string) Plugin

	// GetConfig populates yaml config
	GetConfig(context.Context, interface{}) error
}

// Plugin provides handlers to server
type Plugin interface {
	// Run plugin background tasks until cancelled
	Run(context.Context) error
}

// Logger providers a logging interface
type Logger interface {
	Print(context.Context, ...interface{})
	Printf(context.Context, string, ...interface{})
}

// Router allows handlers to be added for serving URL paths
type Router interface {
	AddHandler(context.Context, http.Handler, ...string) error
	AddHandlerFunc(context.Context, http.HandlerFunc, ...string) error
	AddHandlerFuncEx(context.Context, *regexp.Regexp, http.HandlerFunc, ...string) error
}
