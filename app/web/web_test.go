package web

import (
	"testing"
)

func TestTemplates(t *testing.T) {
	tpl, err := Templates()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if tpl == nil {
		t.Error("expected template but got nil")
	}

	// Verify that templates are loaded by checking if we can get template names
	templates := tpl.Templates()
	if len(templates) == 0 {
		t.Error("expected templates but got none")
	}
}

func TestStatic(t *testing.T) {
	staticFS, err := Static()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if staticFS == nil {
		t.Error("expected filesystem but got nil")
	}
}
