package step

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/ovh/utask/engine/values"
)

func resolveObject(val *values.Values, objson json.RawMessage, item interface{}, stepName string) ([]byte, error) {
	obj, err := rawResolveObject(val, objson, item, stepName)
	if err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

func rawResolveObject(val *values.Values, objson json.RawMessage, item interface{}, stepName string) (interface{}, error) {
	dec := json.NewDecoder(bytes.NewBuffer(objson))
	dec.UseNumber()
	var obj interface{}
	if err := dec.Decode(&obj); err != nil {
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
			if err != nil {
				return err
			}

			if newValue != "" {
				v.SetMapIndex(iter.Key(), reflect.ValueOf(newValue))
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
			if err != nil {
				return err
			}

			if newValue != "" {
				mv.Set(reflect.ValueOf(newValue))
			}
		}
	}

	return nil
}

func applyString(val *values.Values, v reflect.Value, item interface{}, stepName string) (string, error) {
	strval := v.Interface()
	str, ok := strval.(string)
	if !ok {
		return "", nil
	}
	resolved, err := val.Apply(str, item, stepName)
	if err != nil {
		return "", err
	}
	return string(resolved), nil
}
