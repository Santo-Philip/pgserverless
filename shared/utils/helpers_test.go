package utils

import (
	"testing"
)

func TestErrNoRowsAsNil(t *testing.T) {
	if err := ErrNoRowsAsNil(nil); err != nil {
		t.Fatal("expected nil to stay nil")
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{123, "123"},
		{-1, "-1"},
	}
	for _, tc := range tests {
		got := itoa(tc.input)
		if got != tc.want {
			t.Errorf("itoa(%d) = %s, want %s", tc.input, got, tc.want)
		}
	}
}
