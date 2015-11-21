package jq

import "testing"

func TestJvKindName(t *testing.T) {
	if name := JvKindName(JvNull()); name != "null" {
		t.Errorf("Expected %v to equal %v", name, "null")
	}
}

func TestJvKind(t *testing.T) {
	if kind := JvGetKind(JvNull()); kind != JV_KIND_NULL {
		t.Errorf("Expected %v to equal %v", kind, "null")
	}
}
