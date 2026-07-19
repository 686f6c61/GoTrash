package safety

import "testing"

func TestIsProtectedPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{path: "/Library", want: true},
		{path: "/Library/Developer/Xcode", want: true},
		{path: "/opt/homebrew/lib/node_modules", want: true},
		{path: "/private/var/db/something", want: true},
		{path: "/Users/alice/Code/project/node_modules", want: false},
		{path: "/private/tmp/project/node_modules", want: false},
	}

	for _, test := range tests {
		if got := IsProtectedPath(test.path); got != test.want {
			t.Fatalf("IsProtectedPath(%q) = %v, want %v", test.path, got, test.want)
		}
	}
}

func TestGuardReason(t *testing.T) {
	t.Parallel()

	if GuardReason("/System/Library") == "" {
		t.Fatal("expected protected path to return a guard reason")
	}
	if GuardReason("/Users/alice/project") != "" {
		t.Fatal("expected user path to be allowed")
	}
}
