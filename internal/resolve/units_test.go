package resolve

import (
	"testing"
)

func TestRemToPx(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1rem", "16px"},
		{"1.5rem", "24px"},
		{"0.25rem", "4px"},
		{"16px", "16px"},
		{"1rem 2rem", "16px 32px"},
		{"calc(1rem + 2rem)", "calc(16px + 32px)"},
		{"9999px", "9999px"},
		{"0.75rem", "12px"},
	}
	for _, tt := range tests {
		got := RemToPx(tt.input, 16)
		if got != tt.want {
			t.Errorf("RemToPx(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStripUnit(t *testing.T) {
	tests := []struct {
		input  string
		wantV  float64
		wantU  string
		wantOK bool
	}{
		{"16px", 16, "px", true},
		{"1.5rem", 1.5, "rem", true},
		{"400", 400, "", true},
		{"auto", 0, "", false},
	}
	for _, tt := range tests {
		v, u, ok := StripUnit(tt.input)
		if ok != tt.wantOK || v != tt.wantV || u != tt.wantU {
			t.Errorf("StripUnit(%q) = (%v, %q, %v), want (%v, %q, %v)",
				tt.input, v, u, ok, tt.wantV, tt.wantU, tt.wantOK)
		}
	}
}
