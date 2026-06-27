package validator

import (
	"fmt"
	"regexp"
	"strings"
)

type Errors map[string]string

type Validator struct {
	Errors Errors
}

func New() *Validator {
	return &Validator{Errors: make(Errors)}
}

func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v *Validator) Error() string {
	if !v.HasErrors() {
		return ""
	}
	parts := make([]string, 0, len(v.Errors))
	for field, msg := range v.Errors {
		parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(parts, "; ")
}

func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.Errors[field] = fmt.Sprintf("%s is required", field)
	}
}

func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.Errors[field] = fmt.Sprintf("%s must be at least %d characters", field, min)
	}
}

func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.Errors[field] = fmt.Sprintf("%s must not exceed %d characters", field, max)
	}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (v *Validator) Email(field, value string) {
	if !emailRegex.MatchString(value) {
		v.Errors[field] = fmt.Sprintf("%s is not a valid email", field)
	}
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (v *Validator) Slug(field, value string) {
	if !slugRegex.MatchString(value) {
		v.Errors[field] = fmt.Sprintf("%s must be a valid slug (lowercase letters, numbers, hyphens)", field)
	}
}

func (v *Validator) OneOf(field, value string, allowed ...string) {
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.Errors[field] = fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", "))
}

var identRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func (v *Validator) Identifier(field, value string) {
	if !identRegex.MatchString(value) {
		v.Errors[field] = fmt.Sprintf("%s is not a valid identifier", field)
	}
}

func (v *Validator) MinInt(field string, value, min int) {
	if value < min {
		v.Errors[field] = fmt.Sprintf("%s must be at least %d", field, min)
	}
}

func (v *Validator) MaxInt(field string, value, max int) {
	if value > max {
		v.Errors[field] = fmt.Sprintf("%s must not exceed %d", field, max)
	}
}
