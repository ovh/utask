package openapi

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/satori/go.uuid"
)

var (
	tofDataType = reflect.TypeOf((*DataType)(nil)).Elem()

	// Native.
	tofTime      = reflect.TypeOf(time.Time{})
	tofDuration  = reflect.TypeOf(time.Duration(0))
	tofByteSlice = reflect.TypeOf([]byte{})
	tofNetIP     = reflect.TypeOf(net.IP{})
	tofNetURL    = reflect.TypeOf(url.URL{})

	// Imported.
	tofSatoriUUID = reflect.TypeOf(uuid.UUID{})
)

var _ DataType = (*InternalDataType)(nil)
var _ DataType = (*OverridedDataType)(nil)

// Typer is the interface implemented
// by the types that can describe themselves.
type Typer interface {
	TypeName() string
}

// DataType is the interface implemented by types
// that can describe their OAS3 data type and format.
type DataType interface {
	Type() string
	Format() string
}

// InternalDataType represents an internal type.
type InternalDataType int

// OverridedDataType represents a data type
// which details override the default generation.
type OverridedDataType struct {
	format string
	typ    string
}

// Format implements DataType for OverridedDataType.
func (dt *OverridedDataType) Format() string { return dt.format }

// Type implements DataType for OverridedDataType.
func (dt *OverridedDataType) Type() string { return dt.typ }

// Type constants.
const (
	TypeInteger InternalDataType = iota
	TypeLong
	TypeFloat
	TypeDouble
	TypeString
	TypeByte
	TypeBinary
	TypeBoolean
	TypeDate
	TypeDateTime
	TypeDuration
	TypeIP
	TypeURL
	TypePassword

	// TypeComplex represents non-primitive types like
	// Go struct, for which a schema must be generated.
	TypeComplex

	// Imported data types.
	TypeUUID

	TypeUnsupported
)

// String implements fmt.Stringer for DataType.
func (dt InternalDataType) String() string {
	if 0 <= dt && dt < InternalDataType(len(datatypes)) {
		return datatypes[dt]
	}
	return ""
}

// Type returns the type corresponding to the DataType.
func (dt InternalDataType) Type() string {
	if 0 <= dt && dt < InternalDataType(len(types)) {
		return types[dt]
	}
	return ""
}

// Format returns the format corresponding to the DataType.
func (dt InternalDataType) Format() string {
	if 0 <= dt && dt < InternalDataType(len(formats)) {
		return formats[dt]
	}
	return ""
}

// DataTypeFromType returns a DataType for the given type.
func DataTypeFromType(t reflect.Type) DataType {
	// If the type implement the DataType interface,
	// return a new instance of the type.
	if t.Implements(tofDataType) {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return reflect.New(t).Interface().(DataType)
	}
	// Dereference any pointer.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Switch over Golang types.
	switch t {
	case tofTime:
		return TypeDateTime
	case tofDuration:
		return TypeDuration
	case tofByteSlice:
		return TypeByte
	case tofNetIP:
		return TypeIP
	case tofNetURL:
		return TypeURL
	}
	// Treat imported types.
	if dt := isImportedType(t); dt != nil {
		return dt
	}
	// Switch over primitive types.
	switch t.Kind() {
	case reflect.Int64, reflect.Uint64:
		return TypeLong
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return TypeInteger
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return TypeInteger
	case reflect.Float32:
		return TypeFloat
	case reflect.Float64:
		return TypeDouble
	case reflect.Bool:
		return TypeBoolean
	case reflect.String:
		return TypeString
	case reflect.Map, reflect.Struct, reflect.Array, reflect.Slice:
		return TypeComplex
	default:
		// reflect.Func, reflect.Chan, reflect.Uintptr, reflect.UnsafePointer,
		// reflect.Interface, reflect.Complex64, reflect.Complex128, ...
		return TypeUnsupported
	}
}

func isImportedType(t reflect.Type) DataType {
	// github.com/satori/go.uuid
	if t == tofSatoriUUID {
		return TypeUUID
	}
	return nil
}

// stringToType converts val to t's type and return the new value.
func stringToType(val string, t reflect.Type) (interface{}, error) {
	// Compare type to know Golang types.
	// IT MUST BE EXECUTED BEFORE swithing over
	// primitives because a time.Duration is itself
	// an int64.
	if t.AssignableTo(tofTime) {
		return time.Parse(time.RFC3339, val)
	}
	if t.AssignableTo(tofDuration) {
		return time.ParseDuration(val)
	}
	switch t.Kind() {
	case reflect.Bool:
		// ParseBool returns an error if the value
		// is invalid and cannot be converted to a
		// boolean. We assume that invalid values
		// are always falsy.
		v, _ := strconv.ParseBool(val)
		return v, nil
	case reflect.String:
		return val, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(val, 10, t.Bits())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.ParseUint(val, 10, t.Bits())
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(val, t.Bits())
	}
	return nil, fmt.Errorf("unknown type %s", t.String())
}

var datatypes = [...]string{
	TypeInteger:     "Integer",
	TypeLong:        "Long",
	TypeFloat:       "Float",
	TypeDouble:      "Double",
	TypeString:      "String",
	TypeByte:        "Byte",
	TypeBinary:      "Binary",
	TypeBoolean:     "Boolean",
	TypeDate:        "Date",
	TypeDateTime:    "DateTime",
	TypeDuration:    "Duration",
	TypeIP:          "IP Address",
	TypeURL:         "URL (Uniform Resource Locator)",
	TypePassword:    "Password",
	TypeUnsupported: "Unsupported",
	TypeComplex:     "Complex",
	TypeUUID:        "UUID",
}

var types = [...]string{
	TypeInteger:  "integer",
	TypeLong:     "integer",
	TypeFloat:    "number",
	TypeDouble:   "number",
	TypeString:   "string",
	TypeByte:     "string",
	TypeBinary:   "string",
	TypeBoolean:  "boolean",
	TypeDate:     "string",
	TypeDateTime: "string",
	TypeDuration: "string",
	TypeIP:       "string",
	TypeURL:      "string",
	TypePassword: "string",
	TypeComplex:  "string",
	TypeUUID:     "string",
}

var formats = [...]string{
	TypeInteger:  "int32",
	TypeLong:     "int64",
	TypeFloat:    "float",
	TypeDouble:   "double",
	TypeString:   "",
	TypeByte:     "byte",
	TypeBinary:   "binary",
	TypeBoolean:  "",
	TypeDate:     "date",
	TypeDateTime: "date-time",
	TypeDuration: "duration",
	TypeIP:       "ip",
	TypeURL:      "url",
	TypePassword: "password",
	TypeComplex:  "",
	TypeUUID:     "uuid",
}
