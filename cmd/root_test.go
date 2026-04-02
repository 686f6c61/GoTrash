package cmd

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"basura/internal/scan"
)

func TestParseByteSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  int64
	}{
		{"", 0},
		{"512", 512},
		{"1KB", 1024},
		{"2mb", 2 * 1024 * 1024},
		{"1.5GB", int64(1.5 * 1024 * 1024 * 1024)},
	}

	for _, test := range tests {
		got, err := parseByteSize(test.input)
		if err != nil {
			t.Fatalf("parseByteSize(%q) returned error: %v", test.input, err)
		}
		if got != test.want {
			t.Fatalf("parseByteSize(%q) = %d, want %d", test.input, got, test.want)
		}
	}
}

func TestParseSelection(t *testing.T) {
	t.Parallel()

	got, err := parseSelection("1,3-4,2", 5)
	if err != nil {
		t.Fatalf("parseSelection returned error: %v", err)
	}

	want := []int{0, 1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseSelection = %#v, want %#v", got, want)
	}
}

func TestParseSelectionRejectsOutOfRange(t *testing.T) {
	t.Parallel()

	if _, err := parseSelection("9", 3); err == nil {
		t.Fatal("expected error for out-of-range selection")
	}
}

func TestExportCandidatesCSV(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path, err := exportCandidatesCSV([]scan.Candidate{
		{
			Name:      "node_modules",
			SizeBytes: 1234,
			Project:   "/tmp/project",
			Path:      "/tmp/project/node_modules",
		},
	}, tempDir+"/report.csv")
	if err != nil {
		t.Fatalf("exportCandidatesCSV returned error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "index,type,size_bytes,size_human,project,path") {
		t.Fatalf("csv header missing in %q", text)
	}
	if !strings.Contains(text, "node_modules") {
		t.Fatalf("csv row missing candidate in %q", text)
	}
}
