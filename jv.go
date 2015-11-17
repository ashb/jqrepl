package jqrepl

/*
#cgo CFLAGS: -I ${SRCDIR}/jq-1.5/BUILD/include
#cgo LDFLAGS: ${SRCDIR}/jq-1.5/BUILD/lib/libjq.a

#include <jv.h>
*/
import "C"

// Helper functions for dealing with JV objects. You can't use this from
// another go package as the cgo types are 'unique' per go package

type JvKind int

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

func JvNull() C.jv {
	return C.jv_null()
}

func JvGetKind(jv C.jv) JvKind {
	return JvKind(C.jv_get_kind(jv))
}

func JvKindName(jv C.jv) string {
	name := C.jv_kind_name(C.jv_get_kind(jv))
	return C.GoString(name)
}

func JvStringValue(jv C.jv) string {
	cs := C.jv_string_value(jv)
	// Don't free cs - freed when the jv is
	return C.GoString(cs)
}

func JvToGoVal(jv C.jv) interface{} {
	switch C.jv_get_kind(jv) {
	case C.JV_KIND_NULL:
		return nil
	case C.JV_KIND_FALSE:
		return false
	case C.JV_KIND_TRUE:
		return true
	case C.JV_KIND_NUMBER:
		dbl := C.jv_number_value(jv)

		if C.jv_is_integer(jv) == 0 {
			return float64(dbl)
		} else {
			return int(dbl)
		}
	case C.JV_KIND_STRING:
		return JvStringValue(jv)
	case C.JV_KIND_ARRAY:
		return nil
	case C.JV_KIND_OBJECT:
		return nil
	default:
		return nil
	}
}
