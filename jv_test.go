package jqrepl

import "testing"

func TestJvKindName(t *testing.T) {
	if name := JvKindName(JvNull()); name != "null" {
		t.Error("Expected %q to equal %q", name, "null")
	}
}

func TestJvKind(t *testing.T) {
	if kind := JvGetKind(JvNull()); kind != JV_KIND_NULL {
		t.Error("Expected %q to equal %q", kind, "null")
	}
}
