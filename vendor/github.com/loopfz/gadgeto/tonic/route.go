package tonic

import (
	"errors"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

// A Route contains information about a tonic-enabled route.
type Route struct {
	gin.RouteInfo

	defaultStatusCode int
	description       string
	summary           string
	deprecated        bool

	// Handler is the route handler.
	handler reflect.Value

	// HandlerType is the type of the route handler.
	handlerType reflect.Type

	// inputType is the type of the input object.
	// This can be nil if the handler use none.
	inputType reflect.Type

	// outputType is the type of the output object.
	// This can be nil if the handler use none.
	outputType reflect.Type
}

// GetVerb returns the HTTP verb of the route.
func (r *Route) GetVerb() string { return r.Method }

// GetPath returns the path of the route.
func (r *Route) GetPath() string { return r.Path }

// GetDescription returns the description of the route.
func (r *Route) GetDescription() string { return r.description }

// GetSummary returns the summary of the route.
func (r *Route) GetSummary() string { return r.summary }

// GetDefaultStatusCode returns the default status code of the route.
func (r *Route) GetDefaultStatusCode() int { return r.defaultStatusCode }

// GetHandler returns the handler of the route.
func (r *Route) GetHandler() reflect.Value { return r.handler }

// GetDeprecated returns the deprecated flag of the route.
func (r *Route) GetDeprecated() bool { return r.deprecated }

// InputType returns the input type of the handler.
// If the type is a pointer to a concrete type, it
// is dereferenced.
func (r *Route) InputType() reflect.Type {
	if in := r.inputType; in != nil && in.Kind() == reflect.Ptr {
		return in.Elem()
	}
	return r.inputType
}

// OutputType returns the output type of the handler.
// If the type is a pointer to a concrete type, it
// is dereferenced.
func (r *Route) OutputType() reflect.Type {
	if out := r.outputType; out != nil && out.Kind() == reflect.Ptr {
		return out.Elem()
	}
	return r.outputType
}

// HandlerName returns the name of the route handler.
func (r *Route) HandlerName() string {
	parts := strings.Split(r.HandlerNameWithPackage(), ".")
	return parts[len(parts)-1]
}

// HandlerNameWithPackage returns the full name of the rout
// handler with its package path.
func (r *Route) HandlerNameWithPackage() string {
	f := runtime.FuncForPC(r.handler.Pointer()).Name()
	parts := strings.Split(f, "/")
	return parts[len(parts)-1]
}

// GetTags generates a list of tags for the swagger spec
// from one route definition.
// Currently it only takes the first path of the route as the tag.
func (r *Route) GetTags() []string {
	tags := make([]string, 0, 1)
	paths := strings.SplitN(r.GetPath(), "/", 3)
	if len(paths) > 1 {
		tags = append(tags, paths[1])
	}
	return tags
}

// GetRouteByHandler returns the route informations of
// the given wrapped handler.
func GetRouteByHandler(h gin.HandlerFunc) (*Route, error) {
	ctx := &gin.Context{}
	ctx.Set(tonicWantRouteInfos, nil)

	funcsMu.Lock()
	defer funcsMu.Unlock()
	if _, ok := funcs[runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()]; !ok {
		return nil, errors.New("handler is not wrapped by tonic")
	}
	h(ctx)

	i, ok := ctx.Get(tonicRoutesInfos)
	if !ok {
		return nil, errors.New("failed to retrieve handler infos")
	}
	route, ok := i.(*Route)
	if !ok {
		return nil, errors.New("failed to retrieve handler infos")
	}
	return route, nil
}
