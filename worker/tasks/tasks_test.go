package tasks

import (
	"testing"
)

func TestTaskTypes(t *testing.T) {
	if TypeSchemaRefresh != "schema:refresh" {
		t.Errorf("expected schema:refresh, got %s", TypeSchemaRefresh)
	}
}
