package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/romeo-mz/lokai/internal/hardware"
	"github.com/romeo-mz/lokai/internal/models"
)

// DisplayResults shows the recommendation table, then runs the interactive
// model selector with live performance estimates. Returns the chosen model tag,
// whether the user wants to install, and any error.
// benchData, when non-nil, overrides estimated tok/s with measured values in the table and selector.
func DisplayResults(recs []models.Recommendation, specs *hardware.HardwareSpecs, useCase hardware.UseCase, benchData map[string]float64) (string, bool, error) {
	if len(recs) == 0 {
		fmt.Println(WarningBoxStyle.Render(
			WarningStyle.Render("⚠ No compatible models found for your hardware and use case.\n") +
				MutedStyle.Render("Try selecting a different use case or lowering the quality priority."),
		))
		return "", false, nil
	}

	fmt.Println(SubtitleStyle.Render("🏆 Recommended Models"))
	fmt.Println()

	// Build results table.
	var rows [][]string
	for i, rec := range recs {
		rank := fmt.Sprintf("#%d", i+1)

		vramStr := fmt.Sprintf("%.1f GB", rec.Model.EstimatedVRAMGB)
		if !rec.FitsInVRAM {
			vramStr = WarningStyle.Render(vramStr + " ⚠")
		}

		est := models.EstimatePerformance(rec.Model, specs)
		// If we have an actual benchmark measurement, use it in the table.
		speedLabel := fmt.Sprintf("~%.0f tok/s", est.TokensPerSecond)
		if benchData != nil {
			if tps, ok := benchData[rec.Model.OllamaTag]; ok && tps > 0 {
				speedLabel = fmt.Sprintf("✓%.0f tok/s", tps)
			}
		}

		modelDisplay := rec.Model.OllamaTag
		installNote := rec.Reason
		if !rec.Model.IsPullable() {
			modelDisplay = rec.Model.OllamaTag
			installNote = WarningStyle.Render("⚠ Manual setup via " + rec.Model.Pipeline + " — cannot be pulled with ollama")
		}
		rows = append(rows, []string{
			rank,
			modelDisplay,
			rec.Model.ParameterSize,
			vramStr,
			speedLabel,
			est.QualityRating,
			installNote,
		})
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorSuccess)).
		Headers("Rank", "Model", "Size", "VRAM Needed", "Est. Speed", "Quality", "Notes").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().
					Bold(true).
					Foreground(ColorSuccess).
					Padding(0, 1)
			}
			style := lipgloss.NewStyle().Padding(0, 1)
			if col == 0 {
				style = style.Foreground(ColorPrimary).Bold(true)
			}
			return style
		})

	fmt.Println(t)
	fmt.Println()

	// Interactive selector with live performance panel.
	result, err := RunModelSelector(recs, specs, useCase, benchData)
	if err != nil {
		return "", false, err
	}

	if result.Cancelled {
		return "", false, nil
	}

	return result.SelectedModel, result.WantInstall, nil
}
