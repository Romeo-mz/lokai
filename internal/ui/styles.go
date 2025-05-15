package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// ═══════════════════════════════════════════
// Color palette
// ═══════════════════════════════════════════

var (
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#06B6D4") // Cyan
	ColorSuccess   = lipgloss.Color("#10B981") // Green
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorError     = lipgloss.Color("#EF4444") // Red
	ColorMuted     = lipgloss.Color("#6B7280") // Gray
	ColorText      = lipgloss.Color("#F9FAFB") // Near-white
	ColorSubtext   = lipgloss.Color("#9CA3AF") // Light gray
	ColorBg        = lipgloss.Color("#1F2937") // Dark bg
)

// ═══════════════════════════════════════════
// Text styles
// ═══════════════════════════════════════════

var (
	// Title is the main header style.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	// Subtitle for section headers.
	SubtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary)

	// Success message.
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	// Warning message.
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	// Error message.
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorError)

	// Muted/dimmed text.
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Label for key-value displays.
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Width(20)

	// Value for key-value displays.
	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true)
)

// ═══════════════════════════════════════════
// Box styles
// ═══════════════════════════════════════════

var (
	// BoxStyle for bordered boxes.
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	// ResultBoxStyle for recommendation results.
	ResultBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(1, 2).
			MarginTop(1)

	// WarningBoxStyle for warnings.
	WarningBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarning).
			Padding(0, 2)
)

// ═══════════════════════════════════════════
// Banner
// ═══════════════════════════════════════════

const Logo = `
 ██╗      ██████╗ ██╗  ██╗ █████╗ ██╗
 ██║     ██╔═══██╗██║ ██╔╝██╔══██╗██║
 ██║     ██║   ██║█████╔╝ ███████║██║
 ██║     ██║   ██║██╔═██╗ ██╔══██║██║
 ███████╗╚██████╔╝██║  ██╗██║  ██║██║
 ╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝`

var LogoStyle = lipgloss.NewStyle().
	Foreground(ColorPrimary).
	Bold(true)

var TaglineStyle = lipgloss.NewStyle().
	Foreground(ColorSubtext).
	Italic(true).
	MarginBottom(1)

// Banner returns the formatted application banner.
func Banner() string {
	return LogoStyle.Render(Logo) + "\n" +
		TaglineStyle.Render("  Find the best local AI model for your hardware") + "\n"
}
