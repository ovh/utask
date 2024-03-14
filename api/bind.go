package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/yaml"
)

const (
	// default max body bytes: 256KB
	// this can be overridden via configuration
	defaultMaxBodyBytes = 256 * 1024

	// absolute upper limit for configuration max body bytes: 10MB
	upperLimitMaxBodyBytes = 10 * 1024 * 1024

	// absolute lower limit for configuration max body bytes: 1KB
	lowerLimitMaxBodyBytes = 1024
)

var yamlBind = yamlBinding{}

type yamlBinding struct{}

func (yamlBinding) Name() string { return "yamlBinding" }
func (yamlBinding) Bind(req *http.Request, obj interface{}) error {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	return yaml.Unmarshal(bodyBytes, obj, jsonNumberOpt)
}

// defaultBindingHook is a wrapper around the yaml binding.
// It adds the possibility to bind a specific field in an object rather than
// unconditionally binding the whole object.
func defaultBindingHook(maxBodyBytes int64) func(*gin.Context, interface{}) error {
	if maxBodyBytes == 0 {
		maxBodyBytes = defaultMaxBodyBytes
	} else if maxBodyBytes > upperLimitMaxBodyBytes {
		maxBodyBytes = upperLimitMaxBodyBytes
	} else if maxBodyBytes < lowerLimitMaxBodyBytes {
		maxBodyBytes = lowerLimitMaxBodyBytes
	}

	return func(c *gin.Context, v interface{}) error {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		if c.Request.ContentLength == 0 || c.Request.Method == http.MethodGet {
			return nil
		}

		val := reflect.ValueOf(v)
		typ := reflect.TypeOf(v).Elem()

		for i := 0; i < typ.NumField(); i++ {
			ft := typ.Field(i)
			if _, ok := ft.Tag.Lookup("body"); !ok {
				continue
			}
			flt := ft.Type
			var fv reflect.Value
			if flt.Kind() == reflect.Map {
				fv = reflect.New(flt)
			} else {
				fv = reflect.New(flt.Elem())
			}
			if err := c.ShouldBindWith(fv.Interface(), yamlBind); err != nil && err != io.EOF {
				return fmt.Errorf("error parsing request body: %s", err.Error())
			}
			if flt.Kind() == reflect.Map {
				val.Elem().Field(i).Set(fv.Elem())
			} else {
				val.Elem().Field(i).Set(fv)
			}
		}

		if err := c.ShouldBindWith(v, yamlBind); err != nil && err != io.EOF {
			return fmt.Errorf("error parsing request body: %s", err.Error())
		}
		return nil
	}
}

func jsonNumberOpt(dec *json.Decoder) *json.Decoder {
	dec.UseNumber()
	return dec
}
