package db

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/ovh/utask/engine/input"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/pkg/utils"

	"github.com/go-gorp/gorp"
)

type typeConverter struct{}

func (tc typeConverter) ToDb(val interface{}) (interface{}, error) {
	switch t := val.(type) {
	case []string, map[string]*step.Step, map[string]string, map[string]interface{}, []input.Input, []values.Variable, map[string]json.RawMessage:
		b, err := utils.JSONMarshal(t)
		if err != nil {
			return nil, err
		}
		return string(b), nil
	}
	return val, nil
}

func (tc typeConverter) FromDb(target interface{}) (gorp.CustomScanner, bool) {
	switch target.(type) {
	case *[]string, *map[string]*step.Step, *map[string]string, *map[string]interface{}, *[]input.Input, *[]values.Variable, *map[string]json.RawMessage:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New("FromDb: Unable to convert []string to *string")
			}
			return utils.JSONnumberUnmarshal(strings.NewReader(*s), target)
		}
		return gorp.CustomScanner{
			Holder: new(string),
			Target: target,
			Binder: binder,
		}, true
	}
	return gorp.CustomScanner{}, false
}
