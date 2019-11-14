
<h1 align="center">Fizz</h1>
<p align="center"><img src="images/lemon.png" height="200px" width="auto" alt="Gin Fizz"></p><p align="center">Fizz is a wrapper for <strong>Gin</strong> based on <i>gadgeto/tonic</i>.</p>
<p align="center">It generates wrapping gin-compatible handlers that do all the repetitive work and wrap the call to your handlers. It can also generates an *almost* complete <strong>OpenAPI 3</strong> specification of your API.</p>
<p align="center"><br>
<a href="https://godoc.org/github.com/wI2L/fizz"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"></a> <a href="https://goreportcard.com/report/wI2L/fizz"><img src="https://goreportcard.com/badge/github.com/wI2L/fizz"></a> <a href="https://travis-ci.org/wI2L/fizz"><img src="https://travis-ci.org/wI2L/fizz.svg?branch=master"></a> <a href="https://codecov.io/gh/wI2L/fizz"><img src="https://codecov.io/gh/wI2L/fizz/branch/master/graph/badge.svg"/></a> <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
<br>
</p>

---

To create a Fizz instance, you can pass an existing *Gin* engine to `fizz.NewFromEngine`, or use `fizz.New` that will use a new default *Gin* engine.

```go
engine := gin.Default()
engine.Use(...) // register global middlewares

f := fizz.NewFromEngine(engine)
```

A Fizz instance implements the `http.HandlerFunc` interface, which means it can be used as the base handler of your HTTP server.
```go
srv := &http.Server{
   Addr:    ":4242",
   Handler: f,
}
srv.ListenAndServe()
```

### Handlers

Fizz abstracts the `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`, `HEAD` and `TRACE` methods of a *Gin* engine. These functions accept a variadic list of handlers as the last parameter, but since Fizz relies on *tonic* to retrieve the informations required to generate the *OpenAPI* specification of the operation, **only one of the handlers registered MUST be wrapped with Tonic**.

In the following example, the `BarHandler` is a simple middleware that will be executed before the `FooHandler`, but the generator will use the input/output type of the `FooHandler` to generate the specification of the operation.

```go
func BarHandler(c *gin.Context) { ... }
func FooHandler(*gin.Context, *Foo) (*Bar, error) { ... }

fizz := fizz.New()
fizz.GET("/foo/bar", nil, BarHandler, tonic.Handler(FooHandler, 200))
```

However, registering only standard handlers that follow the `gin.HandlerFunc` signature is accepted, but the *OpenAPI* generator will ignore the operation and it won't appear in the specification.

### Operation informations

To enrich an operation, you can pass a list of optional `OperationOption` functions as the second parameters of the `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS` and `HEAD` methods.

```go
// Set the default response description.
// A default status text will be created from the code if it is omitted.
fizz.StatusDescription(desc string)

// Set the summary of the operation.
fizz.Summary(summary string)
fizz.Summaryf(format string, a ...interface{})

// Set the description of the operation.
fizz.Description(desc string)
fizz.Descriptionf(format string, a ...interface{})

// Override the ID of the operation.
// Must be a unique string used to identify the operation among
// all operations described in the API.
fizz.ID(id string)

// Mark the operation as deprecated.
fizz.Deprecated(deprecated bool)

// Add an additional response to the operation.
// model and header may be `nil`.
fizz.Response(statusCode, desc string, model interface{}, headers []*ResponseHeader)

// Add an additional header to the default response.
// Model can be of any type, and may also be `nil`,
// in which case the string type will be used as default.
fizz.Header(name, desc string, model interface{})

// Override the binding model of the operation.
fizz.InputModel(model interface{})
```

**NOTES:**
* `fizz.InputModel` allows to override the operation input regardless of how the handler implementation really binds the request parameters. It is the developer responsibility to ensure that the binding matches the OpenAPI specification.
* The fist argument of the `fizz.Reponse` method which represents an HTTP status code is of type *string* because the spec accept the value `default`. See the [Responses Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#responsesObject) documentation for more informations.

To help you declare additional headers, predefined variables for Go primitives types that you can use as the third argument of the `fizz.Header` method are available.
```go
Integer  int32
Long     int64
Float    float32
Double   float64
String   string
Byte     []byte
Binary   []byte
Boolean  bool
DateTime time.Time
```

### Groups

Exactly like you would do with *Gin*, you can create a group of routes using the method `Group`. Unlike *Gin* own method, Fizz's one takes two other optional arguments, `name` and `description`. These parameters will be used to create a tag in the **OpenAPI** specification that will be applied to all the routes added to the group.

```go
grp := f.Group("/subpath", "MyGroup", "Group description", middlewares...)
```
If the `name` parameter is empty, the tag won't be created and it won't be used.

Subgroups of subgroups can be created to an infinite depth, according yo your needs.

```go
foo := f.Group("/foo", "Foo", "Foo group")

// all routes registered on group bar will have
// a relative path starting with /foo/bar
bar := f.Group("/bar", "Bar", "Bar group")

// /foo/bar/{barID}
bar.GET("/:barID", nil, tonic.Handler(MyBarHandler, 200))
```

The `Use` method can be used with groups to register middlewares after their creation.
```go
grp.Use(middleware1, middleware2, ...)
```

## Tonic

The subpackage *tonic* handles path/query/header/body parameters binding in a single consolidated input object which allows you to remove all the boilerplate code that retrieves and tests the presence of various parameters. The *OpenAPI* generator make use of the input/output types informations of a tonic-wrapped handler reported by *tonic* to document the operation in the specification.

The handlers wrapped with *tonic* must follow the following signature.
```go
func(*gin.Context, [input object ptr]) ([output object], error)
```
Input and output objects are both optional, as such, the minimal accepted signature is:
```go
func(*gin.Context) error
```

To wrap a handler with *tonic*, use the `tonic.Handler` method. It takes a function that follow the above signature and a default status code and return a `gin.HandlerFunc` function that can be used when you register a route with Fizz of *Gin*.

Output objects can be of any type, and will be marshalled to the desired media type.
Note that the input object **MUST always be a pointer to a struct**, or the tonic wrapping will panic at runtime.

If you use closures as handlers, please note that they will all have the same name, and the generator will return an error. To overcome this problem, you have to explicitely set the ID of an operation when you register the handler.

```go
func MyHandler() gin.HandlerFunc {
   return tonic.Handler(func(c *gin.Context) error {}, 200)
}

fizz.GET("/foo", []fizz.OperationOption{
   fizz.ID("MyOperationID")
}, MyHandler())
```

### Location tags

*tonic* uses three struct tags to recognize the parameters it should bind to the input object of your tonic-wrapped handlers:
- `path`: bind from the request path
- `query`: bind from the query string
- `header`: bind from the request header

The fields that doesn't use one of these tags will be considered as part of the request body.

The value of each struct tag represents the name of the field in each location, with options.
```go
type MyHandlerParams struct {
   ID  int64     `path:"id"`
   Foo string    `query:"foo"`
   Bar time.Time `header:"x-foo-bar"`
}
```

*tonic* will automatically convert the value extracted from the location described by the tag to the appropriate type before binding.

**NOTE**: A path parameter is always required and will appear required in the spec regardless of the `validate` tag content.

### Additional tags

You can use additional tags. Some will be interpreted by *tonic*, others will be exclusively used to enrich the *OpenAPI* specification.
- `default`: *tonic* will bind this value if none was passed with the request. This should not be used if a field is also required. Read the [documentation](https://swagger.io/docs/specification/describing-parameters/) (section _Common Mistakes_) for more informations about this behaviour.
- `description`: Add a description of the field in the spec.
- `deprecated`: Indicates if the field is deprecated. Accepted values are _1_, _t_, _T_, _TRUE_, _true_, _True_, _0_, _f_, _F_, _FALSE_. Invalid value are considered to be false.
- `enum`: A coma separated list of acceptable values for the parameter.
- `format`: Override the format of the field in the specification. Read the [documentation](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#dataTypeFormat) for more informations.
- `validate`: Field validation rules. Read the [documentation](https://godoc.org/gopkg.in/go-playground/validator.v8) for more informations.
- `explode`: Specifies whether arrays should generate separate parameters for each array item or object property (limited to query parameters with *form* style). Accepted values are _1_, _t_, _T_, _TRUE_, _true_, _True_, _0_, _f_, _F_, _FALSE_. Invalid value are considered to be false.

### JSON/XML

The JSON/XML encoders usually omit a field that has the tag `"-"`. This behaviour is reproduced by the *OpenAPI* generator ; a field with this tag won't appear in the properties of the schema.

In the following example, the field `Input` is used only for binding request body parameters and won't appear in the output encoding while `Output` will be marshaled but will not be used for parameters binding.
```go
type Model struct {
	Input  string `json:"-"`
	Output string `json:"output" binding:"-"`
}
```

### Request body

If you want to make a request body field mandatory, you can use the tag `validate:"required"`. The validator used by *tonic* will ensure that the field is present.
To be able to make a difference between a missing value and the zero value of a type, use a pointer.

To explicitly ignore a parameter from the request body, use the tag `binding:"-"`.

Note that the *OpenAPI* generator will ignore request body parameters for the routes with a method that is one of `GET`, `DELETE` or `HEAD`.
   > GET, DELETE and HEAD are no longer allowed to have request body because it does not have defined semantics as per [RFC 7231](https://tools.ietf.org/html/rfc7231#section-4.3).
	[*source*](https://swagger.io/docs/specification/describing-request-body/)

### Schema validation

The *OpenAPI* generator recognize some tags of the [go-playground/validator.v8](https://gopkg.in/go-playground/validator.v8) package and translate those to the [properties of the schema](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.1.md#properties) that are taken from the [JSON Schema definition](http://json-schema.org/latest/json-schema-validation.html#rfc.section.6).

The supported tags are: [len](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Length), [max](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Maximum), [min](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Mininum), [eq](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Equals), [gt](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Greater_Than), [gte](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Greater_Than_or_Equal), [lt](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Less_Than), [lte](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Less_Than_or_Equal).

Based on the type of the field that carry the tag, the fields `maximum`, `minimum`, `minLength`, `maxLength`, `minIntems`, `maxItems`, `minProperties` and `maxProperties` of its **JSON Schema** will be filled accordingly.

## OpenAPI specification

To serve the generated OpenAPI specification in either `JSON` or `YAML` format, use the handler returned by the `fizz.OpenAPI` method.

To enrich the specification, you can provide additional informations. Head to the [OpenAPI 3 spec](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#infoObject) for more informations about the API informations that you can specify, or take a look at the type `openapi.Info` in the file [_openapi/spec.go_](openapi/spec.go#L25).

```go
infos := &openapi.Info{
   Title:       "Fruits Market",
   Description: `This is a sample Fruits market server.`,
   Version:     "1.0.0",
}
f.GET("/openapi.json", nil, fizz.OpenAPI(infos, "json"))
```
**NOTE**: The generator will never panic. However, it is strongly recommended to call `fizz.Errors` to retrieve and handle the errors that may have occured during the generation of the specification before starting your API.

#### Components

The output types of your handlers are registered as components within the generated specification. By default, the name used for each component is composed of the package and type name concatenated using _CamelCase_ style, and does not contain the full import path. As such, please ensure that you don't use the same type name in two eponym package in your application.

The names of the components can be customized in two different ways.

##### Global override

Override the name of a type globally before registering your handlers. This has the highest precedence.
```go
f := fizz.New()
f.Generator().OverrideTypeName(reflect.TypeOf(T{}), "OverridedName")
```

##### Interface

Implements the `openapi.Typer` interface on your types.
```go
func (*T) TypeName() string { return "OverridedName" }
```
**WARNING:** You **MUST** not rely on the method receiver to return the name, because the method will be called on a new instance created by the generator with the `reflect` package.

#### Custom schemas

The spec generator creates OpenAPI schemas for your types based on their [reflection kind](https://golang.org/pkg/reflect/#Kind).
If you want to control the output schema of a type manually, you can implement the `DataType` interface for this type.

For example, given a UUID version 4 type, declared as a struct, that should appear as a string with a custom format.
```go
type UUIDv4 struct { ... }

func (*UUIDv4) Format() string { return "uuid" }
func (*UUIDv4) Type() string { return "string" }
```

The schema of the type will look like the following instead of describing all the fields of the struct.
```json
{
   "type": "string",
   "format": "uuid"
}
```
**WARNING:** You **MUST** not rely on the method receivers to return the type and format, because these methods will be called on a new instance created by the generator with the `reflect` package.

You can also override manually the type and format using `OverrideDataType()`. This has the highest precedence.
```go
fizz.Generator().OverrideDataType(reflect.TypeOf(&UUIDv4{}), "string", "uuid")
```

##### Native and imported types support

Fizz supports some native and imported types. A schema with a proper type and format will be generated automatically, removing the need for creating your own custom schema.

* [time.Time](https://golang.org/pkg/time/#Time)
* [time.Duration](https://golang.org/pkg/time/#Duration)
* [net.URL](https://golang.org/pkg/net/url/#URL)
* [net.IP](https://golang.org/pkg/net/#IP)  
Note that, according to the doc, the inherent version of the address is a semantic property, and thus cannot be determined by Fizz. Therefore, the format returned is simply `ip`. If you want to specify the version, you can use the tags `format:"ipv4"` or `format:"ipv6"`.
* [uuid.UUID](https://godoc.org/github.com/satori/go.uuid#UUID)

#### Markdown

> Throughout the specification description fields are noted as supporting CommonMark markdown formatting. Where OpenAPI tooling renders rich text it MUST support, at a minimum, markdown syntax as described by CommonMark 0.27. Tooling MAY choose to ignore some CommonMark features to address security concerns.
[*source*](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.1.md#rich-text-formatting)

To help you write markdown descriptions in Go, a simple builder is available in the sub-package `markdown`. This is quite handy to avoid conflicts with backticks that are both used in Go for litteral multi-lines strings and code blocks in markdown.

## Known limitations

- Since *OpenAPI* is based on the *JSON Schema* specification itself, objects (Go maps) with keys that are not of type `string` are not supported and will be ignored during the generation of the specification.
- Recursive embedding of the same type is not supported, at any level of recursion. The generator will warn and skip the offending fields.
   ```go
   type A struct {
      Foo int
      *A   // ko, embedded and same type as parent
      A *A // ok, not embedded
      *B   // ok, different type
   }

   type B struct {
      Bar string
      *A // ko, type B is embedded in type A
      *C // ok, type C does not contains an embedded field of type A
   }

   type C struct {
      Baz bool
   }
   ```

## Examples

A simple runnable API is available in `examples/market`.
```shell
go build
./market
# Retrieve the specification marshaled in JSON.
curl -i http://localhost:4242/openapi.json
```

## Credits

Fizz is based on [gin-gonic/gin](https://github.com/gin-gonic/gin) and use [gadgeto/tonic](https://github.com/loopfz/gadgeto/tree/master/tonic). :heart:

<p align="right"><img src="https://forthebadge.com/images/badges/built-with-swag.svg"></p>
