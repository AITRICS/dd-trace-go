// Package mux provides tracing functions for tracing the gorilla/mux package (https://github.com/gorilla/mux).
package mux

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/DataDog/dd-trace-go/contrib/internal"
	"github.com/DataDog/dd-trace-go/tracer"
	"github.com/DataDog/dd-trace-go/tracer/ext"
)

// Router registers routes to be matched and dispatches a handler.
type Router struct {
	*mux.Router
	tracer  *tracer.Tracer
	service string
}

// NewRouterWithTracer returns a new router instance traced with the global tracer.
func NewRouter() *Router {
	return NewRouterWithServiceName("mux.router", tracer.DefaultTracer)
}

// NewRouterWithServiceName returns a new router instance which traces under the given service
// name.
//
// TODO(gbbr): Remove tracer parameter once we switch to OpenTracing.
func NewRouterWithServiceName(service string, t *tracer.Tracer) *Router {
	t.SetServiceInfo(service, "gorilla/mux", ext.AppTypeWeb)
	return &Router{
		Router:  mux.NewRouter(),
		tracer:  t,
		service: service,
	}
}

// ServeHTTP dispatches the request to the handler
// whose pattern most closely matches the request URL.
// We only need to rewrite this function to be able to trace
// all the incoming requests to the underlying multiplexer
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		match mux.RouteMatch
		route string
		err   error
	)
	// get the resource associated to this request
	if r.Match(req, &match) {
		route, err = match.Route.GetPathTemplate()
		if err != nil {
			route = "unknown"
		}
	} else {
		route = "unknown"
	}
	resource := req.Method + " " + route
	internal.TraceAndServe(r.Router, w, req, r.service, resource, r.tracer)
}
