package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"basura/internal/scan"

	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Title      lipgloss.Style
	Accent     lipgloss.Style
	Muted      lipgloss.Style
	Panel      lipgloss.Style
	Header     lipgloss.Style
	SuccessTxt lipgloss.Style
	WarnTxt    lipgloss.Style
	DangerTxt  lipgloss.Style
	PromptTxt  lipgloss.Style
	Value      lipgloss.Style
}

func NewStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F4A261")),
		Accent: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2A9D8F")),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7A7A7A")),
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#264653")).
			Padding(0, 1),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E9C46A")),
		SuccessTxt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2A9D8F")),
		WarnTxt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E9C46A")),
		DangerTxt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E76F51")),
		PromptTxt: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8AB17D")),
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FAEE")),
	}
}

func (s Styles) RenderBanner() string {
	titleBox := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F1FAEE")).
		Background(lipgloss.Color("#264653")).
		Padding(0, 2).
		Render("GoTrah")

	subtitle := s.Muted.Render("Limpia carpetas pesadas de desarrollo sin pelearte con el disco.")
	accentLine := s.Accent.Render("scan  review  clean")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleBox,
		accentLine,
		subtitle,
	)
}

func (s Styles) RenderScanStatus(spinner string, roots []string, progress scan.Progress) string {
	lines := []string{
		spinner + " " + s.Header.Render("Escaneando"),
		s.Muted.Render("Raices: ") + strings.Join(roots, ", "),
		s.Muted.Render("Visitadas: ") + s.Value.Render(strconv.Itoa(progress.DirsVisited)) +
			s.Muted.Render("  Coincidencias: ") + s.Value.Render(strconv.Itoa(progress.Matches)) +
			s.Muted.Render("  Potencial: ") + s.Value.Render(FormatBytes(progress.BytesMatched)),
	}

	if progress.Current != "" {
		lines = append(lines, s.Muted.Render("Ahora: ")+truncateMiddle(progress.Current, terminalWidth()-10))
	}

	return s.Panel.Render(strings.Join(lines, "\n"))
}

func (s Styles) RenderSummary(result scan.Result) string {
	lines := []string{
		s.Muted.Render("Raices: ") + s.Value.Render(strings.Join(result.Roots, ", ")),
		s.Muted.Render("Carpetas vistas: ") + s.Value.Render(strconv.Itoa(result.DirsVisited)),
		s.Muted.Render("Coincidencias: ") + s.Value.Render(strconv.Itoa(len(result.Candidates))),
		s.Muted.Render("Espacio recuperable: ") + s.Value.Render(FormatBytes(result.TotalBytes)),
	}

	if len(result.Errors) > 0 {
		lines = append(lines, s.Muted.Render("Avisos de acceso: ")+s.WarnTxt.Render(strconv.Itoa(len(result.Errors))))
	}

	return s.Panel.Render(strings.Join(lines, "\n"))
}

func (s Styles) RenderTable(candidates []scan.Candidate) string {
	return s.RenderTablePage(candidates, 0, 0)
}

func (s Styles) RenderTablePage(candidates []scan.Candidate, start int, limit int) string {
	if start < 0 {
		start = 0
	}
	end := len(candidates)
	if limit > 0 && start+limit < end {
		end = start + limit
	}

	width := terminalWidth()
	projectWidth := clamp(width/4, 20, 36)
	pathWidth := clamp(width-projectWidth-24, 28, 72)

	var builder strings.Builder
	if end-start < len(candidates) {
		builder.WriteString(s.Muted.Render(fmt.Sprintf("Mostrando %d-%d de %d coincidencias", start+1, end, len(candidates))))
		builder.WriteString("\n")
	}
	header := fmt.Sprintf(
		"%-4s %-16s %-10s %-*s %s",
		"#",
		"Tipo",
		"Tamano",
		projectWidth,
		"Proyecto",
		"Carpeta",
	)
	builder.WriteString(s.Header.Render(header))
	builder.WriteString("\n")
	builder.WriteString(s.Muted.Render(strings.Repeat("-", min(width-4, 120))))
	builder.WriteString("\n")

	pageCandidates := candidates[start:end]
	for index, candidate := range pageCandidates {
		line := fmt.Sprintf(
			"%-4d %-16s %-10s %-*s %s",
			start+index+1,
			candidate.Name,
			FormatBytes(candidate.SizeBytes),
			projectWidth,
			truncateMiddle(candidate.Project, projectWidth),
			truncateMiddle(candidate.Path, pathWidth),
		)
		builder.WriteString(line)
		if index < len(pageCandidates)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func (s Styles) RenderProjectTable(groups []scan.ProjectGroup, limit int) string {
	if len(groups) == 0 {
		return ""
	}

	end := len(groups)
	if limit > 0 && limit < end {
		end = limit
	}

	width := terminalWidth()
	projectWidth := clamp(width/2, 28, 56)
	typeWidth := clamp(width-projectWidth-24, 18, 36)

	var builder strings.Builder
	builder.WriteString(s.Header.Render("Proyectos mas pesados"))
	builder.WriteString("\n")
	header := fmt.Sprintf(
		"%-4s %-10s %-10s %-*s %s",
		"#",
		"Carpetas",
		"Tamano",
		projectWidth,
		"Proyecto",
		"Tipos",
	)
	builder.WriteString(s.Header.Render(header))
	builder.WriteString("\n")
	builder.WriteString(s.Muted.Render(strings.Repeat("-", min(width-4, 120))))
	builder.WriteString("\n")

	for index, group := range groups[:end] {
		line := fmt.Sprintf(
			"%-4d %-10d %-10s %-*s %s",
			index+1,
			group.Count,
			FormatBytes(group.TotalBytes),
			projectWidth,
			truncateMiddle(group.Project, projectWidth),
			truncateMiddle(strings.Join(group.Types, ", "), typeWidth),
		)
		builder.WriteString(line)
		if index < end-1 {
			builder.WriteString("\n")
		}
	}

	if end < len(groups) {
		builder.WriteString("\n")
		builder.WriteString(s.Muted.Render(fmt.Sprintf("... y %d proyectos mas", len(groups)-end)))
	}

	return builder.String()
}

func (s Styles) RenderTypeTable(groups []scan.TypeGroup, limit int) string {
	if len(groups) == 0 {
		return ""
	}

	end := len(groups)
	if limit > 0 && limit < end {
		end = limit
	}

	var lines []string
	lines = append(lines, s.Header.Render("Tipos de carpeta mas pesados"))
	for _, group := range groups[:end] {
		lines = append(lines, fmt.Sprintf("  %-16s %4d carpetas  %10s", group.Name, group.Count, FormatBytes(group.TotalBytes)))
	}
	if end < len(groups) {
		lines = append(lines, s.Muted.Render(fmt.Sprintf("  ... y %d tipos mas", len(groups)-end)))
	}

	return strings.Join(lines, "\n")
}

func (s Styles) RenderErrors(errorsFound []scan.ScanError, limit int) string {
	if len(errorsFound) == 0 {
		return ""
	}

	if limit <= 0 || limit > len(errorsFound) {
		limit = len(errorsFound)
	}

	lines := []string{s.Warn("Algunas rutas no se pudieron leer:")}
	for _, issue := range errorsFound[:limit] {
		lines = append(lines, s.Muted.Render("- ")+truncateMiddle(issue.Path, terminalWidth()-8))
	}
	if len(errorsFound) > limit {
		lines = append(lines, s.Muted.Render(fmt.Sprintf("... y %d mas", len(errorsFound)-limit)))
	}
	return strings.Join(lines, "\n")
}

func (s Styles) RenderDeleteSummary(okCount int, failCount int, skipCount int, freed int64) string {
	lines := []string{
		s.Muted.Render("Borradas: ") + s.SuccessTxt.Render(strconv.Itoa(okCount)),
		s.Muted.Render("Fallos: ") + s.DangerTxt.Render(strconv.Itoa(failCount)),
		s.Muted.Render("Omitidas: ") + s.WarnTxt.Render(strconv.Itoa(skipCount)),
		s.Muted.Render("Espacio liberado aprox.: ") + s.Value.Render(FormatBytes(freed)),
	}
	return s.Panel.Render(strings.Join(lines, "\n"))
}

func (s Styles) Info(text string) string {
	return s.Accent.Render(text)
}

func (s Styles) Hint(text string) string {
	return s.Muted.Render(text)
}

func (s Styles) Warn(text string) string {
	return s.WarnTxt.Render(text)
}

func (s Styles) Danger(text string) string {
	return s.DangerTxt.Render(text)
}

func (s Styles) Success(text string) string {
	return s.SuccessTxt.Render(text)
}

func (s Styles) Prompt(text string) string {
	return s.PromptTxt.Render(text)
}

func (s Styles) RenderProjectBadge(text string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F1FAEE")).
		Background(lipgloss.Color("#2A9D8F")).
		Padding(0, 1).
		Render(text)
}

func (s Styles) RenderWarningBadge(text string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1D1400")).
		Background(lipgloss.Color("#E9C46A")).
		Padding(0, 1).
		Render(text)
}

func (s Styles) RenderActiveLine(text string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F1FAEE")).
		Background(lipgloss.Color("#264653")).
		Padding(0, 1).
		Render(text)
}

func (s Styles) RenderActiveDetail(text string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DCE9E4")).
		Background(lipgloss.Color("#1E3A43")).
		Padding(0, 1).
		Render(text)
}

func (s Styles) RenderSizeBadge(sizeBytes int64, text string) string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#08131A")).
		Background(lipgloss.Color("#8AB17D")).
		Padding(0, 1)

	switch {
	case sizeBytes >= 10*1024*1024*1024:
		style = style.Foreground(lipgloss.Color("#FFF5F2")).Background(lipgloss.Color("#E76F51"))
	case sizeBytes >= 1024*1024*1024:
		style = style.Foreground(lipgloss.Color("#1D1400")).Background(lipgloss.Color("#E9C46A"))
	}

	return style.Render(text)
}

func FormatBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}

	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(size)
	unit := ""
	for _, current := range units {
		value /= 1024
		unit = current
		if value < 1024 {
			break
		}
	}

	if value >= 100 {
		return fmt.Sprintf("%.0f %s", value, unit)
	}
	if value >= 10 {
		return fmt.Sprintf("%.1f %s", value, unit)
	}
	return fmt.Sprintf("%.2f %s", value, unit)
}

func terminalWidth() int {
	if raw := os.Getenv("COLUMNS"); raw != "" {
		if width, err := strconv.Atoi(raw); err == nil && width > 40 {
			return width
		}
	}
	return 120
}

func truncateMiddle(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	if limit <= 6 {
		return value[:limit]
	}
	head := (limit - 3) / 2
	tail := limit - 3 - head
	return value[:head] + "..." + value[len(value)-tail:]
}

func clamp(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
