package taskplugin

import (
	"fmt"
	"strings"
)

// common metadata keys
const (
	HTTPStatus  = "HTTPStatus"
	HTTPHeaders = "HTTPHeaders"
	HTTPCookies = "HTTPCookies"
)

// MetadataSchemaBuilder is a helper to generate jsonschema for a metadata payload
type MetadataSchemaBuilder struct {
	properties []string
}

// NewMetadataSchema instantiates a new metadataSchemaBuilder
func NewMetadataSchema() *MetadataSchemaBuilder {
	return &MetadataSchemaBuilder{
		properties: []string{},
	}
}

// WithStatusCode adds an HTTPStatus field to metadata
func (m *MetadataSchemaBuilder) WithStatusCode() *MetadataSchemaBuilder {
	HTTPStatus := fmt.Sprintf(`"%s":{"type":"integer"}`, HTTPStatus)
	m.properties = append(m.properties, HTTPStatus)
	return m
}

// WithHeaders adds httpHeader fields to metadata
func (m *MetadataSchemaBuilder) WithHeaders(headers ...string) *MetadataSchemaBuilder {
	if len(headers) == 0 {
		return m
	}

	properties := []string{}
	for _, header := range headers {
		properties = append(properties, fmt.Sprintf(`"%s":{"type":"string"}`, header))
	}

	HTTPHeaders := fmt.Sprintf(`"%s":{"type":"object","properties":{%s}}`, HTTPHeaders, strings.Join(properties, ","))
	m.properties = append(m.properties, HTTPHeaders)
	return m
}

// String renders a json schema for metadata
func (m *MetadataSchemaBuilder) String() string {
	return fmt.Sprintf(`{"type":"object","properties":{%s}}`, strings.Join(m.properties, ","))
}
