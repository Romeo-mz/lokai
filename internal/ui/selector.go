package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/romeo-mz/lokai/internal/hardware"
	"github.com/romeo-mz/lokai/internal/models"
)

// SelectorResult is returned after the interactive selector completes.
type SelectorResult struct {
	SelectedModel string
	WantInstall   bool
	Cancelled     bool
}

type selectorModel struct {
	recs    []models.Recommendation
	specs   *hardware.HardwareSpecs
	useCase hardware.UseCase
	cursor  int
	chosen  bool // model chosen, now on install confirm
	install bool
	done    bool
	result  SelectorResult
}

func newSelectorModel(recs []models.Recommendation, specs *hardware.HardwareSpecs, useCase hardware.UseCase) selectorModel {
	return selectorModel{
		recs:    recs,
		specs:   specs,
		useCase: useCase,
	}
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.chosen {
			return m.handleInstallKeys(msg)
		}
		return m.handleSelectKeys(msg)
	}
	return m, nil
}

func (m selectorModel) handleSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.recs) { // len(recs) = exit option
			m.cursor++
		}
	case "enter":
		if m.cursor == len(m.recs) {
			// Exit option selected.
			m.done = true
			m.result = SelectorResult{Cancelled: true}
			return m, tea.Quit
		}
		m.chosen = true
	case "q", "esc", "ctrl+c":
		m.done = true
		m.result = SelectorResult{Cancelled: true}
		return m, tea.Quit
	}
	return m, nil
}

func (m selectorModel) handleInstallKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "right", "tab", "h", "l":
		m.install = !m.install
	case "enter":
		m.done = true
		m.result = SelectorResult{
			SelectedModel: m.recs[m.cursor].Model.OllamaTag,
			WantInstall:   m.install,
		}
		return m, tea.Quit
	case "esc":
		m.chosen = false
	case "q", "ctrl+c":
		m.done = true
		m.result = SelectorResult{Cancelled: true}
		return m, tea.Quit
	}
	return m, nil
}

func (m selectorModel) View() string {
	var b strings.Builder

	// Model list.
	b.WriteString(SubtitleStyle.Render("  Which model would you like to use?"))
	b.WriteString("\n\n")

	for i, rec := range m.recs {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(ColorSubtext)
		if i == m.cursor && !m.chosen {
			cursor = SuccessStyle.Render("▸ ")
			style = lipgloss.NewStyle().Foreground(ColorText).Bold(true)
		}
		label := fmt.Sprintf("%s (%s, %.1f GB)", rec.Model.OllamaTag, rec.Model.ParameterSize, rec.Model.EstimatedVRAMGB)
		if !rec.FitsInVRAM {
			label += WarningStyle.Render(" ⚠")
		}
		b.WriteString(cursor + style.Render(label) + "\n")
	}

	// Exit option.
	exitCursor := "  "
	exitStyle := lipgloss.NewStyle().Foreground(ColorSubtext)
	if m.cursor == len(m.recs) && !m.chosen {
		exitCursor = SuccessStyle.Render("▸ ")
		exitStyle = lipgloss.NewStyle().Foreground(ColorText).Bold(true)
	}
	b.WriteString(exitCursor + exitStyle.Render("❌ None — exit without installing") + "\n")

	b.WriteString("\n")

	// Performance panel — only show when a model (not exit) is highlighted.
	if m.cursor < len(m.recs) {
		b.WriteString(m.renderPerformancePanel())
	}

	// Install confirmation.
	if m.chosen {
		b.WriteString(m.renderInstallConfirm())
	} else {
		b.WriteString(MutedStyle.Render("  ↑/↓ navigate • enter select • q quit"))
		b.WriteString("\n")
	}

	return b.String()
}

func (m selectorModel) renderPerformancePanel() string {
	rec := m.recs[m.cursor]
	est := models.EstimatePerformance(rec.Model, m.specs)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary)
	labelStyle := lipgloss.NewStyle().Foreground(ColorSubtext).Width(22)
	valueStyle := lipgloss.NewStyle().Foreground(ColorText).Bold(true)
	taskStyle := lipgloss.NewStyle().Foreground(ColorSuccess).Italic(true)

	var b strings.Builder

	b.WriteString(headerStyle.Render("  📊 Performance Estimate — " + rec.Model.OllamaTag))
	b.WriteString("\n\n")

	b.WriteString("  " + labelStyle.Render("Speed") + valueStyle.Render(fmt.Sprintf("~%.1f tokens/sec", est.TokensPerSecond)) + "\n")
	b.WriteString("  " + labelStyle.Render("First response") + valueStyle.Render(est.TimeToFirstToken) + "\n")
	b.WriteString("  " + labelStyle.Render("Quality") + valueStyle.Render(est.QualityRating) + "\n")

	// Use-case-specific real-world descriptions.
	tasks := models.RealWorldTasks(m.useCase, est.TokensPerSecond)
	if len(tasks) > 0 {
		b.WriteString("\n")
		b.WriteString(headerStyle.Render("  🎯 What this means for you"))
		b.WriteString("\n\n")
		for _, task := range tasks {
			b.WriteString("  " + taskStyle.Render("→ ") + valueStyle.Render(task) + "\n")
		}
	}

	if est.Notes != "" {
		b.WriteString("\n  " + MutedStyle.Render("ℹ "+est.Notes) + "\n")
	}
	b.WriteString("\n")

	return b.String()
}

func (m selectorModel) renderInstallConfirm() string {
	var b strings.Builder

	b.WriteString(SubtitleStyle.Render("  How should we proceed?"))
	b.WriteString("\n\n")

	yesStyle := lipgloss.NewStyle().Foreground(ColorSubtext)
	noStyle := lipgloss.NewStyle().Foreground(ColorSubtext)

	if m.install {
		yesStyle = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	} else {
		noStyle = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	}

	b.WriteString("  " + yesStyle.Render("[ Download & configure now ]"))
	b.WriteString("   ")
	b.WriteString(noStyle.Render("[ Show install instructions only ]"))
	b.WriteString("\n\n")
	b.WriteString(MutedStyle.Render("  ←/→ toggle • enter confirm • esc back"))
	b.WriteString("\n")

	return b.String()
}

// RunModelSelector runs the interactive model selector with live performance estimates.
func RunModelSelector(recs []models.Recommendation, specs *hardware.HardwareSpecs, useCase hardware.UseCase) (SelectorResult, error) {
	m := newSelectorModel(recs, specs, useCase)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return SelectorResult{}, err
	}
	return finalModel.(selectorModel).result, nil
}
