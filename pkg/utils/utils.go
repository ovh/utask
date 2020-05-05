package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ovh/utask"

	"github.com/juju/errors"
)

// ValidString asserts that a string is within the minimum and maximum length configured for µtask
func ValidString(field, value string) error {
	if len(value) < utask.MinTextSize {
		return errors.NotValidf("%s can't be shorter than %d characters", field, utask.MinTextSize)
	}
	if len(value) > utask.MaxTextSize {
		return errors.NotValidf("%s can't be longer than %d characters", field, utask.MaxTextSize)
	}
	return nil
}

// ValidText asserts that a long text string is within the minimum and maximum length configured for µtask
func ValidText(field, value string) error {
	if len(value) < utask.MinTextSize {
		return errors.NotValidf("%s can't be shorter than %d characters", field, utask.MinTextSize)
	}
	if len(value) > utask.MaxTextSizeLong {
		return errors.NotValidf("%s can't be longer than %d characters", field, utask.MaxTextSizeLong)
	}
	return nil
}

// NormalizeName trims leading and trailing spaces on a string, and converts its characters to lowercase
func NormalizeName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ListContainsString asserts that a string slice contains a given string
func ListContainsString(list []string, item string) bool {
	if list != nil {
		for _, i := range list {
			if i == item {
				return true
			}
		}
	}
	return false
}

// PrintJSON prints out an interface{} as an indented json string
func PrintJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "   ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
}

// StrPtr returns the pointer to a string's value
func StrPtr(s string) *string {
	return &s
}

// JSONnumberUnmarshal unmarshals a json string with numbers cast as json.Number, not float64 (to avoid scientific notation on large IDs)
func JSONnumberUnmarshal(r io.Reader, i interface{}) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec.Decode(i)
}

// JSONMarshal will JSON encode a given object, without escaping HTML characters
func JSONMarshal(obj interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}

	// json.NewEncoder.Encode adds a final '\n', json.Marshal does not.
	// Let's keep the default json.Marshal behaviour.
	res := b.Bytes()
	if len(res) >= 1 && res[len(res)-1] == '\n' {
		res = res[:len(res)-1]
	}
	return res, nil
}

// JSONMarshalIndent will JSON encode a given object, without escaping HTML characters and indentation
func JSONMarshalIndent(obj interface{}, prefix, indent string) ([]byte, error) {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	enc.SetIndent(prefix, indent)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}

	// json.NewEncoder.Encode adds a final '\n', json.Marshal does not.
	// Let's keep the default json.Marshal behaviour.
	res := b.Bytes()
	if len(res) >= 1 && res[len(res)-1] == '\n' {
		res = res[:len(res)-1]
	}
	return res, nil
}

// ConvertJSONRowToSlice takes a json-formatted array and returns a string slice
func ConvertJSONRowToSlice(in string) ([]string, error) {
	var tmpslice []string
	err := json.Unmarshal([]byte(in), &tmpslice)
	return tmpslice, err
}

// JSONUseNumber returns a json decoder to use numbers while decoding json
func JSONUseNumber(d *json.Decoder) *json.Decoder {
	d.UseNumber()
	return d
}

// HasDupsArray returns a boolean indicating if array contains duplicates
func HasDupsArray(elements []string) bool {
	encountered := map[string]bool{}
	for v := range elements {
		encountered[elements[v]] = true
	}

	return len(elements) != len(encountered)
}
