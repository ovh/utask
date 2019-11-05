package input

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/juju/errors"
	"github.com/ovh/utask"
)

// accepted input types
const (
	InputTypeString   = "string"
	InputTypePassword = "password"
	InputTypeBool     = "bool"
	InputTypeNumber   = "number"
)

// Input represents a single input for a task
// it can express constraints on the acceptable values,
// such as a type (string by default), a regexp to be matched, an enumeration of legal values,
// wether a collection of values is accepted instead of a single value,
// and wether the input is altogether optional, which can be supported with a default value
type Input struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Regex       *string       `json:"regex,omitempty"`
	LegalValues []interface{} `json:"legal_values,omitempty"`
	Collection  bool          `json:"collection"`
	Type        string        `json:"type,omitempty"`
	Optional    bool          `json:"optional"`
	Default     interface{}   `json:"default"`
}

// Valid asserts that an input definition is valid
// - a regexp and a legal_values list are mutually exclusive
// - a regexp must compile
// - the input's type must be among the accepted types defined above
// - legal_values must match the declared type
// - default value must match the declared type
func (i Input) Valid() error {
	// check that input regex compiles
	if i.Regex != nil {
		if len(i.LegalValues) > 0 {
			return errors.NotValidf("Invalid input '%s': both regex and legal value list configured", i.Name)
		}
		if _, err := regexp.Compile(*i.Regex); err != nil {
			return errors.NotValidf("Invalid regex for input '%s'", i.Name)
		}
	}
	// check that input type is valid
	if i.Type != "" {
		switch i.Type {
		case InputTypeString, InputTypePassword, InputTypeBool, InputTypeNumber:
		default:
			return errors.NotValidf("Invalid input type '%s': must be either %v", i.Type, []string{InputTypeString, InputTypePassword, InputTypeBool, InputTypeNumber})
		}
	}
	// check that legal values match the input type
	if len(i.LegalValues) > 0 {
		for _, lv := range i.LegalValues {
			if err := i.checkValueType(lv); err != nil {
				return err
			}
		}
	}

	// check that default value matches the input type
	if err := i.checkValueType(i.Default); err != nil {
		return err
	}
	return i.CheckValue(i.Default)
}

// CheckValue verifies an input's constraints against a concrete value
func (i Input) CheckValue(val interface{}) error {
	if val != nil {
		if i.Collection {
			col, ok := val.([]interface{})
			if !ok {
				return errors.NotValidf("Input '%s' is expected to be an array", i.Name)
			}
			for _, v := range col {
				if err := i.checkSingleValue(v); err != nil {
					return err
				}
			}
		} else {
			if err := i.checkSingleValue(val); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i Input) checkSingleValue(val interface{}) error {
	// check type
	if err := i.checkValueType(val); err != nil {
		return err
	}

	// check value
	valStr := fmt.Sprintf("%v", val)
	if len(valStr) > utask.MaxTextSizeLong {
		return errors.NotValidf("Invalid input '%s': value can't be longer than %d", i.Name, utask.MaxTextSizeLong)
	}
	if len(i.LegalValues) > 0 {
		matchVal := false
		for _, legalV := range i.LegalValues {
			if legalV == val {
				matchVal = true
				break
			}
		}
		if !matchVal {
			return errors.NotValidf("Invalid input '%s': '%v' is not a legal value (%v)", i.Name, val, i.LegalValues)
		}
	} else if i.Regex != nil {
		if !regexp.MustCompile(*i.Regex).MatchString(valStr) {
			return errors.NotValidf("Invalid input '%s': '%s' doesnt comply with regex '%s'", i.Name, valStr, *i.Regex)
		}
	} else {
		if strings.Contains(valStr, `"`) {
			return errors.NotValidf("Invalid input '%s': cannot contain double quotes", i.Name)
		}
	}
	return nil
}

func (i Input) checkValueType(val interface{}) error {
	if val != nil {
		switch i.Type {
		case InputTypeString, InputTypePassword, "": // string by default
			if _, ok := val.(string); !ok {
				return errors.NotValidf("Invalid value '%s': expected a string", i.Name)
			}
		case InputTypeBool:
			if _, ok := val.(bool); !ok {
				return errors.NotValidf("Invalid value '%s': expected a boolean", i.Name)
			}
		case InputTypeNumber:
			if _, ok := val.(json.Number); !ok {
				if _, ok = val.(float64); !ok {
					return errors.NotValidf("Invalid value '%s': expected a number", i.Name)
				}
			}
		}
	}
	return nil
}
