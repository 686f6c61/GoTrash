package cmd

import (
	"strings"
	"testing"

	"basura/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIChoiceModelMovesCursor(t *testing.T) {
	t.Parallel()

	model := tuiChoiceModel{
		title: "test",
		items: []menuChoice{
			{Label: "one"},
			{Label: "two"},
			{Label: "three"},
		},
		styles: ui.NewStyles(),
	}

	nextModel, _ := model.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune("j")}))
	updated := nextModel.(tuiChoiceModel)
	if updated.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", updated.cursor)
	}
}

func TestTUIChecklistModelToggleAndSelectAll(t *testing.T) {
	t.Parallel()

	model := tuiChecklistModel{
		title: "test",
		items: []tuiChecklistItem{
			{Label: "one"},
			{Label: "two"},
			{Label: "three"},
		},
		styles:  ui.NewStyles(),
		visible: 10,
	}
	model.refreshFilter()

	nextModel, _ := model.Update(tea.KeyMsg(tea.Key{Type: tea.KeySpace}))
	updated := nextModel.(tuiChecklistModel)
	if !updated.items[0].Checked {
		t.Fatal("expected first item to be checked")
	}

	nextModel, _ = updated.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune("a")}))
	updated = nextModel.(tuiChecklistModel)
	for index, item := range updated.items {
		if !item.Checked {
			t.Fatalf("expected item %d to be checked", index)
		}
	}
}

func TestTUIChecklistModelFiltersByQuery(t *testing.T) {
	t.Parallel()

	model := tuiChecklistModel{
		title: "test",
		items: []tuiChecklistItem{
			{Label: "node_modules", Detail: "/tmp/proj-a/node_modules"},
			{Label: ".venv", Detail: "/tmp/proj-b/.venv"},
			{Label: "dist", Detail: "/tmp/proj-a/dist"},
		},
		styles:      ui.NewStyles(),
		visible:     10,
		focusSearch: true,
	}
	model.refreshFilter()

	nextModel, _ := model.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune("proj-b")}))
	updated := nextModel.(tuiChecklistModel)
	if len(updated.filtered) != 1 {
		t.Fatalf("filtered = %d, want 1", len(updated.filtered))
	}
	if updated.filtered[0] != 1 {
		t.Fatalf("filtered[0] = %d, want 1", updated.filtered[0])
	}
}

func TestTUIChecklistModelSelectAllAffectsVisibleOnly(t *testing.T) {
	t.Parallel()

	model := tuiChecklistModel{
		title: "test",
		items: []tuiChecklistItem{
			{Label: "node_modules", Detail: "/tmp/proj-a/node_modules"},
			{Label: ".venv", Detail: "/tmp/proj-b/.venv"},
			{Label: "dist", Detail: "/tmp/proj-a/dist"},
		},
		styles:      ui.NewStyles(),
		visible:     10,
		query:       "proj-a",
		focusSearch: false,
	}
	model.refreshFilter()

	nextModel, _ := model.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune("a")}))
	updated := nextModel.(tuiChecklistModel)

	if !updated.items[0].Checked || !updated.items[2].Checked {
		t.Fatal("expected visible items to be checked")
	}
	if updated.items[1].Checked {
		t.Fatal("expected hidden item to remain unchecked")
	}
}

func TestChecklistWindow(t *testing.T) {
	t.Parallel()

	start, end := checklistWindow(100, 50, 14)
	if start < 0 || end > 100 || end-start != 14 {
		t.Fatalf("unexpected window: %d-%d", start, end)
	}
}

func TestTUIChecklistViewIncludesDetailInList(t *testing.T) {
	t.Parallel()

	model := tuiChecklistModel{
		title: "test",
		items: []tuiChecklistItem{
			{
				Label:       "#1  node_modules  1.2 GB",
				ProjectName: "project",
				SizeText:    "1.2 GB",
				SizeBytes:   1288490188,
				Detail:      "/Users/demo/project  ->  /Users/demo/project/node_modules",
				Warning:     "protegida",
			},
		},
		styles:  ui.NewStyles(),
		visible: 10,
	}
	model.refreshFilter()

	view := model.View()
	if !strings.Contains(view, "/Users/demo/project/node_modules") {
		t.Fatalf("expected detail to appear in checklist view, got %q", view)
	}
	if !strings.Contains(view, "project") {
		t.Fatalf("expected project name to appear in checklist view, got %q", view)
	}
	if !strings.Contains(strings.ToLower(view), "protegida") {
		t.Fatalf("expected warning badge to appear in checklist view, got %q", view)
	}
}
