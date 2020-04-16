package step

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/juju/errors"
	"github.com/ovh/utask/engine/values"
)

// accepted condition operators
const (
	EQ     = "EQ"
	NE     = "NE"
	GT     = "GT"
	LT     = "LT"
	GE     = "GE"
	LE     = "LE"
	REGEXP = "REGEXP"
	IN     = "IN"
	NOTIN  = "NOTIN"

	defaultSeparator = ","
)

type (
	// Assert describes a challenge to a value
	// an expected value is compared through an operator
	// the intent of this condition can be explained through a contextual message
	// for clearer error surfacing
	Assert struct {
		Value         string `json:"value"`
		Operator      string `json:"operator"`
		Expected      string `json:"expected"`
		ListSeparator string `json:"list_separator"`
		Message       string `json:"message"`
	}

	// ErrConditionNotMet is the typed error returned by Condition when its evaluation fails
	ErrConditionNotMet string
)

// Eval applies a condition on a particular item, asserting if the item meets the condition or not
func (a *Assert) Eval(v *values.Values, item interface{}, stepName string) error {
	if a != nil {
		val, err := v.Apply(a.Value, item, stepName)
		if err != nil {
			return err
		}
		expected, err := v.Apply(a.Expected, item, stepName)
		if err != nil {
			return err
		}
		valStr := strings.Replace(string(val), "<no value>", "", -1)
		expStr := strings.Replace(string(expected), "<no value>", "", -1)

		switch strings.ToUpper(a.Operator) { // normalized operator, accept both lower case and upper case from template
		case EQ:
			if valStr != expStr {
				return ErrConditionNotMet(fmt.Sprintf("Condition not met: expected '%s', got '%s': %s", expStr, valStr, a.Message))
			}
		case NE:
			if valStr == expStr {
				return ErrConditionNotMet(fmt.Sprintf("Condition not met: expected a value different from '%s': %s", expStr, a.Message))
			}
		case GT, LT, GE, LE:
			valInt, err := strconv.ParseInt(valStr, 10, 32)
			if err != nil {
				return err
			}
			expInt, err := strconv.ParseInt(expStr, 10, 32)
			if err != nil {
				return err
			}
			switch a.Operator {
			case GT:
				if valInt <= expInt {
					return ErrConditionNotMet(fmt.Sprintf("Condition not met: expected %d > %d: %s", valInt, expInt, a.Message))
				}
			case LT:
				if valInt >= expInt {
					return ErrConditionNotMet(fmt.Sprintf("Condition not met: expected %d < %d: %s", valInt, expInt, a.Message))
				}
			case GE:
				if valInt < expInt {
					return ErrConditionNotMet(fmt.Sprintf("Condition not met: expected %d >= %d: %s", valInt, expInt, a.Message))
				}
			case LE:
				if valInt > expInt {
					return ErrConditionNotMet(fmt.Sprintf("Condition not met: expected %d <= %d: %s", valInt, expInt, a.Message))
				}
			}
		case REGEXP:
			if !regexp.MustCompile(expStr).MatchString(valStr) {
				return ErrConditionNotMet(fmt.Sprintf("Condition not met: %s does not match regular expression %s", valStr, expStr))
			}
		case IN:
			if !matchList(valStr, expStr, a.ListSeparator) {
				return ErrConditionNotMet(fmt.Sprintf("Condition not met: expected %s to be found in list of acceptable values", valStr))
			}
		case NOTIN:
			if matchList(valStr, expStr, a.ListSeparator) {
				return ErrConditionNotMet(fmt.Sprintf("Condition not met: expected %s not to be found in list of unacceptable values", valStr))
			}
		}
	}
	return nil
}

// Valid asserts that a condition's definition is valid
// ie. the operator is among the accepted values listed above
func (a *Assert) Valid() error {
	if a != nil {
		switch strings.ToUpper(a.Operator) {
		case EQ, NE, GT, LT, GE, LE, IN, NOTIN:
		case REGEXP:
			if _, err := regexp.Compile(a.Expected); err != nil {
				return err
			}
		default:
			return errors.NotValidf("Unknown condition operator: %s", a.Operator)
		}
	}
	return nil
}

func matchList(valStr, expStr, sep string) bool {
	if sep == "" {
		sep = defaultSeparator
	}
	values := strings.Split(valStr, sep)
	for _, v := range values {
		if expStr == strings.TrimSpace(v) {
			return true
		}
	}
	return false
}

// Error implements standard error
func (e ErrConditionNotMet) Error() string {
	return string(e)
}
