package static

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"sync"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	plugin "github.com/mutablelogic/go-server/plugin"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type static struct {
	task.Task
	sync.RWMutex

	label   string
	handler http.Handler
}

var _ plugin.Gateway = (*static)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new static serving task
func NewWithPlugin(p iface.Plugin, fs fs.FS) (*static, error) {
	static := new(static)
	static.label = p.Label()

	// Create the handler
	if fs == nil {
		return nil, ErrBadParameter.With("fs")
	} else {
		static.handler = http.FileServer(http.FS(fs))
	}

	// Return success
	return static, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (static *static) String() string {
	str := "<httpserver-static"
	if label := static.Label(); label != "" {
		str += fmt.Sprintf(" label=%q", label)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Label returns the label of the router
func (static *static) Label() string {
	return static.label
}

// Description returns the label of the router
func (static *static) Description() string {
	return "Serves static files"
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register routes for router
func (static *static) RegisterHandlers(parent context.Context, router plugin.Router) {
	// GET /
	//   Return list of prefixes and their handlers
	router.AddHandler(parent, nil, static.handler.ServeHTTP)
}
