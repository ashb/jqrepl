package jq

/*
#cgo CFLAGS: -I ${SRCDIR}/../jq-1.5/BUILD/include
#cgo LDFLAGS: ${SRCDIR}/../jq-1.5/BUILD/lib/libjq.a

#include <stdlib.h>

#include <jv.h>
#include <jq.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// Helper functions for dealing with JV objects. You can't use this from
// another go package as the cgo types are 'unique' per go package

// JvKind represents the type of value that a `Jv` contains.
type JvKind int

// Jv represents a JSON value from libjq.
//
// The go wrapper uses the same memory management semantics as the underlying C
// library, so you should familiarize yourself with
// https://github.com/stedolan/jq/wiki/C-API:-jv#memory-management. In summary
// this package and all JQ functions operate on the assumption that any jv value
// you pass to a function is then owned by that function -- if you do not wish
// this to be the case call Copy() on it first.
type Jv struct {
	jv C.jv
}

const (
	JV_KIND_INVALID JvKind = C.JV_KIND_INVALID
	JV_KIND_NULL    JvKind = C.JV_KIND_NULL
	JV_KIND_FALSE   JvKind = C.JV_KIND_FALSE
	JV_KIND_TRUE    JvKind = C.JV_KIND_TRUE
	JV_KIND_NUMBER  JvKind = C.JV_KIND_NUMBER
	JV_KIND_STRING  JvKind = C.JV_KIND_STRING
	JV_KIND_ARRAY   JvKind = C.JV_KIND_ARRAY
	JV_KIND_OBJECT  JvKind = C.JV_KIND_OBJECT
)

// String returns a string representation of what type this Jv contains
func (kind JvKind) String() string {
	// Rather than rely on converting from a C string to go every time, store our
	// own list
	switch kind {
	case JV_KIND_INVALID:
		return "<invalid>"
	case JV_KIND_NULL:
		return "null"
	case JV_KIND_FALSE:
		return "boolean"
	case JV_KIND_TRUE:
		return "boolean"
	case JV_KIND_NUMBER:
		return "number"
	case JV_KIND_STRING:
		return "string"
	case JV_KIND_ARRAY:
		return "array"
	case JV_KIND_OBJECT:
		return "object"
	default:
		return "<unkown>"
	}
}

// JvNull returns a value representing a JSON null
func JvNull() *Jv {
	return &Jv{C.jv_null()}
}

// JvFromString returns a new jv string-typed value containing the given go
// string.
func JvFromString(str string) *Jv {
	cs := C.CString(str)
	defer C.free(unsafe.Pointer(cs))
	return &Jv{C.jv_string_sized(cs, C.int(len(str)))}
}

// Covert a JQ error stored in a JV error to a native go error
func _ConvertError(inv C.jv) error {
	// We might want to not call this as it prefixes things with "jq: "
	jv := &Jv{C.jq_format_error(inv)}
	defer jv.Free()

	return errors.New(jv._string())
}

// JvFromJSONString takes a JSON string and returns the jv representation of
// it.
func JvFromJSONString(str string) (*Jv, error) {
	cs := C.CString(str)
	defer C.free(unsafe.Pointer(cs))
	jv := C.jv_parse(cs)

	if C.jv_is_valid(jv) == 0 {
		return nil, _ConvertError(jv)
	}
	return &Jv{jv}, nil
}

// Free this reference to a Jv value.
//
// Don't call this more than once per jv - might not actually free the memory
// as libjq uses reference counting. To make this more like the libjq interface
// we return a nil pointer.
func (jv *Jv) Free() *Jv {
	C.jv_free(jv.jv)
	return nil
}

// Kind returns a JvKind saying what type this jv contains.
//
// Does not consume the invocant.
func (jv *Jv) Kind() JvKind {
	return JvKind(C.jv_get_kind(jv.jv))
}

// Copy returns a *Jv so that the original won't get freed.
//
// Does not consume the invocant.
func (jv *Jv) Copy() *Jv {
	C.jv_copy(jv.jv)
	// Becasue jv uses ref counting under the hood we can return the same value
	return jv
}

// IsValid returns true if this Jv represents a valid JSON type, or false if it
// is unitiaizlied or if it represents an error type
//
// Does not consume the invocant.
func (jv *Jv) IsValid() bool {
	return C.jv_is_valid(jv.jv) != 0
}

// GetInvalidMessage gets the error message for this Jv. If there is none it
// will return a jv NULL value (not a go nil value).
//
// Consumes the invocant.
func (jv *Jv) GetInvalidMessage() *Jv {
	return &Jv{C.jv_invalid_get_msg(jv.jv)}
}

func (jv *Jv) _string() string {
	// Raw string value. If called on
	cs := C.jv_string_value(jv.jv)
	// Don't free cs - freed when the jv is
	return C.GoString(cs)
}

// If jv is a string, return its value. Will not stringify other types
//
// Does not consume the invocant.
func (jv *Jv) String() (string, error) {
	// Doing this might be a bad idea as it means we almost implement the Stringer
	// interface but not quite (cos the error type)

	// If we don't do this check JV will assert
	if C.jv_get_kind(jv.jv) != C.JV_KIND_STRING {
		return "", fmt.Errorf("Cannot return String for jv of type %s", jv.Kind())
	}

	return jv._string(), nil
}

// ToGoVal converts a jv into it's closest Go approximation
//
// Does not consume the invocant.
func (jv *Jv) ToGoVal() interface{} {
	switch kind := C.jv_get_kind(jv.jv); kind {
	case C.JV_KIND_NULL:
		return nil
	case C.JV_KIND_FALSE:
		return false
	case C.JV_KIND_TRUE:
		return true
	case C.JV_KIND_NUMBER:
		dbl := C.jv_number_value(jv.jv)

		if C.jv_is_integer(jv.jv) == 0 {
			return float64(dbl)
		}
		return int(dbl)
	case C.JV_KIND_STRING:
		return jv._string()
	case C.JV_KIND_ARRAY:
		fallthrough
	case C.JV_KIND_OBJECT:
		panic(fmt.Sprintf("ToGoVal not implemented for %#v", kind))
	default:
		panic(fmt.Sprintf("Unknown JV kind %d", kind))
	}
}
