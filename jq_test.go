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

	// We should be able to safely close multiple times.
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

func TestInvalidJsonInput(t *testing.T) {
	jq, err := New(make(chan error))
	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}
	defer jq.Close()

	err = jq.SetJsonInput("Not json")

	if err == nil {
		t.Error("Expected an error parsing invalid JSON input but none was returned")
	}
}

func TestSimpleProgram(t *testing.T) {
	errs := make(chan error)
	jq, err := New(errs)

	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}
	defer jq.Close()

	go func() {
		// We shouldn't see any errors reported. If we do it's an error
		for err := range errs {
			t.Errorf("Expected no errors, but got %#v", err)
		}
	}()

	err = jq.SetJsonInput("{\"a\": 123}")
	if err != nil {
		t.Error(err)
	}

	res, err := jq.Execute(".a")

	if err != nil {
		t.Errorf("%#v", err)
	} else if res != 123 {
		t.Errorf("Got %#v, expected %#v", res, 123)
	}

}
