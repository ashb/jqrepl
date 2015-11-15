package jqrepl

import (
	"strings"
	"testing"
)

func TestNewClose(t *testing.T) {
	jq, err := New(make(chan error))

	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}

	jq.Close()
	if jq._state != nil {
		t.Error("Expected jq._state to be nil after Close")
	}
	jq.Close()

}

func TestCompileError(t *testing.T) {
	errs := make(chan error, 100)
	jq, err := New(errs)

	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}
	defer jq.Close()

	const program = "a b"
	if _, err = jq.Execute(program); err == nil {
		t.Error("Error was expected to be not nil")
	}

	// JQ might (and currently does) report multiple errors. One of them will
	// contain our input program. Check for that but don't be overly-specific
	// about the string or order of errors
ForErrs:
	for {
		select {
		case err := <-errs:
			if strings.Contains(err.Error(), program) {
				break ForErrs
			}
		default:
			t.Error("No errors, or no error contained the program string")
			break ForErrs
		}
	}
}
