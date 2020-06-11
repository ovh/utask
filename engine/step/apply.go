package step

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/pkg/utils"
)

var (
	errNotTemplatable = errors.New("value given is not templatable")
)

func resolveObject(val *values.Values, objson json.RawMessage, item interface{}, stepName string) ([]byte, error) {
	obj, err := rawResolveObject(val, objson, item, stepName)
	if err != nil {
		return nil, err
	}
	return utils.JSONMarshal(obj)
}

func rawResolveObject(val *values.Values, objson json.RawMessage, item interface{}, stepName string) (interface{}, error) {
	var obj interface{}
	if err := utils.JSONnumberUnmarshal(bytes.NewBuffer(objson), &obj); err != nil {
		return nil, err
	}
	v := reflect.ValueOf(obj)
	if err := apply(val, v, item, stepName); err != nil {
		return nil, err
	}
	return obj, nil
}

func apply(val *values.Values, v reflect.Value, item interface{}, stepName string) error {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		return applyMap(val, v, item, stepName)
	case reflect.Slice, reflect.Array:
		return applySlice(val, v, item, stepName)
	}

	return nil
}

func applyMap(val *values.Values, v reflect.Value, item interface{}, stepName string) error {
	iter := v.MapRange()

	for iter.Next() {
		mv := iter.Value()
		for mv.Kind() == reflect.Interface || mv.Kind() == reflect.Ptr {
			mv = mv.Elem()
		}

		switch mv.Kind() {
		case reflect.Map:
			applyMap(val, mv, item, stepName)
		case reflect.Slice:
			applySlice(val, mv, item, stepName)
		case reflect.String:
			newValue, err := applyString(val, mv, item, stepName)
			switch err {
			case nil:
				v.SetMapIndex(iter.Key(), reflect.ValueOf(newValue))
			case errNotTemplatable:
				// current value Kind is string, but actual type could not be string (e.g. json.Number)
				// in that case, we should keep the actual value as it will never be templated
			default:
				return err
			}
		}
	}

	return nil
}

func applySlice(val *values.Values, v reflect.Value, item interface{}, stepName string) error {
	for i := 0; i < v.Len(); i++ {
		mv := v.Index(i)

		elem := mv
		for elem.Kind() == reflect.Interface || elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		switch elem.Kind() {
		case reflect.Map:
			applyMap(val, elem, item, stepName)
		case reflect.Slice:
			applySlice(val, elem, item, stepName)
		case reflect.String:
			newValue, err := applyString(val, elem, item, stepName)
			switch err {
			case nil:
				mv.Set(reflect.ValueOf(newValue))
			case errNotTemplatable:
				// current value Kind is string, but actual type could not be string (e.g. json.Number)
				// in that case, we should keep the actual value as it will never be templated
			default:
				return err
			}
		}
	}

	return nil
}

func applyString(val *values.Values, v reflect.Value, item interface{}, stepName string) (string, error) {
	strval := v.Interface()
	str, ok := strval.(string)
	if !ok {
		return "", errNotTemplatable
	}
	resolved, err := val.Apply(str, item, stepName)
	if err != nil {
		return "", err
	}
	return string(resolved), nil
}
