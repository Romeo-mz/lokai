package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/romeo-mz/lokai/internal/benchmark"
	"github.com/romeo-mz/lokai/internal/ollama"
)

// RunBenchmark benchmarks all installed models and displays results.
func RunBenchmark(ctx context.Context, client *ollama.Client) error {
	fmt.Println(SubtitleStyle.Render("⚡ Benchmarking installed models..."))
	fmt.Println()

	models, err := client.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		fmt.Println(WarningStyle.Render("⚠ No models installed. Pull a model first with: ollama pull <model>"))
		return nil
	}

	var tags []string
	for _, m := range models {
		tags = append(tags, m.Name)
	}

	fmt.Printf("   Found %s model(s) to benchmark\n\n", ValueStyle.Render(fmt.Sprintf("%d", len(tags))))

	opts := benchmark.Options{
		Warmup:    true,
		MaxTokens: 128,
	}

	var rows [][]string
	results := benchmark.RunMultiple(ctx, client, tags, opts, func(r benchmark.Result) {
		status := SuccessStyle.Render("✓")
		if !r.Success {
			status = ErrorStyle.Render("✗")
		}
		fmt.Printf("   %s %s — %s\n", status, ValueStyle.Render(r.ModelTag), r.FormattedSpeed())

		row := []string{
			r.ModelTag,
			r.FormattedSpeed(),
			r.FormattedTTFT(),
			fmt.Sprintf("%d", r.TokensGenerated),
			r.FormattedTotal(),
		}
		if !r.Success {
			row = []string{r.ModelTag, "failed", "—", "—", r.Error}
		}
		rows = append(rows, row)
	})

	fmt.Println()
	fmt.Println(SubtitleStyle.Render("📊 Benchmark Results"))
	fmt.Println()

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorSuccess)).
		Headers("Model", "Speed", "First Token", "Tokens", "Total Time").
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

	// Show fastest model.
	var fastest benchmark.Result
	for _, r := range results {
		if r.Success && (fastest.EvalRate == 0 || r.EvalRate > fastest.EvalRate) {
			fastest = r
		}
	}
	if fastest.Success {
		fmt.Printf("   🏆 Fastest: %s at %s\n\n",
			ValueStyle.Render(fastest.ModelTag),
			SuccessStyle.Render(fastest.FormattedSpeed()),
		)
	}

	return nil
}
