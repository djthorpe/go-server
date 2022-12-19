package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	// Package imports
	"golang.org/x/exp/slices"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// route is a (host, prefix, path, method) => hander mapping
type route struct {
	host    string
	prefix  string
	path    *regexp.Regexp
	fn      http.HandlerFunc
	methods []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewRoute(host_prefix string, path *regexp.Regexp, fn http.HandlerFunc, methods ...string) *route {
	// Check hander is not nil
	if fn == nil {
		return nil
	}

	// If the prefix does not contain a '/' then assume it is a host/path
	var host, prefix string
	if !strings.Contains(host_prefix, pathSeparator) {
		host = host_prefix
		prefix = pathSeparator
	} else if !strings.HasPrefix(host_prefix, pathSeparator) {
		if pair := strings.SplitN(host_prefix, pathSeparator, 2); len(pair) == 2 {
			host = pair[0]
			prefix = pair[1]
		} else {
			prefix = pair[0]
		}
	} else {
		prefix = host_prefix
	}

	// If methods is empty, default to GET
	if len(methods) == 0 {
		methods = []string{"GET"}
	}

	// Create route
	return &route{
		host:    normalizeHost(host),
		prefix:  normalizePath(prefix, (path != nil)),
		path:    path,
		fn:      fn,
		methods: methods,
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (route *route) String() string {
	str := "<httpserver-route"
	if host := route.Host(); host != "" {
		str += fmt.Sprintf(" host=%q", host)
	}
	if path := route.Path(); path != "" {
		str += fmt.Sprintf(" path=%q", path)
	}
	if methods := route.Methods(); len(methods) > 0 {
		str += fmt.Sprintf(" methods=%q", methods)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the path of the route, including the prefix
func (route *route) Path() string {
	str := route.prefix
	if route.path != nil {
		re := route.path.String()
		re = strings.TrimPrefix(re, "^")
		re = strings.TrimSuffix(re, "$")
		str += re
	}
	return str
}

// Return a wild-carded host
func (route *route) Host() string {
	return "*" + route.host
}

// Return the methods of the route
func (route *route) Methods() []string {
	return route.methods
}

// Return true if a request matches the (prefix,path) pair.
// Also returns the parameters from the path if the route has a path,
// and the remaining path.
func (route *route) MatchesPath(path string) ([]string, string, bool) {
	path = normalizePath(path, false)
	if path == route.prefix {
		return nil, "/", true
	}

	// Check for default route: this is the route that matches everything
	if !strings.HasPrefix(path, route.prefix) {
		return nil, "", false
	} else if route.path == nil {
		return nil, normalizePath(path[len(route.prefix):], false), true
	}

	// Check with a regular expression
	relpath := path[len(route.prefix):]
	if params := route.path.FindStringSubmatch(relpath); params == nil {
		return nil, "", false
	} else {
		return params[1:], normalizePath(relpath, false), true
	}
}

// Return true if a request matches a host. If the host is empty, then
// it matches any host.
func (route *route) MatchesHost(host string) bool {
	if route.host == "" {
		return true
	}

	// Add a . at the beginning of the request host
	host = normalizeHost(host)

	// No host match if the route host is longer than the request host
	if len(route.host) > len(host) {
		return false
	}

	// Matches if the route host is a suffix of the request host
	return strings.HasSuffix(host, route.host)
}

// Return true if a request method matches the route
func (route *route) MatchesMethod(method string) bool {
	return slices.Contains(route.methods, method)
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Add a / to the beginning and optionally to the end of the path
func normalizePath(path string, atEnd bool) string {
	if !strings.HasPrefix(path, pathSeparator) {
		path = pathSeparator + path
	}
	if path == pathSeparator {
		return path
	} else if !atEnd && strings.HasSuffix(path, pathSeparator) {
		path = strings.TrimSuffix(path, pathSeparator)
	} else if atEnd && !strings.HasSuffix(path, pathSeparator) {
		path = path + pathSeparator
	}
	return path
}

// Add a . to the beginning of the host, and remove from the end
func normalizeHost(host string) string {
	host = strings.TrimSuffix(host, hostSeparator)
	if host == "" {
		return host
	}
	if !strings.HasPrefix(host, hostSeparator) {
		host = hostSeparator + host
	}
	return strings.ToLower(host)
}
