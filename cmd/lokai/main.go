package main

import (
	"context"
	"fmt"
	"os"
	"time"

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
	refreshCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	if err := reg.Refresh(refreshCtx); err == nil && reg.DynamicCount() > 0 {
		fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  ✓ Discovered %d additional models from Ollama & GitHub", reg.DynamicCount())))
		fmt.Println()
	}
	cancel()

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

	// Display results and get user selection.
	selectedModel, wantInstall, err := ui.DisplayResults(recs, specs, prefs.UseCase)
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ " + err.Error()))
		os.Exit(1)
	}

	// Show context-specific notes.
	if prefs.UseCase == hardware.UseCaseVideo {
		ui.ShowVideoPipelineNote()
	}
	if prefs.UseCase == hardware.UseCaseImage {
		ui.ShowImagePipelineNote()
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

	// Show heretic suggestion for unrestricted use case.
	if prefs.UseCase == hardware.UseCaseUnrestricted {
		ui.ShowHereticSuggestion()
	}
}
