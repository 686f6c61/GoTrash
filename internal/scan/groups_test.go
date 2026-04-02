package scan

import "testing"

func TestGroupByProject(t *testing.T) {
	t.Parallel()

	candidates := []Candidate{
		{Name: "node_modules", Project: "/tmp/a", SizeBytes: 300},
		{Name: ".venv", Project: "/tmp/b", SizeBytes: 200},
		{Name: "dist", Project: "/tmp/a", SizeBytes: 100},
	}

	groups := GroupByProject(candidates)
	if len(groups) != 2 {
		t.Fatalf("GroupByProject returned %d groups, want 2", len(groups))
	}
	if groups[0].Project != "/tmp/a" {
		t.Fatalf("first group project = %q, want /tmp/a", groups[0].Project)
	}
	if groups[0].Count != 2 || groups[0].TotalBytes != 400 {
		t.Fatalf("unexpected first group stats: %#v", groups[0])
	}
}

func TestGroupByType(t *testing.T) {
	t.Parallel()

	candidates := []Candidate{
		{Name: "node_modules", SizeBytes: 300},
		{Name: "node_modules", SizeBytes: 200},
		{Name: ".venv", SizeBytes: 100},
	}

	groups := GroupByType(candidates)
	if len(groups) != 2 {
		t.Fatalf("GroupByType returned %d groups, want 2", len(groups))
	}
	if groups[0].Name != "node_modules" || groups[0].Count != 2 || groups[0].TotalBytes != 500 {
		t.Fatalf("unexpected first type group: %#v", groups[0])
	}
}
