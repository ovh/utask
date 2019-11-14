package fizz

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/wI2L/fizz/openapi"
)

const ctxOpenAPIOperation = "_ctx_openapi_operation"

// Primitive type helpers.
var (
	Integer  int32
	Long     int64
	Float    float32
	Double   float64
	String   string
	Byte     []byte
	Binary   []byte
	Boolean  bool
	DateTime time.Time
)

// Fizz is an abstraction of a Gin engine that wraps the
// routes handlers with Tonic and generates an OpenAPI
// 3.0 specification from it.
type Fizz struct {
	gen    *openapi.Generator
	engine *gin.Engine
	*RouterGroup
}

// RouterGroup is an abstraction of a Gin router group.
type RouterGroup struct {
	group       *gin.RouterGroup
	gen         *openapi.Generator
	Name        string
	Description string
}

// New creates a new Fizz wrapper for
// a default Gin engine.
func New() *Fizz {
	return NewFromEngine(gin.New())
}

// NewFromEngine creates a new Fizz wrapper
// from an existing Gin engine.
func NewFromEngine(e *gin.Engine) *Fizz {
	// Create a new spec with the config
	// based on tonic internals.
	gen, _ := openapi.NewGenerator(
		&openapi.SpecGenConfig{
			ValidatorTag:      tonic.ValidationTag,
			PathLocationTag:   tonic.PathTag,
			QueryLocationTag:  tonic.QueryTag,
			HeaderLocationTag: tonic.HeaderTag,
			EnumTag:           tonic.EnumTag,
			DefaultTag:        tonic.DefaultTag,
		},
	)
	return &Fizz{
		engine: e,
		gen:    gen,
		RouterGroup: &RouterGroup{
			group: &e.RouterGroup,
			gen:   gen,
		},
	}
}

// ServeHTTP implements http.HandlerFunc for Fizz.
func (f *Fizz) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.engine.ServeHTTP(w, r)
}

// Engine returns the underlying Gin engine.
func (f *Fizz) Engine() *gin.Engine {
	return f.engine
}

// Generator returns the underlying OpenAPI generator.
func (f *Fizz) Generator() *openapi.Generator {
	return f.gen
}

// Errors returns the errors that may have occurred
// during the spec generation.
func (f *Fizz) Errors() []error {
	return f.gen.Errors()
}

// Group creates a new group of routes.
func (g *RouterGroup) Group(path, name, description string, handlers ...gin.HandlerFunc) *RouterGroup {
	// Create the tag in the specification
	// for this groups.
	g.gen.AddTag(name, description)

	return &RouterGroup{
		gen:         g.gen,
		group:       g.group.Group(path, handlers...),
		Name:        name,
		Description: description,
	}
}

// Use adds middleware to the group.
func (g *RouterGroup) Use(handlers ...gin.HandlerFunc) {
	g.group.Use(handlers...)
}

// GET is a shortcut to register a new handler with the GET method.
func (g *RouterGroup) GET(path string, infos []OperationOption, handlers ...gin.HandlerFunc) *RouterGroup {
	return g.Handle(path, "GET", infos, handlers...)
}

// POST is a shortcut to register a new handler with the POST method.
func (g *RouterGroup) POST(path string, infos []OperationOption, handlers ...gin.HandlerFunc) *RouterGroup {
	return g.Handle(path, "POST", infos, handlers...)
}

// PUT is a shortcut to register a new handler with the PUT method.
func (g *RouterGroup) PUT(path string, infos []OperationOption, handlers ...gin.HandlerFunc) *RouterGroup {
	return g.Handle(path, "PUT", infos, handlers...)
}

// PATCH is a shortcut to register a new handler with the PATCH method.
func (g *RouterGroup) PATCH(path string, infos []OperationOption, handlers ...gin.HandlerFunc) *RouterGroup {
	return g.Handle(path, "PATCH", infos, handlers...)
}

// DELETE is a shortcut to register a new handler with the DELETE method.
func (g *RouterGroup) DELETE(path string, infos []OperationOption, handlers ...gin.HandlerFunc) *RouterGroup {
	return g.Handle(path, "DELETE", infos, handlers...)
}

// OPTIONS is a shortcut to register a new handler with the OPTIONS method.
func (g *RouterGroup) OPTIONS(path string, infos []OperationOption, handlers ...gin.HandlerFunc) *RouterGroup {
	return g.Handle(path, "OPTIONS", infos, handlers...)
}

// HEAD is a shortcut to register a new handler with the HEAD method.
func (g *RouterGroup) HEAD(path string, infos []OperationOption, handlers ...gin.HandlerFunc) *RouterGroup {
	return g.Handle(path, "HEAD", infos, handlers...)
}

// TRACE is a shortcut to register a new handler with the TRACE method.
func (g *RouterGroup) TRACE(path string, infos []OperationOption, handlers ...gin.HandlerFunc) *RouterGroup {
	return g.Handle(path, "TRACE", infos, handlers...)
}

// Handle registers a new request handler that is wrapped
// with Tonic and documented in the OpenAPI specification.
func (g *RouterGroup) Handle(path, method string, infos []OperationOption, handlers ...gin.HandlerFunc) *RouterGroup {
	oi := &openapi.OperationInfo{}
	for _, info := range infos {
		info(oi)
	}
	type wrap struct {
		h gin.HandlerFunc
		r *tonic.Route
	}
	var wrapped []wrap

	// Find the handlers wrapped with Tonic.
	for _, h := range handlers {
		r, err := tonic.GetRouteByHandler(h)
		if err == nil {
			wrapped = append(wrapped, wrap{h: h, r: r})
		}
	}
	// Check that no more that one tonic-wrapped handler
	// is registered for this operation.
	if len(wrapped) > 1 {
		panic(fmt.Sprintf("multiple tonic-wrapped handler used for operation %s %s", method, path))
	}
	// If we have a tonic-wrapped handler, generate the
	// specification of this operation.
	if len(wrapped) == 1 {
		hfunc := wrapped[0].r

		// Set an operation ID if none is provided.
		if oi.ID == "" {
			oi.ID = hfunc.HandlerName()
		}
		oi.StatusCode = hfunc.GetDefaultStatusCode()

		// Set an input type if provided.
		it := hfunc.InputType()
		if oi.InputModel != nil {
			it = reflect.TypeOf(oi.InputModel)
		}

		// Consolidate path for OpenAPI spec.
		operationPath := joinPaths(g.group.BasePath(), path)

		// Add operation to the OpenAPI spec.
		operation, err := g.gen.AddOperation(operationPath, method, g.Name, it, hfunc.OutputType(), oi)
		if err != nil {
			panic(fmt.Sprintf(
				"error while generating OpenAPI spec on operation %s %s: %s",
				method, path, err,
			))
		}
		// If an operation was generated for the handler,
		// wrap the Tonic-wrapped handled with a closure
		// to inject it into the Gin context.
		if operation != nil {
			for i, h := range handlers {
				if funcEqual(h, wrapped[0].h) {
					orig := h // copy the original func
					handlers[i] = func(c *gin.Context) {
						c.Set(ctxOpenAPIOperation, operation)
						orig(c)
					}
				}
			}
		}
	}
	// Register the handlers with Gin underlying group.
	g.group.Handle(method, path, handlers...)

	return g
}

// OpenAPI returns a Gin HandlerFunc that serves
// the marshalled OpenAPI specification of the API.
func (f *Fizz) OpenAPI(info *openapi.Info, ct string) gin.HandlerFunc {
	f.gen.SetInfo(info)

	ct = strings.ToLower(ct)
	if ct == "" {
		ct = "json"
	}
	switch ct {
	case "json":
		return func(c *gin.Context) {
			c.JSON(200, f.gen.API())
		}
	case "yaml":
		return func(c *gin.Context) {
			c.YAML(200, f.gen.API())
		}
	}
	panic("invalid content type, use JSON or YAML")
}

// OperationOption represents an option-pattern function
// used to add informations to an operation.
type OperationOption func(*openapi.OperationInfo)

// StatusDescription sets the default status description of the operation.
func StatusDescription(desc string) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.StatusDescription = desc
	}
}

// Summary adds a summary to an operation.
func Summary(summary string) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Summary = summary
	}
}

// Summaryf adds a summary to an operation according
// to a format specifier.
func Summaryf(format string, a ...interface{}) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Summary = fmt.Sprintf(format, a...)
	}
}

// Description adds a description to an operation.
func Description(desc string) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Description = desc
	}
}

// Descriptionf adds a description to an operation
// according to a format specifier.
func Descriptionf(format string, a ...interface{}) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Description = fmt.Sprintf(format, a...)
	}
}

// ID overrides the operation ID.
func ID(id string) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.ID = id
	}
}

// Deprecated marks the operation as deprecated.
func Deprecated(deprecated bool) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Deprecated = deprecated
	}
}

// Response adds an additional response to the operation.
func Response(statusCode, desc string, model interface{}, headers []*openapi.ResponseHeader) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Responses = append(o.Responses, &openapi.OperationReponse{
			Code:        statusCode,
			Description: desc,
			Model:       model,
			Headers:     headers,
		})
	}
}

// Header adds a header to the operation.
func Header(name, desc string, model interface{}) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Headers = append(o.Headers, &openapi.ResponseHeader{
			Name:        name,
			Description: desc,
			Model:       model,
		})
	}
}

// InputModel overrides the binding model of the operation.
func InputModel(model interface{}) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.InputModel = model
	}
}

// OperationFromContext returns the OpenAPI operation from
// the givent Gin context or an error if none is found.
func OperationFromContext(c *gin.Context) (*openapi.Operation, error) {
	if v, ok := c.Get(ctxOpenAPIOperation); ok {
		if op, ok := v.(*openapi.Operation); ok {
			return op, nil
		}
		return nil, errors.New("invalid type: not an operation")
	}
	return nil, errors.New("operation not found")
}

func joinPaths(abs, rel string) string {
	if rel == "" {
		return abs
	}
	final := path.Join(abs, rel)
	as := lastChar(rel) == '/' && lastChar(final) != '/'
	if as {
		return final + "/"
	}
	return final
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("empty string")
	}
	return str[len(str)-1]
}

func funcEqual(f1, f2 interface{}) bool {
	v1 := reflect.ValueOf(f1)
	v2 := reflect.ValueOf(f2)

	if v1.Kind() == reflect.Func && v2.Kind() == reflect.Func { // prevent panic on call to Pointer()
		return runtime.FuncForPC(v1.Pointer()).Entry() == runtime.FuncForPC(v2.Pointer()).Entry()
	}
	return false
}
