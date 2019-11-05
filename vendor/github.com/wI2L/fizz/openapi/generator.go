package openapi

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/loopfz/gadgeto/tonic"
)

const (
	version        = "3.0.1"
	anyMediaType   = "*/*"
	formatTag      = "format"
	deprecatedTag  = "deprecated"
	descriptionTag = "description"
)

var (
	paramsInPathRe = regexp.MustCompile(`\{(.*?)\}`)
	ginPathParamRe = regexp.MustCompile(`\/:([^\/]*)`)
)

// mediaTags maps media types to well-known
// struct tags used for marshaling.
var mediaTags = map[string]string{
	"application/json": "json",
	"application/xml":  "xml",
}

// Generator is an OpenAPI 3 generator.
type Generator struct {
	api           *OpenAPI
	config        *SpecGenConfig
	schemaTypes   map[reflect.Type]struct{}
	typeNames     map[reflect.Type]string
	dataTypes     map[reflect.Type]*OverridedDataType
	operationsIDS map[string]struct{}
	errors        []error
	fullNames     bool
	sortParams    bool
}

// NewGenerator returns a new OpenAPI generator.
func NewGenerator(conf *SpecGenConfig) (*Generator, error) {
	if conf == nil {
		return nil, errors.New("missing config")
	}
	components := &Components{
		Schemas:    make(map[string]*SchemaOrRef),
		Responses:  make(map[string]*ResponseOrRef),
		Parameters: make(map[string]*ParameterOrRef),
		Headers:    make(map[string]*HeaderOrRef),
	}
	return &Generator{
		config: conf,
		api: &OpenAPI{
			OpenAPI:    version,
			Info:       &Info{},
			Paths:      make(Paths),
			Components: components,
		},
		schemaTypes:   make(map[reflect.Type]struct{}),
		typeNames:     make(map[reflect.Type]string),
		dataTypes:     make(map[reflect.Type]*OverridedDataType),
		operationsIDS: make(map[string]struct{}),
		fullNames:     true,
		sortParams:    true,
	}, nil
}

// SpecGenConfig represents the configuration
// of the spec generator.
type SpecGenConfig struct {
	// Name of the tag used by the validator.v8
	// package. This is used by the spec generator
	// to determine if a field is required.
	ValidatorTag      string
	PathLocationTag   string
	QueryLocationTag  string
	HeaderLocationTag string
	EnumTag           string
	DefaultTag        string
}

// SetInfo uses the given OpenAPI info for the
// current specification.
func (g *Generator) SetInfo(info *Info) {
	g.api.Info = info
}

// SetServers sets the server list for the
// current specification.
func (g *Generator) SetServers(servers []*Server) {
	g.api.Servers = servers
}

// API returns a copy of the internal OpenAPI object.
func (g *Generator) API() *OpenAPI {
	cpy := *g.api
	return &cpy
}

// Errors returns the errors thar occurred during
// the generation of the specification.
func (g *Generator) Errors() []error {
	return g.errors
}

// UseFullSchemaNames defines whether the generator should generates
// a full name for the components using the package name of the type
// as a prefix.
// Omitting the package part of the name increases the risks of conflicts.
// It is the responsibility of the developper to ensure that unique type
// names are used across all the packages of the application.
// Default to true.
func (g *Generator) UseFullSchemaNames(b bool) {
	g.fullNames = b
}

// SetSortParams controls whether the generator should sort the
// parameters of an operation by location and name in ascending
// order.
func (g *Generator) SetSortParams(b bool) {
	g.sortParams = b
}

// OverrideTypeName registers a custom name for a
// type that will override the default generation
// and have precedence over types that implements
// the Typer interface.
func (g *Generator) OverrideTypeName(t reflect.Type, name string) error {
	if name == "" {
		return errors.New("type name is empty")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if _, ok := g.typeNames[t]; ok {
		return errors.New("type name already overrided")
	}
	g.typeNames[t] = name

	return nil
}

// OverrideDataType registers a custom schema type and
// format for the given type that will overrided the
// default generation.
func (g *Generator) OverrideDataType(t reflect.Type, typ, format string) error {
	if typ == "" {
		return errors.New("type is mandatory")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if _, ok := g.dataTypes[t]; ok {
		return errors.New("data type already overrided")
	}
	g.dataTypes[t] = &OverridedDataType{
		format: format,
		typ:    typ,
	}
	return nil
}

func (g *Generator) datatype(t reflect.Type) DataType {
	if dt, ok := g.dataTypes[t]; ok {
		return dt
	}
	return DataTypeFromType(t)
}

// AddTag adds a new tag to the OpenAPI specification.
// If a tag already exists with the same name, it is
// overwritten.
func (g *Generator) AddTag(name, desc string) {
	if name == "" {
		return
	}
	// Search for an existing tag with the same name,
	// and update its description before returning
	// if one is found.
	for _, tag := range g.api.Tags {
		if tag != nil {
			if tag.Name == name {
				tag.Description = desc
				return
			}
		}
	}
	// Add a new tag to the spec.
	g.api.Tags = append(g.api.Tags, &Tag{
		Name:        name,
		Description: desc,
	})
	sort.SliceStable(g.api.Tags, func(i, j int) bool {
		if g.api.Tags[i] != nil && g.api.Tags[j] != nil {
			return g.api.Tags[i].Name < g.api.Tags[j].Name
		}
		return false
	})
}

// AddOperation add a new operation to the OpenAPI specification
// using the method and path of the route and the tonic
// handler informations.
func (g *Generator) AddOperation(path, method, tag string, in, out reflect.Type, info *OperationInfo) (*Operation, error) {
	op := &Operation{
		ID: uuid.Must(uuid.NewV4()).String(),
	}
	path = rewritePath(path)

	if info != nil {
		// Ensure that the provided operation ID is unique.
		if _, ok := g.operationsIDS[info.ID]; ok {
			return nil, fmt.Errorf("ID %s is already used by another operation", info.ID)
		}
		g.operationsIDS[info.ID] = struct{}{}
	}
	// If a PathItem does not exists for this
	// path, create a new one.
	item, ok := g.api.Paths[path]
	if !ok {
		item = new(PathItem)
		g.api.Paths[path] = item
	}
	// Create a new operation and set it
	// to the according method of the PathItem.
	if info != nil {
		op.ID = info.ID
		op.Summary = info.Summary
		op.Description = info.Description
		op.Deprecated = info.Deprecated
		op.Responses = make(Responses)
	}
	if tag != "" {
		op.Tags = append(op.Tags, tag)
	}
	// Operations with methods GET/HEAD/DELETE cannot have a body.
	// Non parameters fields will be ignored.
	allowBody := method != http.MethodGet &&
		method != http.MethodHead &&
		method != http.MethodDelete

	if in != nil {
		if in.Kind() == reflect.Ptr {
			in = in.Elem()
		}
		if in.Kind() != reflect.Struct {
			return nil, errors.New("input type is not a struct")
		}
		if err := g.setOperationParams(op, in, in, allowBody, path); err != nil {
			return nil, err
		}
	}
	// Generate the default response from the tonic
	// handler return type. If the handler has no output
	// type, the response won't have a schema.
	if err := g.setOperationResponse(op, out, strconv.Itoa(info.StatusCode), tonic.MediaType(), info.StatusDescription, info.Headers); err != nil {
		return nil, err
	}
	// Generate additional responses from the operation
	// informations.
	for _, resp := range info.Responses {
		if resp != nil {
			if err := g.setOperationResponse(op,
				reflect.TypeOf(resp.Model),
				resp.Code,
				tonic.MediaType(),
				resp.Description,
				resp.Headers,
			); err != nil {
				return nil, err
			}
		}
	}
	setOperationBymethod(item, op, method)

	return op, nil
}

// rewritePath converts a Gin operation path that use
// colons and asterisks to declare path parameters, to
// an OpenAPI representation that use curly braces.
func rewritePath(path string) string {
	return ginPathParamRe.ReplaceAllString(path, "/{$1}")
}

// setOperationBymethod sets the operation op to the appropriate
// field of item according to the given method.
func setOperationBymethod(item *PathItem, op *Operation, method string) {
	switch method {
	case "GET":
		item.GET = op
	case "PUT":
		item.PUT = op
	case "POST":
		item.POST = op
	case "PATCH":
		item.PATCH = op
	case "HEAD":
		item.HEAD = op
	case "OPTIONS":
		item.OPTIONS = op
	case "TRACE":
		item.TRACE = op
	case "DELETE":
		item.DELETE = op
	}
}

func isResponseCodeRange(code string) bool {
	if len(code) != 3 {
		return false
	}
	// First char must be 1, 2, 3, 4 or 5.
	pre := code[0]
	if pre < 49 || pre > 53 {
		return false
	}
	// Last two chars are wildcard letter X.
	if code[1] != 'X' || code[2] != 'X' {
		return false
	}
	return true
}

// setOperationResponse adds a response to the operation that
// return the type t with the given media type and status code.
func (g *Generator) setOperationResponse(op *Operation, t reflect.Type, code, mt, desc string, headers []*ResponseHeader) error {
	if _, ok := op.Responses[code]; ok {
		// A response already exists for this code.
		return fmt.Errorf("response with code %s already exists", code)
	}
	// Check that the response code is valid per the spec:
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.2.md#patterned-fields-1
	if code != "default" {
		if !isResponseCodeRange(code) { // ignore ranges
			// Convert code to number and check that it is
			// between 100 and 599.
			ci, err := strconv.Atoi(code)
			if err != nil {
				return fmt.Errorf("invalid response code: %s", err)
			}
			if ci < 100 || ci > 599 {
				return fmt.Errorf("response code out of range: %s", code)
			}
			desc = http.StatusText(ci)
		}
	}
	r := &Response{
		Description: desc,
		Content:     make(map[string]*MediaTypeOrRef),
		Headers:     make(map[string]*HeaderOrRef),
	}
	// The response may have no content type specified,
	// in which case we don't assign a schema.
	schema := g.newSchemaFromType(t)
	if schema != nil {
		r.Content[mt] = &MediaTypeOrRef{MediaType: &MediaType{
			Schema: schema,
		}}
	}
	// Assign headers.
	for _, h := range headers {
		if h != nil {
			var sor *SchemaOrRef
			if h.Model == nil {
				// default to string if no type is given.
				sor = &SchemaOrRef{Schema: &Schema{Type: "string"}}
			} else {
				sor = g.newSchemaFromType(reflect.TypeOf(h.Model))
			}
			r.Headers[h.Name] = &HeaderOrRef{Header: &Header{
				Description: h.Description,
				Schema:      sor,
			}}
		}
	}
	op.Responses[code] = &ResponseOrRef{Response: r}

	return nil
}

// setOperationParams adds the fields of the struct type t
// to the given operation.
func (g *Generator) setOperationParams(op *Operation, t, parent reflect.Type, allowBody bool, path string) error {
	if t.Kind() != reflect.Struct {
		return errors.New("input type is not a struct")
	}
	if err := g.buildParamsRecursive(op, t, parent, allowBody); err != nil {
		return err
	}
	// Extract all the path parameter names.
	matches := paramsInPathRe.FindAllStringSubmatch(path, -1)
	var pathParams []string
	for _, m := range matches {
		pathParams = append(pathParams, m[1])
	}
	// Check that all declared path parameters are
	// defined in the operation.
	for _, pp := range pathParams {
		has := false
		for _, param := range op.Parameters {
			if param.In == "path" && param.Name == pp {
				has = true
				break
			}
		}
		if !has {
			return fmt.Errorf("semantic error for path %s: declared path parameter %s needs to be defined at operation level", path, pp)
		}
	}
	// Sort operations parameters by location and name
	// in ascending order.
	if g.sortParams {
		paramsOrderedBy(
			g.paramyByLocation,
			g.paramyByName,
		).Sort(op.Parameters)
	}
	return nil
}

func (g *Generator) buildParamsRecursive(op *Operation, t, parent reflect.Type, allowBody bool) error {
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		sft := sf.Type

		// Dereference pointer.
		if sft.Kind() == reflect.Ptr {
			sft = sft.Elem()
		}
		isUnexported := sf.PkgPath != ""

		if sf.Anonymous {
			if isUnexported && sft.Kind() != reflect.Struct {
				// Ignore embedded fields of unexported non-struct types.
				continue
			}
			// Do not ignore embedded fields of unexported struct
			// types since they may have exported fields, and recursively
			// use its fields as operations params. This allow developers
			// to reuse input models using type composition.
			if sft.Kind() == reflect.Struct {
				// If the type of the embedded struct is the same as
				// the topmost parent, skip it to avoid an infinite
				// recursive loop.
				if sft == parent {
					g.error(&FieldError{
						Message:  "recursive embedding",
						Name:     sf.Name,
						TypeName: g.typeName(parent),
						Type:     parent,
					})
				} else if err := g.buildParamsRecursive(op, sft, parent, allowBody); err != nil {
					return err
				}
			}
			continue
		} else if isUnexported {
			// Ignore unexported non-embedded fields.
			continue
		}
		if err := g.addStructFieldToOperation(op, t, i, allowBody); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) paramyByName(p1, p2 *ParameterOrRef) bool {
	return g.resolveParameter(p1).Name < g.resolveParameter(p2).Name
}

func (g *Generator) paramyByLocation(p1, p2 *ParameterOrRef) bool {
	return locationsOrder[g.resolveParameter(p1).In] < locationsOrder[g.resolveParameter(p2).In]
}

// addStructFieldToOperation add the struct field of the type
// t at index idx to the operation op. A field will be considered
// as a parameter if it has a valid location tag key, or it will
// be treated as part of the request body.
func (g *Generator) addStructFieldToOperation(op *Operation, t reflect.Type, idx int, allowBody bool) error {
	sf := t.Field(idx)

	param, err := g.newParameterFromField(idx, t)
	if err != nil {
		return err
	}
	if param != nil {
		// Check if a parameter with same name/location
		// already exists.
		for _, p := range op.Parameters {
			if p != nil && (p.Name == param.Name) && (p.In == param.In) {
				g.error(&FieldError{
					Message:           "duplicate parameter",
					Name:              param.Name,
					TypeName:          g.typeName(t),
					Type:              t,
					ParameterLocation: param.In,
				})
				return nil
			}
		}
		op.Parameters = append(op.Parameters, &ParameterOrRef{
			Parameter: param,
		})
	} else {
		if !allowBody {
			return nil
		}
		// If binding is disabled for this field, don't
		// add it to the request body. This allow using
		// a model type as an operation input while also
		// omitting some fields that are computed by the
		// server.
		if sf.Tag.Get("binding") == "-" {
			return nil
		}
		// The field is not a parameter, add it to
		// the request body.
		if op.RequestBody == nil {
			op.RequestBody = &RequestBody{
				Content: make(map[string]*MediaType),
			}
		}
		// Select the corresponding media type for the
		// given field tag, or default to any type.
		mt := tonic.MediaType()
		if mt == "" {
			mt = anyMediaType
		}
		var schema *Schema

		// Create the media type if no fields
		// have been added yet.
		if _, ok := op.RequestBody.Content[mt]; !ok {
			schema = &Schema{
				Type:       "object",
				Properties: make(map[string]*SchemaOrRef),
			}
			op.RequestBody.Content[mt] = &MediaType{
				Schema: &SchemaOrRef{Schema: schema},
			}
		} else {
			schema = op.RequestBody.Content[mt].Schema.Schema
		}
		fname := fieldNameFromTag(sf, mediaTags[tonic.MediaType()])

		// Check if a field with the same name already exists.
		if _, ok := schema.Properties[fname]; ok {
			g.error(&FieldError{
				Message:           "duplicate request body parameter",
				Name:              fname,
				TypeName:          g.typeName(t),
				Type:              t,
				ParameterLocation: "body",
			})
			return nil
		}

		var required bool
		// The required property of a field is not part of its
		// own schema but specified in the parent schema.
		if fname != "" && g.isStructFieldRequired(sf) {
			required = true
			schema.Required = append(schema.Required, fname)
			sort.Strings(schema.Required)
		}
		sfs := g.newSchemaFromStructField(sf, required, fname, t)
		if schema != nil {
			schema.Properties[fname] = sfs
		}
	}
	return nil
}

// newParameterFromField create a new operation parameter
// from the struct field at index idx in type in. Only the
// parameters of type path, query, header or cookie are concerned.
func (g *Generator) newParameterFromField(idx int, t reflect.Type) (*Parameter, error) {
	field := t.Field(idx)

	location, err := g.paramLocation(field, t)
	if err != nil {
		return nil, err
	}
	// The parameter location is empty, return nil
	// to indicate that the field is not a parameter.
	if location == "" {
		return nil, nil
	}
	name, err := tonic.ParseTagKey(field.Tag.Get(location))
	if err != nil {
		return nil, err
	}
	required := g.isStructFieldRequired(field)

	// Path parameters are always required.
	if location == g.config.PathLocationTag {
		required = true
	}
	// Consider invalid values as false.
	deprecated, _ := strconv.ParseBool(field.Tag.Get(deprecatedTag))

	p := &Parameter{
		Name:        name,
		In:          location,
		Description: field.Tag.Get(descriptionTag),
		Required:    required,
		Deprecated:  deprecated,
		Schema:      g.newSchemaFromStructField(field, required, name, t),
	}
	if field.Type.Kind() == reflect.Bool && location == g.config.QueryLocationTag {
		p.AllowEmptyValue = true
	}
	// Style.
	if location == g.config.QueryLocationTag {
		if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array {
			p.Explode = true // default
			p.Style = "form" // default in spec, but make it obvious
			if t := field.Tag.Get(tonic.ExplodeTag); t != "" {
				if explode, err := strconv.ParseBool(t); err == nil && !explode { // ignore invalid values
					p.Explode = explode
				}
			}
		}
	}
	return p, nil
}

// paramLocation parses the tags of the struct field to extract
// the location of an operation parameter.
func (g *Generator) paramLocation(f reflect.StructField, in reflect.Type) (string, error) {
	var c, p int

	has := func(name string, tag reflect.StructTag, i int) {
		if _, ok := tag.Lookup(name); ok {
			c++
			// save name position to extract
			// the value of the unique key.
			p = i
		}
	}
	// Count the number of keys that represents
	// a parameter location from the tag of the
	// struct field.
	var parameterLocations = []string{
		g.config.PathLocationTag,
		g.config.QueryLocationTag,
		g.config.HeaderLocationTag,
	}
	for i, n := range parameterLocations {
		has(n, f.Tag, i)
	}
	if c == 0 {
		// This will be considered to be part
		// of the request body.
		return "", nil
	}
	if c > 1 {
		return "", &FieldError{
			Message:  "conflicting parameter location",
			Name:     f.Name,
			TypeName: g.typeName(in),
			Type:     in,
		}
	}
	return parameterLocations[p], nil
}

// newSchemaFromStructField returns a new Schema builded
// from the field's type and its tags.
func (g *Generator) newSchemaFromStructField(sf reflect.StructField, required bool, fname string, parent reflect.Type) *SchemaOrRef {
	sor := g.newSchemaFromType(sf.Type)
	if sor == nil {
		return nil
	}
	// Get the underlying schema, it may be a reference
	// to a component, and update its fields using the
	// informations in the struct field tags.
	schema := g.resolveSchema(sor)

	if schema == nil {
		return sor
	}
	// Default value.
	// See section 'Common Mistakes' at
	// https://swagger.io/docs/specification/describing-parameters/
	if d := sf.Tag.Get(g.config.DefaultTag); d != "" {
		if required {
			g.error(&FieldError{
				Message:  "field cannot be required and have a default value",
				Name:     fname,
				Type:     sf.Type,
				TypeName: g.typeName(sf.Type),
				Parent:   parent,
			})
		} else {
			if v, err := stringToType(d, sf.Type); err != nil {
				g.error(&FieldError{
					Message:  fmt.Sprintf("default value %s cannot be converted to field type: %s", d, err),
					Name:     fname,
					Type:     sf.Type,
					TypeName: g.typeName(sf.Type),
					Parent:   parent,
				})
			} else {
				schema.Default = v
			}
		}
	}
	// Enum.
	// Must be applied to underlying items schema if the
	// parameter is an array, instead of the parameter schema.
	enum := g.enumFromStructField(sf, fname, parent)

	if schema.Type == "array" && schema.Items != nil {
		itemsSchema := g.resolveSchema(schema.Items)
		if itemsSchema != nil {
			itemsSchema.Enum = enum
		}
	} else {
		schema.Enum = enum
	}
	// Field description.
	if desc, ok := sf.Tag.Lookup(descriptionTag); ok {
		schema.Description = desc
	}
	// Deprecated.
	// Consider invalid values as false.
	schema.Deprecated, _ = strconv.ParseBool(sf.Tag.Get(deprecatedTag))

	// Update schema fields related to the JSON Validation
	// spec based on the content of the validator tag.
	schema = g.updateSchemaValidation(schema, sf)

	// Allow overidding schema properties that were
	// auto inferred manually via tags.
	if t, ok := sf.Tag.Lookup(formatTag); ok {
		schema.Format = t
	}
	return sor
}

func (g *Generator) enumFromStructField(sf reflect.StructField, fname string, parent reflect.Type) []interface{} {
	var enum []interface{}

	etag := sf.Tag.Get(g.config.EnumTag)
	if etag != "" {
		values := strings.Split(etag, ",")
		sftype := sf.Type

		// Use underlying element type if its an array or a slice.
		if sftype.Kind() == reflect.Slice || sftype.Kind() == reflect.Array {
			sftype = sftype.Elem()
		}
		for _, val := range values {
			if v, err := stringToType(val, sftype); err != nil {
				g.error(&FieldError{
					Message:  fmt.Sprintf("enum value %s cannot be converted to field type: %s", val, err),
					Name:     fname,
					Type:     sf.Type,
					TypeName: g.typeName(sf.Type),
					Parent:   parent,
				})
			} else {
				enum = append(enum, v)
			}
		}
	}
	return enum
}

// newSchemaFromType creates a new OpenAPI schema from
// the given reflect type.
func (g *Generator) newSchemaFromType(t reflect.Type) *SchemaOrRef {
	if t == nil {
		return nil
	}
	var nullable bool

	// Dereference pointer.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		nullable = true
	}
	dt := g.datatype(t)
	if dt == TypeUnsupported {
		g.error(&TypeError{
			Message: "unsupported type",
			Type:    t,
		})
		return nil
	}
	if dt == TypeComplex {
		switch t.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return g.buildSchemaRecursive(t)
		case reflect.Struct:
			return g.newSchemaFromStruct(t)
		}
	}
	schema := &Schema{
		Type:     dt.Type(),
		Format:   dt.Format(),
		Nullable: nullable,
	}
	return &SchemaOrRef{Schema: schema}
}

// buildSchemaRecursive recursively decomposes the complex
// type t into subsequent schemas.
func (g *Generator) buildSchemaRecursive(t reflect.Type) *SchemaOrRef {
	schema := &Schema{}

	switch t.Kind() {
	case reflect.Ptr:
		return g.buildSchemaRecursive(t.Elem())
	case reflect.Struct:
		return g.newSchemaFromStruct(t)
	case reflect.Map:
		// Map type is considered as a type "object"
		// and should declare underlying items type
		// in additional properties field.
		schema.Type = "object"

		// JSON Schema allow only strings as object key.
		if t.Key().Kind() != reflect.String {
			g.error(&TypeError{
				Message: "encountered type Map with keys of unsupported type",
				Type:    t,
			})
			return nil
		}
		schema.AdditionalProperties = g.buildSchemaRecursive(t.Elem())
	case reflect.Slice, reflect.Array:
		// Slice/Array types are considered as a type
		// "array" and should declare underlying items
		// type in items field.
		schema.Type = "array"

		// Go arrays have fixed size.
		if t.Kind() == reflect.Array {
			schema.MinItems = t.Len()
			schema.MaxItems = t.Len()
		}
		schema.Items = g.buildSchemaRecursive(t.Elem())
	default:
		dt := g.datatype(t)
		schema.Type, schema.Format = dt.Type(), dt.Format()
	}
	return &SchemaOrRef{Schema: schema}
}

// structSchema returns an OpenAPI schema that describe
// the Go struct represented by the type t.
func (g *Generator) newSchemaFromStruct(t reflect.Type) *SchemaOrRef {
	if t.Kind() != reflect.Struct {
		return nil
	}
	name := g.typeName(t)

	// If the type of the field has already been registered,
	// skip the schema generation to avoid a recursive loop.
	// We're not returning directly a reference from the components,
	// because there is no guarantee the generation is complete yet.
	if _, ok := g.schemaTypes[t]; ok {
		return &SchemaOrRef{Reference: &Reference{
			Ref: "#/components/schemas/" + name,
		}}
	}
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*SchemaOrRef),
	}
	// Register the type once before diving into
	// the recursive hole if it has a name. Anonymous
	// struct are all considered unique.
	if name != "" {
		g.schemaTypes[t] = struct{}{}
	}
	schema = g.flattenStructSchema(t, t, schema)

	sor := &SchemaOrRef{Schema: schema}

	// Register the schema within the speccomponents and return a
	// relative reference. Unnamed types, like anonymous structs,
	// will always be inlined in the specification.
	if name != "" {
		g.api.Components.Schemas[name] = sor

		return &SchemaOrRef{Reference: &Reference{
			Ref: "#/components/schemas/" + name,
		}}
	}
	// Return an inlined schema for types with no name.
	return sor
}

// flattenStructSchema recursively flatten the embedded
// fields of the struct type t to the given schema.
func (g *Generator) flattenStructSchema(t, parent reflect.Type, schema *Schema) *Schema {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ft := f.Type

		// Dereference pointer.
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		isUnexported := f.PkgPath != ""

		if f.Anonymous {
			if isUnexported && ft.Kind() != reflect.Struct {
				// Ignore embedded fields of unexported non-struct types.
				continue
			}
			// Do not ignore embedded fields of unexported struct
			// types since they may have exported fields.
			if ft.Kind() == reflect.Struct {
				// However, if the type of the embedded field is the same
				// as the topmost parent, skip it to avoid an infinite
				// recursive loop.
				if ft == parent {
					g.error(&FieldError{
						Message:  "recursive embedding",
						Name:     f.Name,
						TypeName: g.typeName(parent),
						Type:     parent,
						Parent:   parent,
					})
				} else {
					schema = g.flattenStructSchema(ft, parent, schema)
				}
			}
			continue
		} else if isUnexported {
			// Ignore unexported non-embedded fields.
			continue
		}
		fname := fieldNameFromTag(f, mediaTags[tonic.MediaType()])
		if fname == "" {
			// Field has no name, skip it.
			continue
		}
		var required bool
		// The required property of a field is not part of its
		// own schema but specified in the parent schema.
		if fname != "" && g.isStructFieldRequired(f) {
			required = true
			schema.Required = append(schema.Required, fname)
			sort.Strings(schema.Required)
		}
		sfs := g.newSchemaFromStructField(f, required, fname, t)
		if sfs != nil {
			schema.Properties[fname] = sfs
		}
	}
	return schema
}

// isStructFieldRequired returns whether a struct field
// is required. The information is read from the field
// tag 'binding'.
func (g *Generator) isStructFieldRequired(sf reflect.StructField) bool {
	if t, ok := sf.Tag.Lookup(g.config.ValidatorTag); ok {
		options := strings.Split(t, ",")
		for _, o := range options {
			// As soon as we see a 'dive' or 'keys'
			// options, the following options won't
			// apply to the given field.
			if o == "dive" || o == "keys" {
				return false
			}
			if o == "required" {
				return true
			}
		}
	}
	return false
}

// resolveSchema returns either the inlined schema
// in s or the one referenced in the API components.
func (g *Generator) resolveSchema(s *SchemaOrRef) *Schema {
	if s.Schema != nil && s.Reference == nil {
		return s.Schema
	}
	if s.Reference != nil {
		parts := strings.Split(s.Reference.Ref, "/")
		if len(parts) == 4 {
			if parts[0] == "#" && // relative ref
				parts[1] == "components" &&
				parts[2] == "schemas" &&
				parts[3] != "" {
				ref, ok := g.api.Components.Schemas[parts[3]]
				if ok && ref != nil {
					return ref.Schema
				}
				return nil
			}
		}
	}
	return nil
}

// resolveParameter returns either the inlined parameter
// in p or the one referenced in the API components.
func (g *Generator) resolveParameter(p *ParameterOrRef) *Parameter {
	// Parameters are always automatically inlined in the spec.
	if p.Parameter != nil && p.Reference == nil {
		return p.Parameter
	}
	return nil
}

// typeName returns the unique name of a type, which is
// the concatenation of the package name and the name
// of the given type, transformed to CamelCase without
// a dot separator between the two parts.
func (g *Generator) typeName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.PkgPath() == "" {
		// Predeclared or unnamed type, return an empty
		// name, the schema will be inlined in the spec.
		return ""
	}
	// If the name of the type was overidded,
	// use it in priority.
	if name, ok := g.typeNames[t]; ok {
		return name
	}
	// Create a new instance of t's type and use a
	// type assertion to check if it implements the
	// Typer interface.
	v := reflect.New(t)
	if v.CanInterface() {
		if tn, ok := v.Interface().(Typer); ok {
			return tn.TypeName()
		}
	}
	name := t.String() // package.name.
	sp := strings.Index(name, ".")
	pkg := name[:sp]

	// If the package is the main package, remove
	// the package part from the name.
	if pkg == "main" {
		pkg = ""
	}
	typ := name[sp+1:]

	if !g.fullNames {
		return strings.Title(typ)
	}
	return strings.Title(pkg) + strings.Title(typ)
}

// updateSchemaValidation fills the fields of the schema
// related to the JSON Schema Validation RFC based on the
// content of the validator tag.
// see https://godoc.org/gopkg.in/go-playground/validator.v8
func (g *Generator) updateSchemaValidation(schema *Schema, sf reflect.StructField) *Schema {
	ts := sf.Tag.Get(g.config.ValidatorTag)
	if ts == "" {
		return schema
	}
	ft := sf.Type
	if sf.Type.Kind() == reflect.Ptr {
		ft = sf.Type.Elem()
	}
	tags := strings.Split(ts, ",")

	for _, t := range tags {
		if t == "dive" || t == "keys" {
			break
		}
		// Tags can be joined together with an OR operator.
		parts := strings.Split(t, "|")

		for _, p := range parts {
			var k, v string
			// Split k/v pair using separator.
			sepIdx := strings.Index(p, "=")
			if sepIdx == -1 {
				k = p
			} else {
				k = p[:sepIdx]
				v = p[sepIdx+1:]
			}
			// Handle validators with value.
			switch k {
			case "len", "max", "min", "eq", "gt", "gte", "lt", "lte":
				n, err := strconv.Atoi(v)
				if err != nil {
					continue
				}
				switch k {
				case "len":
					setSchemaLen(schema, n, ft)
				case "max", "lte":
					setSchemaMax(schema, n, ft)
				case "min", "gte":
					setSchemaMin(schema, n, ft)
				case "lt":
					setSchemaMax(schema, n-1, ft)
				case "gt":
					setSchemaMin(schema, n+1, ft)
				case "eq":
					setSchemaEq(schema, n, ft)
				}
			}
		}
	}
	return schema
}

func (g *Generator) error(err error) {
	g.errors = append(g.errors, err)
}

// fieldTagName returns the name of a struct field
// extracted from a serialization tag using its name.
func fieldNameFromTag(sf reflect.StructField, tagName string) string {
	v, ok := sf.Tag.Lookup(tagName)
	if !ok {
		return sf.Name
	}
	parts := strings.Split(strings.TrimSpace(v), ",")

	// Split return a one item slice if
	// the input string is empty, thus we
	// don't check the length.
	name := parts[0]
	if name == "" {
		return sf.Name
	}
	if name == "-" {
		return ""
	}
	return name
}
