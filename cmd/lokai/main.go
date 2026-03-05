package main

import (
	"context"
	"fmt"
	"os"

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

	ctx := context.Background()

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

	// Get recommendations.
	recs := models.Recommend(specs, models.RecommendOptions{
		UseCase:       prefs.UseCase,
		SubTask:       prefs.SubTask,
		Priority:      prefs.Priority,
		IncludeRemote: prefs.IncludeRemote,
		MaxResults:    5,
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

	// Show heretic suggestion for NSFW/uncensored use case.
	if prefs.UseCase == hardware.UseCaseNSFW {
		ui.ShowHereticSuggestion()
	}
}
