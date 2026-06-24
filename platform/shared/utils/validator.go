package utils

import (
	"net/mail"
	"strings"
	"unicode"
)

type ValidationErrors map[string]string

type Validator struct {
	Errors ValidationErrors
}

func NewValidator() *Validator {
	return &Validator{Errors: make(ValidationErrors)}
}

func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v *Validator) Error() string {
	if !v.HasErrors() {
		return ""
	}
	var msgs []string
	for field, msg := range v.Errors {
		msgs = append(msgs, field+": "+msg)
	}
	return strings.Join(msgs, "; ")
}

func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.Errors[field] = "is required"
	}
}

func (v *Validator) Email(field, value string) {
	if value == "" {
		return
	}
	_, err := mail.ParseAddress(value)
	if err != nil {
		v.Errors[field] = "must be a valid email address"
	}
}

func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.Errors[field] = "must be at least " + itoa(min) + " characters"
	}
}

func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.Errors[field] = "must not exceed " + itoa(max) + " characters"
	}
}

func (v *Validator) Slug(field, value string) {
	if value == "" {
		return
	}
	for _, r := range value {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			v.Errors[field] = "must contain only lowercase letters, digits, hyphens, and underscores"
			return
		}
	}
}

func (v *Validator) OneOf(field, value string, allowed ...string) {
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.Errors[field] = "must be one of: " + strings.Join(allowed, ", ")
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
