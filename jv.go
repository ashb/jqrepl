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
