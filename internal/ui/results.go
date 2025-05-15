package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/romeo-mz/lokai/internal/hardware"
	"github.com/romeo-mz/lokai/internal/models"
)

// DisplayResults shows the recommendation table and returns the user's choice.
func DisplayResults(recs []models.Recommendation, specs *hardware.HardwareSpecs) (string, bool, error) {
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

		rows = append(rows, []string{
			rank,
			rec.Model.OllamaTag,
			rec.Model.ParameterSize,
			vramStr,
			fmt.Sprintf("~%.0f tok/s", est.TokensPerSecond),
			est.QualityRating,
			rec.Reason,
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

	// Show performance details for top recommendation.
	if len(recs) > 0 {
		topEst := models.EstimatePerformance(recs[0].Model, specs)
		fmt.Println(SubtitleStyle.Render("📊 Performance Estimate for " + recs[0].Model.OllamaTag))
		fmt.Printf("   Speed:          ~%.1f tokens/second\n", topEst.TokensPerSecond)
		fmt.Printf("   First token:    %s\n", topEst.TimeToFirstToken)
		fmt.Printf("   Typical reply:  %s\n", topEst.GenerationTime)
		if topEst.Notes != "" {
			fmt.Printf("   Notes:          %s\n", MutedStyle.Render(topEst.Notes))
		}
		fmt.Println()
	}

	// Ask user to pick a model.
	var modelOptions []huh.Option[string]
	for _, rec := range recs {
		label := fmt.Sprintf("%s (%s, %.1f GB)", rec.Model.OllamaTag, rec.Model.ParameterSize, rec.Model.EstimatedVRAMGB)
		modelOptions = append(modelOptions, huh.NewOption(label, rec.Model.OllamaTag))
	}
	modelOptions = append(modelOptions, huh.NewOption("❌ None — exit without installing", ""))

	var selectedModel string
	var wantInstall bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which model would you like to use?").
				Options(modelOptions...).
				Value(&selectedModel),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("How should we proceed?").
				Affirmative("Download & configure now").
				Negative("Show install instructions only").
				Value(&wantInstall),
		).WithHideFunc(func() bool {
			return selectedModel == ""
		}),
	)

	err := form.Run()
	if err != nil {
		return "", false, err
	}

	return selectedModel, wantInstall, nil
}
