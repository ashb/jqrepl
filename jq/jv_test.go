package jq_test

import (
	"testing"

	"github.com/ashb/jq-repl/jq"
)

func TestJvKindName(t *testing.T) {
	if name := jq.JvKindName(jq.JvNull()); name != "null" {
		t.Errorf("Expected %v to equal %v", name, "null")
	}
}

func TestJvKind(t *testing.T) {
	if kind := jq.JvGetKind(jq.JvNull()); kind != jq.JV_KIND_NULL {
		t.Errorf("Expected %v to equal %v", kind, "null")
	}
}
