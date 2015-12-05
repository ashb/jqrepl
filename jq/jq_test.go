package jq_test

import (
	"strings"
	"testing"

	"github.com/ashb/jq-repl/jq"
)

func TestJqNewClose(t *testing.T) {
	jq, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}

	jq.Close()

	// We should be able to safely close multiple times.
	jq.Close()

}

func TestJqCompileError(t *testing.T) {
	jq, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}
	defer jq.Close()

	const program = "a b"
	in, _, errs := jq.Start(program)
	// We aren't sending any input this time.
	close(in)

	// JQ might (and currently does) report multiple errors. One of them will
	// contain our input program. Check for that but don't be overly-specific
	// about the string or order of errors

	gotErrors := false
	for err := range errs {
		gotErrors = true
		if strings.Contains(err.Error(), program) {
			// t.Pass("Found the error we expected: %#v\n",
			return
		}
	}

	if !gotErrors {
		t.Fatal("Errors were expected but none seen")
	}
	t.Fatal("No error containing the program source found")
}

func TestJqSimpleProgram(t *testing.T) {
	state, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing state_state: %v", err)
	}
	defer state.Close()

	input, err := jq.JvFromJSONString("{\"a\": 123}")
	if err != nil {
		t.Error(err)
	}

	in, out, errs := state.Start(".a")

	go func() {
		// We shouldn't see any errors reported. If we do it's an error
		for err := range errs {
			close(in)
			t.Errorf("Expected no errors, but got %#v", err)
		}
	}()

	getAllOutputs := func(out <-chan *jq.Jv) []*jq.Jv {
		var outputs []*jq.Jv
		for jv := range out {
			outputs = append(outputs, jv)
		}
		return outputs
	}

	in <- input
	close(in)

	outputs := getAllOutputs(out)

	if l := len(outputs); l != 1 {
		t.Errorf("Got %d outputs (%#v), expected %d", l, outputs, 1)
	} else if val := outputs[0].ToGoVal(); val != 123 {
		t.Errorf("Got %#v, expected %#v", val, 123)
	}

}
