package tonic

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	validator "gopkg.in/go-playground/validator.v9"
)

var (
	validatorObj  *validator.Validate
	validatorOnce sync.Once
)

// Handler returns a Gin HandlerFunc that wraps the handler passed
// in parameters.
// The handler may use the following signature:
//
//  func(*gin.Context, [input object ptr]) ([output object], error)
//
// Input and output objects are both optional.
// As such, the minimal accepted signature is:
//
//  func(*gin.Context) error
//
// The wrapping gin-handler will bind the parameters from the query-string,
// path, body and headers, and handle the errors.
//
// Handler will panic if the tonic handler or its input/output values
// are of incompatible type.
func Handler(h interface{}, status int, options ...func(*Route)) gin.HandlerFunc {
	hv := reflect.ValueOf(h)

	if hv.Kind() != reflect.Func {
		panic(fmt.Sprintf("handler parameters must be a function, got %T", h))
	}
	ht := hv.Type()
	fname := fmt.Sprintf("%s_%s", runtime.FuncForPC(hv.Pointer()).Name(), uuid.Must(uuid.NewRandom()).String())

	in := input(ht, fname)
	out := output(ht, fname)

	// Wrap Gin handler.
	f := func(c *gin.Context) {
		_, ok := c.Get(tonicWantRouteInfos)
		if ok {
			r := &Route{}
			r.defaultStatusCode = status
			r.handler = hv
			r.handlerType = ht
			r.inputType = in
			r.outputType = out
			for _, opt := range options {
				opt(r)
			}
			c.Set(tonicRoutesInfos, r)
			c.Abort()
			return
		}
		// funcIn contains the input parameters of the
		// tonic handler call.
		args := []reflect.Value{reflect.ValueOf(c)}

		// Tonic handler has custom input, handle
		// binding.
		if in != nil {
			input := reflect.New(in)
			// Bind the body with the hook.
			if err := bindHook(c, input.Interface()); err != nil {
				handleError(c, BindError{message: err.Error(), typ: in})
				return
			}
			// Bind query-parameters.
			if err := bind(c, input, QueryTag, extractQuery); err != nil {
				handleError(c, err)
				return
			}
			// Bind path arguments.
			if err := bind(c, input, PathTag, extractPath); err != nil {
				handleError(c, err)
				return
			}
			// Bind headers.
			if err := bind(c, input, HeaderTag, extractHeader); err != nil {
				handleError(c, err)
				return
			}
			// validating query and path inputs if they have a validate tag
			initValidator()
			args = append(args, input)
			if err := validatorObj.Struct(input.Interface()); err != nil {
				handleError(c, BindError{message: err.Error(), validationErr: err})
				return
			}
		}
		// Call tonic handler with the arguments
		// and extract the returned values.
		var err, val interface{}

		ret := hv.Call(args)
		if out != nil {
			val = ret[0].Interface()
			err = ret[1].Interface()
		} else {
			err = ret[0].Interface()
		}
		// Handle the error returned by the
		// handler invocation, if any.
		if err != nil {
			handleError(c, err.(error))
			return
		}
		renderHook(c, status, val)
	}
	// Register route in tonic-enabled routes map
	route := &Route{
		defaultStatusCode: status,
		handler:           hv,
		handlerType:       ht,
		inputType:         in,
		outputType:        out,
	}
	for _, opt := range options {
		opt(route)
	}
	routes[fname] = route

	ret := func(c *gin.Context) { execHook(c, f, fname) }

	funcsMu.Lock()
	defer funcsMu.Unlock()
	funcs[runtime.FuncForPC(reflect.ValueOf(ret).Pointer()).Name()] = struct{}{}

	return ret
}

// RegisterValidation registers a custom validation on the validator.Validate instance of the package
// NOTE: calling this function may instantiate the validator itself.
// NOTE: this function is not thread safe, since the validator validation registration isn't
func RegisterValidation(tagName string, validationFunc validator.Func) error {
	initValidator()
	return validatorObj.RegisterValidation(tagName, validationFunc)
}

func initValidator() {
	validatorOnce.Do(func() {
		validatorObj = validator.New()
		validatorObj.SetTagName(ValidationTag)
	})
}

// bind binds the fields the fields of the input object in with
// the values of the parameters extracted from the Gin context.
// It reads tag to know what to extract using the extractor func.
func bind(c *gin.Context, v reflect.Value, tag string, extract extractor) error {
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		field := v.Field(i)

		// Handle embedded fields with a recursive call.
		// If the field is a pointer, but is nil, we
		// create a new value of the same type, or we
		// take the existing memory address.
		if ft.Anonymous {
			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					field.Set(reflect.New(field.Type().Elem()))
				}
			} else {
				if field.CanAddr() {
					field = field.Addr()
				}
			}
			err := bind(c, field, tag, extract)
			if err != nil {
				return err
			}
			continue
		}
		tagValue := ft.Tag.Get(tag)
		if tagValue == "" {
			continue
		}
		// Set-up context for extractors.
		// Query.
		c.Set(ExplodeTag, true) // default
		if explodeVal, ok := ft.Tag.Lookup(ExplodeTag); ok {
			if explode, err := strconv.ParseBool(explodeVal); err == nil && !explode {
				c.Set(ExplodeTag, false)
			}
		}
		_, fieldValues, err := extract(c, tagValue)
		if err != nil {
			return BindError{field: ft.Name, typ: t, message: err.Error()}
		}
		// Extract default value and use it in place
		// if no values were returned.
		def, ok := ft.Tag.Lookup(DefaultTag)
		if ok && len(fieldValues) == 0 {
			if c.GetBool(ExplodeTag) {
				fieldValues = append(fieldValues, strings.Split(def, ",")...)
			} else {
				fieldValues = append(fieldValues, def)
			}
		}
		if len(fieldValues) == 0 {
			continue
		}
		// If the field is a nil pointer to a concrete type,
		// create a new addressable value for this type.
		if field.Kind() == reflect.Ptr && field.IsNil() {
			f := reflect.New(field.Type().Elem())
			field.Set(f)
		}
		// Dereference pointer.
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}
		kind := field.Kind()

		// Multiple values can only be filled to types
		// Slice and Array.
		if len(fieldValues) > 1 && (kind != reflect.Slice && kind != reflect.Array) {
			return BindError{field: ft.Name, typ: t, message: "multiple values not supported"}
		}
		// Ensure that the number of values to fill does
		// not exceed the length of a field of type Array.
		if kind == reflect.Array {
			if field.Len() != len(fieldValues) {
				return BindError{field: ft.Name, typ: t, message: fmt.Sprintf(
					"parameter expect %d values, got %d", field.Len(), len(fieldValues)),
				}
			}
		}
		if kind == reflect.Slice || kind == reflect.Array {
			// Create a new slice with an adequate
			// length to set all the values.
			if kind == reflect.Slice {
				field.Set(reflect.MakeSlice(field.Type(), 0, len(fieldValues)))
			}
			for i, val := range fieldValues {
				v := reflect.New(field.Type().Elem()).Elem()
				err = bindStringValue(val, v)
				if err != nil {
					return BindError{field: ft.Name, typ: t, message: err.Error()}
				}
				if kind == reflect.Slice {
					field.Set(reflect.Append(field, v))
				}
				if kind == reflect.Array {
					field.Index(i).Set(v)
				}
			}
			continue
		}
		// Handle enum values.
		enum := ft.Tag.Get(EnumTag)
		if enum != "" {
			enumValues := strings.Split(strings.TrimSpace(enum), ",")
			if len(enumValues) != 0 {
				if !contains(enumValues, fieldValues[0]) {
					return BindError{field: ft.Name, typ: t, message: fmt.Sprintf(
						"parameter has not an acceptable value, %s=%v", EnumTag, enumValues),
					}
				}
			}
		}
		// Fill string value into input field.
		err = bindStringValue(fieldValues[0], field)
		if err != nil {
			return BindError{field: ft.Name, typ: t, message: err.Error()}
		}
	}
	return nil
}

// input checks the input parameters of a tonic handler
// and return the type of the second parameter, if any.
func input(ht reflect.Type, name string) reflect.Type {
	n := ht.NumIn()
	if n < 1 || n > 2 {
		panic(fmt.Sprintf(
			"incorrect number of input parameters for handler %s, expected 1 or 2, got %d",
			name, n,
		))
	}
	// First parameter of tonic handler must be
	// a pointer to a Gin context.
	if !ht.In(0).ConvertibleTo(reflect.TypeOf(&gin.Context{})) {
		panic(fmt.Sprintf(
			"invalid first parameter for handler %s, expected *gin.Context, got %v",
			name, ht.In(0),
		))
	}
	if n == 2 {
		// Check the type of the second parameter
		// of the handler. Must be a pointer to a struct.
		if ht.In(1).Kind() != reflect.Ptr || ht.In(1).Elem().Kind() != reflect.Struct {
			panic(fmt.Sprintf(
				"invalid second parameter for handler %s, expected pointer to struct, got %v",
				name, ht.In(1),
			))
		} else {
			return ht.In(1).Elem()
		}
	}
	return nil
}

// output checks the output parameters of a tonic handler
// and return the type of the return type, if any.
func output(ht reflect.Type, name string) reflect.Type {
	n := ht.NumOut()

	if n < 1 || n > 2 {
		panic(fmt.Sprintf(
			"incorrect number of output parameters for handler %s, expected 1 or 2, got %d",
			name, n,
		))
	}
	// Check the type of the error parameter, which
	// should always come last.
	if !ht.Out(n - 1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		panic(fmt.Sprintf(
			"unsupported type for handler %s output parameter: expected error interface, got %v",
			name, ht.Out(n-1),
		))
	}
	if n == 2 {
		t := ht.Out(0)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return t
	}
	return nil
}

// handleError handles any error raised during the execution
// of the wrapping gin-handler.
func handleError(c *gin.Context, err error) {
	if len(c.Errors) == 0 {
		c.Error(err)
	}
	code, resp := errorHook(c, err)
	renderHook(c, code, resp)
}

// contains returns whether in contain s.
func contains(in []string, s string) bool {
	for _, v := range in {
		if v == s {
			return true
		}
	}
	return false
}
