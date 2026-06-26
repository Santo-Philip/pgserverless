package utils

import (
	"testing"
)

func TestValidatorRequired(t *testing.T) {
	v := NewValidator()
	v.Required("name", "")
	if !v.HasErrors() {
		t.Error("expected validation error for empty value")
	}
}

func TestValidatorRequiredValid(t *testing.T) {
	v := NewValidator()
	v.Required("name", "test")
	if v.HasErrors() {
		t.Errorf("unexpected validation error: %s", v.Error())
	}
}

func TestValidatorEmail(t *testing.T) {
	v := NewValidator()
	v.Email("email", "invalid")
	if !v.HasErrors() {
		t.Error("expected validation error for invalid email")
	}
}

func TestValidatorEmailValid(t *testing.T) {
	v := NewValidator()
	v.Email("email", "test@example.com")
	if v.HasErrors() {
		t.Errorf("unexpected validation error: %s", v.Error())
	}
}

func TestValidatorMinLength(t *testing.T) {
	v := NewValidator()
	v.MinLength("field", "ab", 3)
	if !v.HasErrors() {
		t.Error("expected validation error for short value")
	}
}

func TestValidatorMaxLength(t *testing.T) {
	v := NewValidator()
	v.MaxLength("field", "abcdef", 3)
	if !v.HasErrors() {
		t.Error("expected validation error for long value")
	}
}

func TestValidatorSlug(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"my-app", true},
		{"my_app", true},
		{"myapp123", true},
		{"MyApp", false},
		{"my app", false},
		{"", true},
	}
	for _, tc := range tests {
		v := NewValidator()
		v.Slug("slug", tc.input)
		if tc.valid && v.HasErrors() {
			t.Errorf("expected %q to be valid, got error: %s", tc.input, v.Error())
		}
		if !tc.valid && !v.HasErrors() {
			t.Errorf("expected %q to be invalid", tc.input)
		}
	}
}

func TestValidatorOneOf(t *testing.T) {
	v := NewValidator()
	v.OneOf("type", "invalid", "a", "b", "c")
	if !v.HasErrors() {
		t.Error("expected validation error for invalid oneof")
	}
}

func TestValidatorOneOfValid(t *testing.T) {
	v := NewValidator()
	v.OneOf("type", "b", "a", "b", "c")
	if v.HasErrors() {
		t.Errorf("unexpected validation error: %s", v.Error())
	}
}
