package cmd

import (
	"fmt"
	"strings"

	"basura/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type tuiChoiceModel struct {
	title  string
	items  []menuChoice
	cursor int
	styles ui.Styles
}

type tuiChecklistItem struct {
	Label       string
	ProjectName string
	SizeText    string
	SizeBytes   int64
	Detail      string
	Warning     string
	Checked     bool
}

type tuiChecklistModel struct {
	title       string
	items       []tuiChecklistItem
	filtered    []int
	cursor      int
	query       string
	focusSearch bool
	styles      ui.Styles
	visible     int
}

func runChoicePromptTUI(title string, items []menuChoice, styles ui.Styles) (int, error) {
	model := tuiChoiceModel{
		title:  title,
		items:  items,
		styles: styles,
	}

	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return 0, err
	}

	choiceModel, ok := finalModel.(tuiChoiceModel)
	if !ok {
		return 0, nil
	}
	return choiceModel.cursor, nil
}

func runChecklistPromptTUI(title string, items []tuiChecklistItem, styles ui.Styles) ([]int, error) {
	model := tuiChecklistModel{
		title:   title,
		items:   items,
		styles:  styles,
		visible: 10,
	}
	model.refreshFilter()

	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return nil, err
	}

	listModel, ok := finalModel.(tuiChecklistModel)
	if !ok {
		return nil, nil
	}

	indexes := make([]int, 0)
	for index, item := range listModel.items {
		if item.Checked {
			indexes = append(indexes, index)
		}
	}
	return indexes, nil
}

func (m tuiChoiceModel) Init() tea.Cmd {
	return nil
}

func (m tuiChoiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch typed.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Interrupt
		}
	}

	return m, nil
}

func (m tuiChoiceModel) View() string {
	var builder strings.Builder
	builder.WriteString(m.styles.RenderBanner())
	builder.WriteString("\n\n")
	builder.WriteString(m.styles.Header.Render(m.title))
	builder.WriteString("\n")
	builder.WriteString(m.styles.Muted.Render("Usa ↑ ↓ y Enter"))
	builder.WriteString("\n\n")

	for index, item := range m.items {
		prefix := "  "
		lineStyle := m.styles.Value
		if index == m.cursor {
			prefix = m.styles.Accent.Render("› ")
			lineStyle = m.styles.Accent
		}

		builder.WriteString(prefix)
		builder.WriteString(lineStyle.Render(item.Label))
		if item.Hint != "" {
			builder.WriteString("  ")
			builder.WriteString(m.styles.Muted.Render(item.Hint))
		}
		if index < len(m.items)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func (m tuiChecklistModel) Init() tea.Cmd {
	return nil
}

func (m tuiChecklistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		if m.focusSearch {
			switch typed.String() {
			case "esc", "tab":
				m.focusSearch = false
				return m, nil
			case "enter":
				m.focusSearch = false
				return m, nil
			case "backspace":
				if len(m.query) > 0 {
					m.query = string([]rune(m.query)[:len([]rune(m.query))-1])
					m.refreshFilter()
				}
				return m, nil
			case "ctrl+u":
				m.query = ""
				m.refreshFilter()
				return m, nil
			}

			if typed.Type == tea.KeyRunes {
				m.query += string(typed.Runes)
				m.refreshFilter()
			}
			return m, nil
		}

		switch typed.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "/", "tab":
			m.focusSearch = true
		case " ":
			if len(m.filtered) > 0 {
				index := m.filtered[m.cursor]
				m.items[index].Checked = !m.items[index].Checked
			}
		case "a":
			for _, index := range m.filtered {
				m.items[index].Checked = true
			}
		case "n":
			for _, index := range m.filtered {
				m.items[index].Checked = false
			}
		case "enter":
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Interrupt
		}
	}

	return m, nil
}

func (m tuiChecklistModel) View() string {
	var builder strings.Builder
	builder.WriteString(m.styles.RenderBanner())
	builder.WriteString("\n\n")
	builder.WriteString(m.styles.Header.Render(m.title))
	builder.WriteString("\n")
	builder.WriteString(m.styles.Muted.Render("Usa ↑ ↓ para moverte, espacio para marcar, / para buscar, a para marcar visibles, n para limpiar visibles, Enter para confirmar"))
	builder.WriteString("\n")
	builder.WriteString(renderSearchBox(m.styles, m.query, m.focusSearch))
	builder.WriteString("\n")
	builder.WriteString(m.styles.Muted.Render(fmt.Sprintf("Seleccionados: %d de %d  |  Visibles: %d", countChecked(m.items), len(m.items), len(m.filtered))))
	builder.WriteString("\n\n")

	if len(m.filtered) == 0 {
		builder.WriteString(m.styles.Warn("No hay resultados para ese filtro."))
		builder.WriteString("\n")
		builder.WriteString(m.styles.Muted.Render("Pulsa / para editar la busqueda, Backspace para borrar, Esc para volver a la lista."))
		return builder.String()
	}

	start, end := checklistWindow(len(m.filtered), m.cursor, m.visible)
	if start > 0 {
		builder.WriteString(m.styles.Muted.Render(fmt.Sprintf("↑ %d items mas arriba", start)))
		builder.WriteString("\n")
	}

	for index := start; index < end; index++ {
		item := m.items[m.filtered[index]]
		prefix := "  "
		labelStyle := m.styles.Value
		detailStyle := m.styles.Muted
		isActive := index == m.cursor
		if isActive {
			prefix = "› "
			labelStyle = m.styles.Value
			detailStyle = m.styles.Value
		}

		check := m.styles.Muted.Render("[ ]")
		if item.Checked {
			check = m.styles.SuccessTxt.Render("[x]")
		}

		var lineBuilder strings.Builder
		lineBuilder.WriteString(prefix)
		lineBuilder.WriteString(check)
		lineBuilder.WriteString(" ")
		lineBuilder.WriteString(labelStyle.Render(truncateForTUI(item.Label, 72)))
		if item.SizeText != "" {
			lineBuilder.WriteString("  ")
			lineBuilder.WriteString(m.styles.RenderSizeBadge(item.SizeBytes, item.SizeText))
		}
		if item.ProjectName != "" {
			lineBuilder.WriteString("  ")
			lineBuilder.WriteString(m.styles.RenderProjectBadge(truncateForTUI(item.ProjectName, 24)))
		}
		if item.Warning != "" {
			lineBuilder.WriteString("  ")
			lineBuilder.WriteString(m.styles.RenderWarningBadge(truncateForTUI(item.Warning, 24)))
		}

		line := lineBuilder.String()
		if isActive {
			builder.WriteString(m.styles.RenderActiveLine(line))
		} else {
			builder.WriteString(line)
		}
		if item.Detail != "" {
			builder.WriteString("\n")
			detail := "    " + detailStyle.Render(truncateForTUI(item.Detail, 118))
			if isActive {
				builder.WriteString(m.styles.RenderActiveDetail(detail))
			} else {
				builder.WriteString(detail)
			}
		}
		if index < end-1 {
			builder.WriteString("\n")
			builder.WriteString("\n")
		}
	}

	if end < len(m.filtered) {
		builder.WriteString("\n")
		builder.WriteString(m.styles.Muted.Render(fmt.Sprintf("↓ %d items mas abajo", len(m.filtered)-end)))
	}

	return builder.String()
}

func (m *tuiChecklistModel) refreshFilter() {
	query := strings.ToLower(strings.TrimSpace(m.query))
	m.filtered = m.filtered[:0]
	for index, item := range m.items {
		if query == "" {
			m.filtered = append(m.filtered, index)
			continue
		}

		haystack := strings.ToLower(item.Label + " " + item.ProjectName + " " + item.Detail)
		if strings.Contains(haystack, query) {
			m.filtered = append(m.filtered, index)
		}
	}

	if len(m.filtered) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func checklistWindow(total int, cursor int, visible int) (int, int) {
	if total <= visible {
		return 0, total
	}

	start := cursor - (visible / 2)
	if start < 0 {
		start = 0
	}

	end := start + visible
	if end > total {
		end = total
		start = end - visible
	}

	return start, end
}

func countChecked(items []tuiChecklistItem) int {
	count := 0
	for _, item := range items {
		if item.Checked {
			count++
		}
	}
	return count
}

func renderSearchBox(styles ui.Styles, query string, focused bool) string {
	label := "Buscar"
	if focused {
		label = "Buscar (activo)"
	}

	value := query
	if value == "" {
		value = "filtra por proyecto, tipo o ruta"
	}

	style := styles.Muted
	if focused {
		style = styles.Accent
	}

	return style.Render(label + ": " + truncateForTUI(value, 96))
}

func truncateForTUI(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}
