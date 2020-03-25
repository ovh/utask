package jsonschema

import (
	"bytes"
	"encoding/json"

	"github.com/juju/errors"
	"github.com/santhosh-tekuri/jsonschema"

	"github.com/ovh/utask/pkg/utils"
)

const (
	// RootKey is a reference point for building json schema, used internally
	RootKey = "utaskRootKey"
	draft7  = "http://json-schema.org/draft-07/schema#"
	draft6  = "http://json-schema.org/draft-06/schema#"
	draft4  = "http://json-schema.org/draft-04/schema#"
)

// ValidateFunc is jsonschema validator
type ValidateFunc func(interface{}) error

// Validator generates a ValidateFunc from a json Schema definition
func Validator(url string, rawSchema json.RawMessage) ValidateFunc {
	schema, err := compile(url, rawSchema)
	if schema == nil || err != nil {
		return nil
	}
	return schema.ValidateInterface
}

// NormalizeAndCompile normalizes the version and then compile the json schema.
func NormalizeAndCompile(url string, s json.RawMessage) (json.RawMessage, error) {
	s = computeVersion(s)
	_, err := compile(url, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func compile(url string, s json.RawMessage) (*jsonschema.Schema, error) {
	if len(s) == 0 {
		return nil, nil
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(url, bytes.NewReader(s)); err != nil {
		return nil, err
	}

	schema, err := compiler.Compile(url)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

// computeVersion compute the json schema version following theses rules:
//      - "$schema" is used if present
//      - "version" is used if "$schema" is absent
//      - if "version" is absent, fallback to draft7 (latest as time of writing)
func computeVersion(rawSchema json.RawMessage) json.RawMessage {
	var m map[string]interface{}
	if err := json.Unmarshal(rawSchema, &m); err != nil {
		return rawSchema
	}

	if _, ok := m["$schema"]; ok {
		return rawSchema
	}

	var version float64
	if v, ok := m["version"]; ok {
		version, _ = v.(float64)
		delete(m, "version")
	}

	switch version {
	case 6:
		m["$schema"] = draft6
	case 4:
		m["$schema"] = draft4
	case 7:
		fallthrough
	default:
		m["$schema"] = draft7
	}

	newSchema, err := utils.JSONMarshal(&m)
	if err != nil {
		return rawSchema
	}
	return newSchema
}

// ExtractProperty extract possible variables from a json schema.
func ExtractProperty(url string, s json.RawMessage) (map[string][]string, error) {
	schema, err := compile(url, s)
	if err != nil {
		return nil, nil
	}

	if schema == nil {
		return nil, nil
	}

	extracted := extractSchemas(schema, make(map[string][]*jsonschema.Schema))
	output := make(map[string][]string)

	for name, schemaList := range extracted {
		output[name] = extractPropertiesFromSchema(schemaList...)
	}

	properties := extractProperties(schema, output, RootKey)

	if _, ok := properties[RootKey]; ok {
		return nil, errors.Errorf("Cannot use property %q as it's a reserved keyword. Please rename your property", RootKey)
	}

	// add root properties to a special key to avoid exposing other properties
	for name := range schema.Properties {
		properties[RootKey] = append(properties[RootKey], name)
	}

	return properties, nil
}

// extractProperties extract nested properties from schema and merge them with properties map.
func extractProperties(schema *jsonschema.Schema, properties map[string][]string, key string) map[string][]string {
	for name, s := range schema.Properties {
		// ignore recursion beginning
		if key != RootKey {
			properties[key] = append(properties[key], name)
		}
		if _, ok := properties[name]; !ok {
			properties[name] = make([]string, 0)
		}
		mergeProperties(properties, extractProperties(s, properties, name))
	}
	return properties
}

// mergeProperties merges "b" values in "a".
func mergeProperties(a, b map[string][]string) map[string][]string {
	for k := range a {
		for _, v2 := range b {
			_, ok := a[k]
			_, ok2 := b[k]
			if ok && ok2 {
				a[k] = appendWithoutDuplicate(a[k], b[k])
			}
			if !ok {
				a[k] = append(a[k], v2...)
			}
		}
	}
	return a
}

func appendWithoutDuplicate(a, b []string) []string {
	for _, v2 := range b {
		if !utils.ListContainsString(a, v2) {
			a = append(a, v2)
		}
	}
	return a
}

func extractPropertiesFromSchema(schemas ...*jsonschema.Schema) []string {
	result := make([]string, 0)

	for _, schema := range schemas {

		if schema.Ref != nil {
			schema = schema.Ref
		}

		for name := range schema.Properties {
			result = append(result, name)
		}
	}

	return result
}

// extractSchemas returns each possible properties with their associated schemas (ref, allOf...)
func extractSchemas(s *jsonschema.Schema, properties map[string][]*jsonschema.Schema) map[string][]*jsonschema.Schema {
	// If not nil, all remainings fields are ignored, so we jump directly to the
	// definitions of the schema.
	// see https://godoc.org/github.com/santhosh-tekuri/jsonschema#Schema
	if s.Ref != nil {
		return extractSchemas(s.Ref, properties)
	}

	length := len(properties)

	// Loop over actual properties first
	for name, schema := range s.Properties {
		if _, ok := properties[name]; !ok {
			properties[name] = extractSchemasFromSchema(schema)
		}
	}

	if len(properties) == length {
		return properties
	}

	// In depth exploration
	for _, prop := range s.Properties {
		if prop.Ref != nil {
			properties = extractSchemas(prop.Ref, properties)
		}
		if prop.Not != nil {
			properties = extractSchemas(prop.Not, properties)
		}
		if prop.If != nil {
			properties = extractSchemas(prop.If, properties)
		}
		if prop.Then != nil {
			properties = extractSchemas(prop.Then, properties)
		}
		if prop.Else != nil {
			properties = extractSchemas(prop.Else, properties)
		}
		if len(prop.AnyOf) > 0 {
			for _, any := range prop.AnyOf {
				properties = extractSchemas(any, properties)
			}
		}
		if len(prop.AllOf) > 0 {
			for _, allof := range prop.AllOf {
				properties = extractSchemas(allof, properties)
			}
		}
		if len(prop.OneOf) > 0 {
			for _, oneof := range prop.OneOf {
				properties = extractSchemas(oneof, properties)
			}
		}
	}

	return properties
}

func extractSchemasFromSchema(s *jsonschema.Schema) []*jsonschema.Schema {
	schemas := make([]*jsonschema.Schema, 0)
	schemas = appendIfNotNil(schemas, s.Ref, s.Not, s.If, s.Then, s.Else)
	schemas = appendIfNotNil(schemas, s.AllOf...)
	schemas = appendIfNotNil(schemas, s.AnyOf...)
	return appendIfNotNil(schemas, s.OneOf...)
}

func appendIfNotNil(schemas []*jsonschema.Schema, s ...*jsonschema.Schema) []*jsonschema.Schema {
	for _, schema := range s {
		if schema != nil {
			schemas = append(schemas, schema)
		}
	}
	return schemas
}
