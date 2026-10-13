package services

import (
	"testing"
)

func TestPipelineConfigEquivalent(t *testing.T) {
	// Same operators; the second string carries the provider-injected id and
	// orderId plus different key order and whitespace.
	plan := `[{"type":"add","name":"add-env","enabled":true,"field":"env","value":"production"}]`
	state := `[{"id":"op-1","orderId":1,"name":"add-env","type":"add","field":"env","value":"production","enabled":true}]`

	if !pipelineConfigEquivalent(plan, state) {
		t.Fatalf("normalized configs must compare equal:\n plan:  %s\n state: %s", plan, state)
	}

	changed := `[{"type":"add","name":"add-env","enabled":true,"field":"env","value":"staging"}]`
	if pipelineConfigEquivalent(plan, changed) {
		t.Fatalf("a real value change must not compare equal")
	}

	if pipelineConfigEquivalent(plan, "not json") {
		t.Fatalf("invalid JSON must not compare equal")
	}

	// User-authored ids are kept, not regenerated: an id change is a real diff.
	otherID := `[{"id":"custom","type":"add","name":"add-env","enabled":true,"field":"env","value":"production"}]`
	if pipelineConfigEquivalent(plan, otherID) {
		t.Fatalf("explicit id must not be ignored")
	}
}
