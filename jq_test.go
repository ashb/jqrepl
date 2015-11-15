package jqrepl

import "testing"

func TestNewClose(t *testing.T) {
	jq, err := New()

	if err != nil {
		t.Errorf("Error initializing jq_state: %q", err)
	}

	jq.Close()
	if jq._state != nil {
		t.Error("Expected jq._state to be nil after Close")
	}
	jq.Close()

}
