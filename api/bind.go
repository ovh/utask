package api

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
)

// bodyBindHook is a wrapper around the default binding hook of tonic.
// It adds the possibility to bind a specific field in an object rather than
// unconditionally binding the whole object.
func bodyBindHook(c *gin.Context, v interface{}) error {
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
		if err := tonic.DefaultBindingHook(c, fv.Interface()); err != nil {
			return err
		}
		if flt.Kind() == reflect.Map {
			val.Elem().Field(i).Set(fv.Elem())
		} else {
			val.Elem().Field(i).Set(fv)
		}
	}

	return tonic.DefaultBindingHook(c, v)
}
