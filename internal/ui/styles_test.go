package ui

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		size int64
		want string
	}{
		{size: 12, want: "12 B"},
		{size: 1024, want: "1.00 KB"},
		{size: 10 * 1024 * 1024, want: "10.0 MB"},
		{size: 150 * 1024 * 1024, want: "150 MB"},
	}

	for _, test := range tests {
		if got := FormatBytes(test.size); got != test.want {
			t.Fatalf("FormatBytes(%d) = %q, want %q", test.size, got, test.want)
		}
	}
}

func TestTerminalWidthUsesColumnsWhenValid(t *testing.T) {
	t.Setenv("COLUMNS", "160")

	if got := terminalWidth(); got != 160 {
		t.Fatalf("terminalWidth() = %d, want 160", got)
	}
}

func TestTerminalWidthFallsBackWhenColumnsInvalid(t *testing.T) {
	t.Setenv("COLUMNS", "10")

	if got := terminalWidth(); got != 120 {
		t.Fatalf("terminalWidth() = %d, want 120", got)
	}
}

func TestTruncateMiddle(t *testing.T) {
	t.Parallel()

	got := truncateMiddle("/Users/alice/very/long/path/node_modules", 18)
	if got != "/Users/..._modules" {
		t.Fatalf("truncateMiddle returned %q", got)
	}
}

func TestClamp(t *testing.T) {
	t.Parallel()

	if got := clamp(5, 10, 20); got != 10 {
		t.Fatalf("clamp lower bound = %d, want 10", got)
	}
	if got := clamp(30, 10, 20); got != 20 {
		t.Fatalf("clamp upper bound = %d, want 20", got)
	}
	if got := clamp(15, 10, 20); got != 15 {
		t.Fatalf("clamp middle = %d, want 15", got)
	}
}
