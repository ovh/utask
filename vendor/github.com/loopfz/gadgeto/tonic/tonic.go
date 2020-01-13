package tonic

import (
	"encoding"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	validator "gopkg.in/go-playground/validator.v9"
)

// DefaultMaxBodyBytes is the maximum allowed size of a request body in bytes.
const DefaultMaxBodyBytes = 256 * 1024

// Fields tags used by tonic.
const (
	QueryTag      = "query"
	PathTag       = "path"
	HeaderTag     = "header"
	EnumTag       = "enum"
	RequiredTag   = "required"
	DefaultTag    = "default"
	ValidationTag = "validate"
	ExplodeTag    = "explode"
)

const (
	defaultMediaType    = "application/json"
	tonicRoutesInfos    = "_tonic_route_infos"
	tonicWantRouteInfos = "_tonic_want_route_infos"
)

var (
	errorHook  ErrorHook  = DefaultErrorHook
	bindHook   BindHook   = DefaultBindingHook
	renderHook RenderHook = DefaultRenderHook
	execHook   ExecHook   = DefaultExecHook

	mediaType = defaultMediaType

	routes   = make(map[string]*Route)
	routesMu = sync.Mutex{}
	funcs    = make(map[string]struct{})
	funcsMu  = sync.Mutex{}
)

// BindHook is the hook called by the wrapping gin-handler when
// binding an incoming request to the tonic-handler's input object.
type BindHook func(*gin.Context, interface{}) error

// RenderHook is the last hook called by the wrapping gin-handler
// before returning. It takes the Gin context, the HTTP status code
// and the response payload as parameters.
// Its role is to render the payload to the client to the
// proper format.
type RenderHook func(*gin.Context, int, interface{})

// ErrorHook lets you interpret errors returned by your handlers.
// After analysis, the hook should return a suitable http status code
// and and error payload.
// This lets you deeply inspect custom error types.
type ErrorHook func(*gin.Context, error) (int, interface{})

// An ExecHook is the func called to handle a request.
// The default ExecHook simply calle the wrapping gin-handler
// with the gin context.
type ExecHook func(*gin.Context, gin.HandlerFunc, string)

// DefaultErrorHook is the default error hook.
// It returns a StatusBadRequest with a payload containing
// the error message.
func DefaultErrorHook(c *gin.Context, e error) (int, interface{}) {
	return http.StatusBadRequest, gin.H{
		"error": e.Error(),
	}
}

// DefaultBindingHook is the default binding hook.
// It uses Gin JSON binding to bind the body parameters of the request
// to the input object of the handler.
// Ir teturns an error if Gin binding fails.
var DefaultBindingHook BindHook = DefaultBindingHookMaxBodyBytes(DefaultMaxBodyBytes)

// DefaultBindingHookMaxBodyBytes returns a BindHook with the default logic, with configurable MaxBodyBytes.
func DefaultBindingHookMaxBodyBytes(maxBodyBytes int64) BindHook {
	return func(c *gin.Context, i interface{}) error {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		if c.Request.ContentLength == 0 || c.Request.Method == http.MethodGet {
			return nil
		}
		if err := c.ShouldBindWith(i, binding.JSON); err != nil && err != io.EOF {
			return fmt.Errorf("error parsing request body: %s", err.Error())
		}
		return nil
	}
}

// DefaultRenderHook is the default render hook.
// It marshals the payload to JSON, or returns an empty body if the payload is nil.
// If Gin is running in debug mode, the marshalled JSON is indented.
func DefaultRenderHook(c *gin.Context, statusCode int, payload interface{}) {
	var status int
	if c.Writer.Written() {
		status = c.Writer.Status()
	} else {
		status = statusCode
	}
	if payload != nil {
		if gin.IsDebugging() {
			c.IndentedJSON(status, payload)
		} else {
			c.JSON(status, payload)
		}
	} else {
		c.String(status, "")
	}
}

// DefaultExecHook is the default exec hook.
// It simply executes the wrapping gin-handler with
// the given context.
func DefaultExecHook(c *gin.Context, h gin.HandlerFunc, fname string) {
	h(c)
}

// GetRoutes returns the routes handled by a tonic-enabled handler.
func GetRoutes() map[string]*Route {
	return routes
}

// MediaType returns the current media type (MIME)
// used by the actual render hook.
func MediaType() string {
	return defaultMediaType
}

// GetErrorHook returns the current error hook.
func GetErrorHook() ErrorHook {
	return errorHook
}

// SetErrorHook sets the given hook as the
// default error handling hook.
func SetErrorHook(eh ErrorHook) {
	if eh != nil {
		errorHook = eh
	}
}

// GetBindHook returns the current bind hook.
func GetBindHook() BindHook {
	return bindHook
}

// SetBindHook sets the given hook as the
// default binding hook.
func SetBindHook(bh BindHook) {
	if bh != nil {
		bindHook = bh
	}
}

// GetRenderHook returns the current render hook.
func GetRenderHook() RenderHook {
	return renderHook
}

// SetRenderHook sets the given hook as the default
// rendering hook. The media type is used to generate
// the OpenAPI specification.
func SetRenderHook(rh RenderHook, mt string) {
	if rh != nil {
		renderHook = rh
	}
	if mt != "" {
		mediaType = mt
	}
}

// SetExecHook sets the given hook as the
// default execution hook.
func SetExecHook(eh ExecHook) {
	if eh != nil {
		execHook = eh
	}
}

// GetExecHook returns the current execution hook.
func GetExecHook() ExecHook {
	return execHook
}

// Description set the description of a route.
func Description(s string) func(*Route) {
	return func(r *Route) {
		r.description = s
	}
}

// Summary set the summary of a route.
func Summary(s string) func(*Route) {
	return func(r *Route) {
		r.summary = s
	}
}

// Deprecated set the deprecated flag of a route.
func Deprecated(b bool) func(*Route) {
	return func(r *Route) {
		r.deprecated = b
	}
}

// BindError is an error type returned when tonic fails
// to bind parameters, to differentiate from errors returned
// by the handlers.
type BindError struct {
	validationErr error
	message       string
	typ           reflect.Type
	field         string
}

// Error implements the builtin error interface for BindError.
func (be BindError) Error() string {
	if be.field != "" && be.typ != nil {
		return fmt.Sprintf(
			"binding error on field '%s' of type '%s': %s",
			be.field,
			be.typ.Name(),
			be.message,
		)
	}
	return fmt.Sprintf("binding error: %s", be.message)
}

// ValidationErrors returns the errors from the validate process.
func (be BindError) ValidationErrors() validator.ValidationErrors {
	switch t := be.validationErr.(type) {
	case validator.ValidationErrors:
		return t
	}
	return nil
}

// An extractorFunc extracts data from a gin context according to
// parameters specified in a field tag.
type extractor func(*gin.Context, string) (string, []string, error)

// extractQuery is an extractor tgat operated on the query
// parameters of a request.
func extractQuery(c *gin.Context, tag string) (string, []string, error) {
	name, required, defaultVal, err := parseTagKey(tag)
	if err != nil {
		return "", nil, err
	}
	var params []string
	query := c.Request.URL.Query()[name]

	if c.GetBool(ExplodeTag) {
		// Delete empty elements so default and required arguments
		// will play nice together. Append to a new collection to
		// preserve order without too much copying.
		params = make([]string, 0, len(query))
		for i := range query {
			if query[i] != "" {
				params = append(params, query[i])
			}
		}
	} else {
		splitFn := func(c rune) bool {
			return c == ','
		}
		if len(query) > 1 {
			return name, nil, errors.New("repeating values not supported: use comma-separated list")
		} else if len(query) == 1 {
			params = strings.FieldsFunc(query[0], splitFn)
		}
	}

	// XXX: deprecated, use of "default" tag is preferred
	if len(params) == 0 && defaultVal != "" {
		return name, []string{defaultVal}, nil
	}
	// XXX: deprecated, use of "validate" tag is preferred
	if len(params) == 0 && required {
		return "", nil, fmt.Errorf("missing query parameter: %s", name)
	}
	return name, params, nil
}

// extractPath is an extractor that operates on the path
// parameters of a request.
func extractPath(c *gin.Context, tag string) (string, []string, error) {
	name, required, defaultVal, err := parseTagKey(tag)
	if err != nil {
		return "", nil, err
	}
	p := c.Param(name)

	// XXX: deprecated, use of "default" tag is preferred
	if p == "" && defaultVal != "" {
		return name, []string{defaultVal}, nil
	}
	// XXX: deprecated, use of "validate" tag is preferred
	if p == "" && required {
		return "", nil, fmt.Errorf("missing path parameter: %s", name)
	}

	return name, []string{p}, nil
}

// extractHeader is an extractor that operates on the headers
// of a request.
func extractHeader(c *gin.Context, tag string) (string, []string, error) {
	name, required, defaultVal, err := parseTagKey(tag)
	if err != nil {
		return "", nil, err
	}
	header := c.GetHeader(name)

	// XXX: deprecated, use of "default" tag is preferred
	if header == "" && defaultVal != "" {
		return name, []string{defaultVal}, nil
	}
	// XXX: deprecated, use of "validate" tag is preferred
	if required && header == "" {
		return "", nil, fmt.Errorf("missing header parameter: %s", name)
	}
	return name, []string{header}, nil
}

// Public signature does not expose "required" and "default" because
// they are deprecated in favor of the "validate" and "default" tags
func parseTagKey(tag string) (string, bool, string, error) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return "", false, "", fmt.Errorf("empty tag")
	}
	name, options := parts[0], parts[1:]

	var defaultVal string

	// XXX: deprecated, required + default are kept here for backwards compatibility
	// use of "default" and "validate" tags is preferred
	// Iterate through the tag options to
	// find the required key.
	var required bool
	for _, o := range options {
		o = strings.TrimSpace(o)
		if o == RequiredTag {
			required = true
		} else if strings.HasPrefix(o, fmt.Sprintf("%s=", DefaultTag)) {
			defaultVal = strings.TrimPrefix(o, fmt.Sprintf("%s=", DefaultTag))
		} else {
			return "", false, "", fmt.Errorf("malformed tag for param '%s': unknown option '%s'", name, o)
		}
	}
	return name, required, defaultVal, nil
}

// ParseTagKey parses the given struct tag key and return the
// name of the field
func ParseTagKey(tag string) (string, error) {
	s, _, _, err := parseTagKey(tag)
	return s, err
}

// bindStringValue converts and bind the value s
// to the the reflected value v.
func bindStringValue(s string, v reflect.Value) error {
	// Ensure that the reflected value is addressable
	// and wasn't obtained by the use of an unexported
	// struct field, or calling a setter will panic.
	if !v.CanSet() {
		return fmt.Errorf("unaddressable value: %v", v)
	}
	i := reflect.New(v.Type()).Interface()

	// If the value implements the encoding.TextUnmarshaler
	// interface, bind the returned string representation.
	if unmarshaler, ok := i.(encoding.TextUnmarshaler); ok {
		if err := unmarshaler.UnmarshalText([]byte(s)); err != nil {
			return err
		}
		v.Set(reflect.Indirect(reflect.ValueOf(unmarshaler)))
		return nil
	}
	// Handle time.Duration.
	if _, ok := i.(time.Duration); ok {
		d, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(d))
	}
	// Switch over the kind of the reflected value
	// and convert the string to the proper type.
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetUint(i)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Float32, reflect.Float64:
		i, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetFloat(i)
	default:
		return fmt.Errorf("unsupported parameter type: %v", v.Kind())
	}
	return nil
}
