package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	// Modules
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Renderers []string `yaml:"plugins"`
}

type plugin struct {
	mimetypes map[string]Renderer
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	pathSeparator = string(os.PathSeparator)
)

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the template module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Load configuration
	var cfg Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	}

	// Set renderers
	p.mimetypes = make(map[string]Renderer)
	for _, name := range cfg.Renderers {
		if renderer, ok := provider.GetPlugin(ctx, name).(Renderer); !ok {
			provider.Printf(ctx, "Failed to load renderer: %q", name)
			return nil
		} else {
			for _, mimetype := range renderer.Mimetypes() {
				if err := p.setRenderer(mimetype, renderer); err != nil {
					provider.Printf(ctx, err.Error())
				}
			}
		}
	}

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<renderer"
	for key, renderer := range p.mimetypes {
		str += fmt.Sprintf(" %q=%v", key, renderer)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "renderer"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	<-ctx.Done()
	return nil
}

// Return default mimetypes and file extensions handled by this renderer
func (p *plugin) Mimetypes() []string {
	result := make([]string, 0, len(p.mimetypes))
	for key := range p.mimetypes {
		result = append(result, key)
	}
	return result
}

// Render a file into a document, with reader and optional file info
func (p *plugin) Read(ctx context.Context, r io.Reader, mimetype string, info fs.FileInfo) (Document, error) {
	if mimetype != "" {
		if renderer := p.getRenderer(mimetype); renderer != nil {
			return renderer.Read(ctx, r, mimetype, info)
		}
	}
	if info != nil {
		if ext := filepath.Ext(info.Name()); ext != "" {
			if renderer := p.getRenderer(ext); renderer != nil {
				return renderer.Read(ctx, r, mimetype, info)
			}
		}
		if renderer := p.getRenderer(info.Name()); renderer != nil {
			return renderer.Read(ctx, r, mimetype, info)
		}
	}
	return nil, ErrNotFound.Withf("Read: no renderer found for %q", info.Name())
}

// Render a directory into a document, with optional file info
func (p *plugin) ReadDir(ctx context.Context, dir fs.ReadDirFile, info fs.FileInfo) (Document, error) {
	if renderer := p.getRenderer(pathSeparator); renderer == nil {
		return nil, ErrNotFound.With("ReadDir: no renderer found")
	} else {
		return renderer.ReadDir(ctx, dir, info)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *plugin) setRenderer(key string, renderer Renderer) error {
	key = strings.ToLower(key)
	if r, exists := p.mimetypes[key]; exists {
		return ErrDuplicateEntry.Withf("%q will be handled by %q", key, r)
	}
	p.mimetypes[key] = renderer
	return nil
}

func (p *plugin) getRenderer(key string) Renderer {
	key = strings.ToLower(key)
	if r, exists := p.mimetypes[key]; exists {
		return r
	} else {
		return nil
	}
}
