package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/romeo-mz/lokai/internal/benchmark"
	"github.com/romeo-mz/lokai/internal/cache"
	"github.com/romeo-mz/lokai/internal/hardware"
	"github.com/romeo-mz/lokai/internal/models"
	"github.com/romeo-mz/lokai/internal/ollama"
	"github.com/romeo-mz/lokai/internal/ui"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("lokai", version)
		return
	}

	wantClean := len(os.Args) > 1 && (os.Args[1] == "--clean" || os.Args[1] == "clean")
	wantBenchmark := len(os.Args) > 1 && (os.Args[1] == "--benchmark" || os.Args[1] == "benchmark")
	wantClearCache := len(os.Args) > 1 && (os.Args[1] == "--clear-cache" || os.Args[1] == "clear-cache")

	ctx := context.Background()

	// Handle --clear-cache: wipe all local cached data.
	if wantClearCache {
		store, err := cache.New()
		if err == nil {
			_ = store.Clear()
			fmt.Println("✓ Cache cleared (" + store.Dir() + ")")
		}
		return
	}

	// Print banner.
	fmt.Println(ui.Banner())

	// Check Ollama installation.
	status := ollama.CheckInstallation()
	if !status.Installed {
		fmt.Println(ui.ErrorStyle.Render("✗ " + status.ErrorMessage))
		fmt.Println()
		fmt.Println(ollama.GetInstallInstructions())
		os.Exit(1)
	}

	// Create Ollama client and check health.
	client, err := ollama.NewClient()
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ " + err.Error()))
		os.Exit(1)
	}

	if err := client.CheckHealth(ctx); err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ Ollama server is not running."))
		fmt.Println(ui.MutedStyle.Render("  Start it with: ollama serve"))
		os.Exit(1)
	}

	// Handle --clean: remove all installed models.
	if wantClean {
		if err := ui.RunCleanup(ctx, client); err != nil {
			fmt.Println(ui.ErrorStyle.Render("✗ " + err.Error()))
			os.Exit(1)
		}
		return
	}

	// Handle --benchmark: benchmark all installed models.
	if wantBenchmark {
		if err := ui.RunBenchmark(ctx, client); err != nil {
			fmt.Println(ui.ErrorStyle.Render("✗ " + err.Error()))
			os.Exit(1)
		}
		return
	}

	// Scan hardware.
	specs, err := ui.ScanAndDisplay(ctx)
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ " + err.Error()))
		os.Exit(1)
	}

	// Run questionnaire.
	prefs, err := ui.RunQuestionnaire(specs)
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ " + err.Error()))
		os.Exit(1)
	}

	// Refresh model registry from Ollama + GitHub (best-effort, 5s timeout).
	reg := models.NewRegistry()
	cachedCount := reg.DynamicCount()
	refreshCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	refreshErr := reg.Refresh(refreshCtx)
	cancel()

	if refreshErr == nil && reg.DynamicCount() > 0 {
		if cachedCount > 0 && cachedCount == reg.DynamicCount() {
			fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  ✓ %d additional models available (cached)", reg.DynamicCount())))
		} else {
			fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  ✓ Discovered %d additional models from Ollama & GitHub", reg.DynamicCount())))
		}
		fmt.Println()
	} else if cachedCount > 0 {
		fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  ✓ %d additional models available (offline, using cache)", cachedCount)))
		fmt.Println()
	}

	// Get recommendations (uses dynamic catalog if available, else static).
	var catalog []models.ModelEntry
	if reg.DynamicCount() > 0 {
		catalog = reg.FullCatalog()
	}
	recs := models.Recommend(specs, models.RecommendOptions{
		UseCase:       prefs.UseCase,
		SubTask:       prefs.SubTask,
		Priority:      prefs.Priority,
		IncludeRemote: prefs.IncludeRemote,
		MaxResults:    5,
		Catalog:       catalog,
	})

	// Load real performance baselines from benchmark cache (if available).
	if store, err := cache.New(); err == nil {
		var cached struct {
			Results []benchmark.Result `json:"results"`
		}
		if store.Get("benchmarks", &cached) && len(cached.Results) > 0 {
			bl := make(map[string]models.Baseline)
			for _, r := range cached.Results {
				if r.Success {
					bl[r.ModelTag] = models.Baseline{
						TokensPerSecond:  r.EvalRate,
						TimeToFirstToken: r.TimeToFirstToken,
					}
				}
			}
			if len(bl) > 0 {
				models.SetBaselines(bl)
			}
		}
	}

	// Display results and get user selection.
	selectedModel, wantInstall, err := ui.DisplayResults(recs, specs, prefs.UseCase)
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ " + err.Error()))
		os.Exit(1)
	}

	// Show context-specific notes.
	if prefs.UseCase == hardware.UseCaseVideo || prefs.UseCase == hardware.UseCaseImage {
		// Find the selected model's family for ComfyUI workflow generation.
		modelFamily := ""
		for _, rec := range recs {
			if rec.Model.OllamaTag == selectedModel {
				modelFamily = rec.Model.Family
				break
			}
		}
		ui.ShowComfyUIPipelineNote(modelFamily, string(prefs.UseCase))
	}
	if prefs.UseCase == hardware.UseCaseAudio {
		ui.ShowAudioPipelineNote()
	}

	if selectedModel == "" {
		fmt.Println(ui.MutedStyle.Render("No model selected. Goodbye!"))
		return
	}

	if wantInstall {
		if err := ui.PullWithProgress(ctx, client, selectedModel); err != nil {
			os.Exit(1)
		}
	} else {
		ui.ShowInstallInstructions(selectedModel)
	}

	// Show heretic suggestion for NSFW/uncensored use case.
	if prefs.UseCase == hardware.UseCaseNSFW {
		ui.ShowHereticSuggestion()
	}
}
