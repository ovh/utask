package openapi

// OpenAPI represents the root document object of
// an OpenAPI document.
type OpenAPI struct {
	OpenAPI    string      `json:"openapi" yaml:"openapi"`
	Info       *Info       `json:"info" yaml:"info"`
	Servers    []*Server   `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths      Paths       `json:"paths" yaml:"paths"`
	Components *Components `json:"components,omitempty" yaml:"components,omitempty"`
	Tags       []*Tag      `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Components holds a set of reusable objects for different
// ascpects of the specification.
type Components struct {
	Schemas    map[string]*SchemaOrRef    `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses  map[string]*ResponseOrRef  `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters map[string]*ParameterOrRef `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples   map[string]*ExampleOrRef   `json:"examples,omitempty" yaml:"examples,omitempty"`
	Headers    map[string]*HeaderOrRef    `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// Info represents the metadata of an API.
type Info struct {
	Title          string   `json:"title" yaml:"title"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string   `json:"version" yaml:"version"`
}

// Contact represents the the contact informations
// exposed for an API.
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// License represents the license informations
// exposed for an API.
type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Server represents a server.
type Server struct {
	URL         string                     `json:"url" yaml:"url"`
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]*ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// ServerVariable represents a server variable for server
// URL template substitution.
type ServerVariable struct {
	Ennum       []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// Paths represents the relative paths to the individual
// endpoints and their operations.
type Paths map[string]*PathItem

// PathItem describes the operations available on a single
// API path.
type PathItem struct {
	Ref         string            `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string            `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	GET         *Operation        `json:"get,omitempty" yaml:"get,omitempty"`
	PUT         *Operation        `json:"put,omitempty" yaml:"put,omitempty"`
	POST        *Operation        `json:"post,omitempty" yaml:"post,omitempty"`
	DELETE      *Operation        `json:"delete,omitempty" yaml:"delete,omitempty"`
	OPTIONS     *Operation        `json:"options,omitempty" yaml:"options,omitempty"`
	HEAD        *Operation        `json:"head,omitempty" yaml:"head,omitempty"`
	PATCH       *Operation        `json:"patch,omitempty" yaml:"patch,omitempty"`
	TRACE       *Operation        `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers     []*Server         `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters  []*ParameterOrRef `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Reference is a simple object to allow referencing
// other components in the specification, internally and
// externally.
type Reference struct {
	Ref string `json:"$ref" yaml:"$ref"`
}

// Parameter describes a single operation parameter.
type Parameter struct {
	Name            string       `json:"name" yaml:"name"`
	In              string       `json:"in" yaml:"in"`
	Description     string       `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool         `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool         `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool         `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Schema          *SchemaOrRef `json:"schema,omitempty" yaml:"schema,omitempty"`
	Style           string       `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool         `json:"explode,omitempty" yaml:"explode,omitempty"`
}

// ParameterOrRef represents a Parameter that can be inlined
// or referenced in the API description.
type ParameterOrRef struct {
	*Parameter
	*Reference
}

// MarshalYAML implements yaml.Marshaler for ParameterOrRef.
func (por *ParameterOrRef) MarshalYAML() (interface{}, error) {
	if por.Parameter != nil {
		return por.Parameter, nil
	}
	return por.Reference, nil
}

// RequestBody represents a request body.
type RequestBody struct {
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]*MediaType `json:"content" yaml:"content"`
	Required    bool                  `json:"required,omitempty" yaml:"required,omitempty"`
}

// SchemaOrRef represents a Schema that can be inlined
// or referenced in the API description.
type SchemaOrRef struct {
	*Schema
	*Reference
}

// MarshalYAML implements yaml.Marshaler for SchemaOrRef.
func (sor *SchemaOrRef) MarshalYAML() (interface{}, error) {
	if sor.Schema != nil {
		return sor.Schema, nil
	}
	return sor.Reference, nil
}

// Schema represents the definition of input and output data
// types of the API.
type Schema struct {
	// The following properties are taken from the JSON Schema
	// definition but their definitions were adjusted to the
	// OpenAPI Specification.
	Type                 string                  `json:"type,omitempty" yaml:"type,omitempty"`
	AllOf                *SchemaOrRef            `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                *SchemaOrRef            `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                *SchemaOrRef            `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Items                *SchemaOrRef            `json:"items,omitempty" yaml:"items,omitempty"`
	Properties           map[string]*SchemaOrRef `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties *SchemaOrRef            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Description          string                  `json:"description,omitempty" yaml:"description,omitempty"`
	Format               string                  `json:"format,omitempty" yaml:"format,omitempty"`
	Default              interface{}             `json:"default,omitempty" yaml:"default,omitempty"`

	// The following properties are taken directly from the
	// JSON Schema definition and follow the same specifications
	Title            string        `json:"title,omitempty" yaml:"title,omitempty"`
	MultipleOf       int           `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum          int           `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum          int           `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength        int           `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength        int           `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern          string        `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MaxItems         int           `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems         int           `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems      bool          `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxProperties    int           `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties    int           `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Required         []string      `json:"required,omitempty" yaml:"required,omitempty"`
	Enum             []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
	Nullable         bool          `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Deprecated       bool          `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

// Operation describes an API operation on a path.
type Operation struct {
	Tags        []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string            `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	ID          string            `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  []*ParameterOrRef `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody      `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   Responses         `json:"responses,omitempty" yaml:"responses,omitempty"`
	Deprecated  bool              `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Servers     []*Server         `json:"servers,omitempty" yaml:"servers,omitempty"`
}

// Responses represents a container for the expected responses
// of an opration. It maps a HTTP response code to the expected
// response.
type Responses map[string]*ResponseOrRef

// ResponseOrRef represents a Response that can be inlined
// or referenced in the API description.
type ResponseOrRef struct {
	*Response
	*Reference
}

// MarshalYAML implements yaml.Marshaler for ResponseOrRef.
func (ror *ResponseOrRef) MarshalYAML() (interface{}, error) {
	if ror.Response != nil {
		return ror.Response, nil
	}
	return ror.Reference, nil
}

// Response describes a single response from an API.
type Response struct {
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Headers     map[string]*HeaderOrRef    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]*MediaTypeOrRef `json:"content,omitempty" yaml:"content,omitempty"`
}

// HeaderOrRef represents a Header that can be inlined
// or referenced in the API description.
type HeaderOrRef struct {
	*Header
	*Reference
}

// MarshalYAML implements yaml.Marshaler for HeaderOrRef.
func (hor *HeaderOrRef) MarshalYAML() (interface{}, error) {
	if hor.Header != nil {
		return hor.Header, nil
	}
	return hor.Reference, nil
}

// Header represents an HTTP header.
type Header struct {
	Description     string       `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool         `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool         `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool         `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Schema          *SchemaOrRef `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// MediaTypeOrRef represents a MediaType that can be inlined
// or referenced in the API description.
type MediaTypeOrRef struct {
	*MediaType
	*Reference
}

// MarshalYAML implements yaml.Marshaler for MediaTypeOrRef.
func (mtor *MediaTypeOrRef) MarshalYAML() (interface{}, error) {
	if mtor.MediaType != nil {
		return mtor.MediaType, nil
	}
	return mtor.Reference, nil
}

// MediaType represents the type of a media.
type MediaType struct {
	Schema   *SchemaOrRef             `json:"schema" yaml:"schema"`
	Example  interface{}              `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]*ExampleOrRef `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]*Encoding     `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// ExampleOrRef represents an Example that can be inlined
// or referenced in the API description.
type ExampleOrRef struct {
	*Example
	*Reference
}

// MarshalYAML implements yaml.Marshaler for ExampleOrRef.
func (eor *ExampleOrRef) MarshalYAML() (interface{}, error) {
	if eor.Example != nil {
		return eor.Example, nil
	}
	return eor.Reference, nil
}

// Example represents the exanple of a media type.
type Example struct {
	Summary       string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"`
	Value         interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Encoding represents a single encoding definition
// applied to a single schema property.
type Encoding struct {
	ContentType   string                  `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]*HeaderOrRef `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string                  `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       bool                    `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool                    `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

// Tag represents the metadata of a single tag.
type Tag struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}
