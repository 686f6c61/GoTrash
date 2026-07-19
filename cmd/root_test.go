package cmd

import (
	"bufio"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"basura/internal/scan"
	"basura/internal/ui"
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

func TestBuildDeletionPlanSeparatesProtectedCandidates(t *testing.T) {
	t.Parallel()

	candidates := []scan.Candidate{
		{
			Name:      "node_modules",
			Project:   "/Users/alice/project",
			Path:      "/Users/alice/project/node_modules",
			SizeBytes: 25,
		},
		{
			Name:      "node_modules",
			Project:   "/opt/homebrew/lib",
			Path:      "/opt/homebrew/lib/node_modules",
			SizeBytes: 50,
		},
	}

	plan := buildDeletionPlan(candidates)
	if len(plan.Actionable) != 1 {
		t.Fatalf("buildDeletionPlan actionable = %d, want 1", len(plan.Actionable))
	}
	if len(plan.SkippedResults) != 1 {
		t.Fatalf("buildDeletionPlan skipped = %d, want 1", len(plan.SkippedResults))
	}
	if plan.ActionableBytes != 25 {
		t.Fatalf("buildDeletionPlan actionable bytes = %d, want 25", plan.ActionableBytes)
	}
}

func TestConfirmDeletionUsesActionableSelection(t *testing.T) {
	t.Parallel()

	plan := buildDeletionPlan([]scan.Candidate{
		{
			Name:      "node_modules",
			Project:   "/Users/alice/project",
			Path:      "/Users/alice/project/node_modules",
			SizeBytes: 25,
		},
		{
			Name:      "node_modules",
			Project:   "/opt/homebrew/lib",
			Path:      "/opt/homebrew/lib/node_modules",
			SizeBytes: 50,
		},
	})

	reader := bufio.NewReader(strings.NewReader("BORRAR\n"))
	if err := confirmDeletion(plan, ui.NewStyles(), reader); err != nil {
		t.Fatalf("confirmDeletion returned error: %v", err)
	}
}

func TestDeleteCandidatesPreservesSkippedResults(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	target := tempDir + "/node_modules"
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	plan := buildDeletionPlan([]scan.Candidate{
		{
			Name:      "node_modules",
			Project:   tempDir,
			Path:      target,
			SizeBytes: 25,
		},
		{
			Name:      "node_modules",
			Project:   "/opt/homebrew/lib",
			Path:      "/opt/homebrew/lib/node_modules",
			SizeBytes: 50,
		},
	})

	results, freed := deleteCandidates(plan)
	if len(results) != 2 {
		t.Fatalf("deleteCandidates results = %d, want 2", len(results))
	}
	if freed != 25 {
		t.Fatalf("deleteCandidates freed = %d, want 25", freed)
	}
	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected %q to be deleted, stat err = %v", target, err)
	}
}
