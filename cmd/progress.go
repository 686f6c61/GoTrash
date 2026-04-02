package cmd

import (
	"basura/internal/scan"
	"basura/internal/ui"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type scanProgressMsg struct {
	Progress scan.Progress
}

type scanCompleteMsg struct {
	Result scan.Result
	Err    error
}

type scanModel struct {
	spinner  spinner.Model
	styles   ui.Styles
	roots    []string
	events   <-chan tea.Msg
	progress scan.Progress
	result   scan.Result
	err      error
	done     bool
}

func newScanModel(styles ui.Styles, roots []string, events <-chan tea.Msg) scanModel {
	spin := spinner.New()
	spin.Spinner = spinner.Points
	spin.Style = styles.Accent

	return scanModel{
		spinner: spin,
		styles:  styles,
		roots:   roots,
		events:  events,
	}
}

func (m scanModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitForScanEvent(m.events))
}

func (m scanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		if typed.String() == "ctrl+c" {
			m.err = tea.ErrProgramKilled
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case scanProgressMsg:
		m.progress = typed.Progress
		return m, waitForScanEvent(m.events)
	case scanCompleteMsg:
		m.result = typed.Result
		m.err = typed.Err
		m.done = true
		return m, tea.Quit
	}

	return m, nil
}

func (m scanModel) View() string {
	if m.done {
		return m.styles.Success("Escaneo completado.")
	}

	return m.styles.RenderScanStatus(
		m.spinner.View(),
		m.roots,
		m.progress,
	)
}

func runProgressScan(options scan.Options, styles ui.Styles) (scan.Result, error) {
	if !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stdin.Fd())) {
		return scan.Scan(options)
	}

	events := startScanEvents(options)
	model := newScanModel(styles, options.Roots, events)
	program := tea.NewProgram(model)

	finalModel, err := program.Run()
	if err != nil {
		return scan.Result{}, err
	}

	finished, ok := finalModel.(scanModel)
	if !ok {
		return scan.Result{}, nil
	}

	return finished.result, finished.err
}

func startScanEvents(options scan.Options) <-chan tea.Msg {
	events := make(chan tea.Msg)
	go func() {
		defer close(events)

		opts := options
		opts.Progress = func(progress scan.Progress) {
			events <- scanProgressMsg{Progress: progress}
		}

		result, err := scan.Scan(opts)
		events <- scanCompleteMsg{Result: result, Err: err}
	}()
	return events
}

func waitForScanEvent(events <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-events
		if !ok {
			return nil
		}
		return msg
	}
}
