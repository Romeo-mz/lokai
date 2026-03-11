package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/romeo-mz/lokai/internal/benchmark"
	"github.com/romeo-mz/lokai/internal/cache"
	"github.com/romeo-mz/lokai/internal/ollama"
)

const (
	benchCacheKey = "benchmarks"
	benchCacheTTL = 6 * time.Hour
)

// cachedBenchmarks is the on-disk shape of cached benchmark results.
type cachedBenchmarks struct {
	Results []benchmark.Result `json:"results"`
}

// RunBenchmark benchmarks installed models, using cached results when available.
// A model is re-benchmarked only when:
//   - it has never been benchmarked, or
//   - the --benchmark flag was explicitly passed (force=true), or
//   - the entire cache has expired (6 h TTL).
func RunBenchmark(ctx context.Context, client *ollama.Client) error {
	return runBenchmarkInner(ctx, client, true) // explicit flag = force refresh
}

// RunBenchmarkCached returns cached results if available, only benchmarking
// models that are missing from the cache.
func RunBenchmarkCached(ctx context.Context, client *ollama.Client) ([]benchmark.Result, error) {
	store, err := cache.New()
	if err != nil {
		return nil, err
	}

	installedModels, err := client.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	if len(installedModels) == 0 {
		return nil, nil
	}

	var installedTags []string
	for _, m := range installedModels {
		installedTags = append(installedTags, m.Name)
	}

	// Load cache.
	var cached cachedBenchmarks
	store.Get(benchCacheKey, &cached)

	// Find which models are already benchmarked.
	have := make(map[string]benchmark.Result)
	for _, r := range cached.Results {
		if r.Success {
			have[r.ModelTag] = r
		}
	}

	// Determine which ones need benchmarking.
	var missing []string
	for _, tag := range installedTags {
		if _, ok := have[tag]; !ok {
			missing = append(missing, tag)
		}
	}

	// Benchmark only missing models.
	if len(missing) > 0 {
		opts := benchmark.Options{Warmup: true, MaxTokens: 128}
		newResults := benchmark.RunMultiple(ctx, client, missing, opts, func(r benchmark.Result) {
			status := SuccessStyle.Render("✓")
			if !r.Success {
				status = ErrorStyle.Render("✗")
			}
			fmt.Printf("   %s %s — %s\n", status, ValueStyle.Render(r.ModelTag), r.FormattedSpeed())
		})
		for _, r := range newResults {
			if r.Success {
				have[r.ModelTag] = r
			}
		}
	}

	// Build final list (only models currently installed).
	var results []benchmark.Result
	for _, tag := range installedTags {
		if r, ok := have[tag]; ok {
			results = append(results, r)
		}
	}

	// Save to cache.
	_ = store.Set(benchCacheKey, cachedBenchmarks{Results: results}, benchCacheTTL)

	return results, nil
}

// LoadBenchmarkCache reads the on-disk benchmark cache without running any benchmarks.
// Returns a map of model tag → measured tokens/sec for all successfully cached models.
func LoadBenchmarkCache() map[string]float64 {
	store, err := cache.New()
	if err != nil {
		return nil
	}
	var cached cachedBenchmarks
	if !store.Get(benchCacheKey, &cached) {
		return nil
	}
	result := make(map[string]float64, len(cached.Results))
	for _, r := range cached.Results {
		if r.Success && r.EvalRate > 0 {
			result[r.ModelTag] = r.EvalRate
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func runBenchmarkInner(ctx context.Context, client *ollama.Client, force bool) error {
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

	store, _ := cache.New()

	// Check cache unless forced.
	var cached cachedBenchmarks
	if !force && store != nil && store.Get(benchCacheKey, &cached) && len(cached.Results) > 0 {
		// Show cached results if they cover all installed models.
		have := make(map[string]bool)
		for _, r := range cached.Results {
			have[r.ModelTag] = true
		}
		allCovered := true
		for _, tag := range tags {
			if !have[tag] {
				allCovered = false
				break
			}
		}
		if allCovered {
			fmt.Println(MutedStyle.Render("   Using cached results (run --benchmark to refresh)"))
			fmt.Println()
			displayBenchResults(cached.Results)
			return nil
		}
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

	// Save to cache.
	if store != nil {
		_ = store.Set(benchCacheKey, cachedBenchmarks{Results: results}, benchCacheTTL)
	}

	fmt.Println()
	displayBenchResults(results)
	return nil
}

func displayBenchResults(results []benchmark.Result) {
	// Sort by speed descending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].EvalRate > results[j].EvalRate
	})

	var rows [][]string
	for _, r := range results {
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
	}

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

	// Show cache info.
	fmt.Println(MutedStyle.Render(strings.Repeat(" ", 3) + "Results cached for 6 hours. Run --benchmark to force refresh."))
	fmt.Println()
}
