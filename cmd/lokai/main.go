package main

import (
	"context"
	"encoding/json"
	"flag"
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
	// ── Flag definitions ──────────────────────────────────────────────
	versionFlag := flag.Bool("version", false, "Print version and exit")
	scanOnly := flag.Bool("scan-only", false, "Scan hardware and exit")
	wantClean := flag.Bool("clean", false, "Remove all installed models")
	wantBenchmark := flag.Bool("benchmark", false, "Benchmark all installed models")
	wantClearCache := flag.Bool("clear-cache", false, "Clear all local cached data")
	jsonOutput := flag.Bool("json", false, "Output recommendations as JSON (non-interactive)")
	useCaseFlag := flag.String("use-case", "", "Use case for non-interactive mode: chat|code|vision|embedding|reasoning|image|video|audio|unrestricted")
	priorityFlag := flag.String("priority", "balanced", "Priority for non-interactive mode: speed|balanced|quality")

	// ── ComfyUI image generation flags ───────────────────────────────
	generatePrompt := flag.String("generate", "", "Generate an image via ComfyUI: --generate \"a red fox in snow\"")
	generateModel := flag.String("model", "", "Model tag for --generate (e.g. flux-schnell)")
	generateCheckpoint := flag.String("checkpoint", "", "ComfyUI checkpoint filename for --generate")
	generateSteps := flag.Int("steps", 0, "Sampling steps for --generate (0 = auto)")
	generateWidth := flag.Int("width", 1024, "Output image width for --generate")
	generateHeight := flag.Int("height", 1024, "Output image height for --generate")
	generateSeed := flag.Int64("seed", -1, "Random seed for --generate (-1 = random)")
	flag.Parse()

	// Legacy positional-word commands kept for backward compat.
	if flag.NArg() > 0 {
		switch flag.Arg(0) {
		case "clean":
			*wantClean = true
		case "benchmark":
			*wantBenchmark = true
		case "clear-cache":
			*wantClearCache = true
		}
	}

	if *versionFlag {
		fmt.Println("lokai", version)
		return
	}

	ctx := context.Background()

	// Handle --clear-cache: wipe all local cached data.
	if *wantClearCache {
		store, err := cache.New()
		if err == nil {
			_ = store.Clear()
			fmt.Println("✓ Cache cleared (" + store.Dir() + ")")
		}
		return
	}

	// --scan-only: detect hardware, display, exit (no Ollama needed).
	if *scanOnly {
		specs, err := ui.ScanAndDisplay(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "✗ "+err.Error())
			os.Exit(1)
		}
		if *jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(specs)
		}
		return
	}

	// Print banner (suppress in JSON mode for clean piping).
	if !*jsonOutput {
		fmt.Println(ui.Banner())
	}

	// Handle --generate: run image generation via ComfyUI (no Ollama needed).
	if *generatePrompt != "" {
		var entry *models.ModelEntry
		if *generateModel != "" {
			entry = models.GetModelByTag(*generateModel)
		}
		genCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
		if err := ui.RunGenerate(genCtx, entry, ui.GenerateOptions{
			Prompt:     *generatePrompt,
			Checkpoint: *generateCheckpoint,
			Steps:      *generateSteps,
			Width:      *generateWidth,
			Height:     *generateHeight,
			Seed:       *generateSeed,
		}); err != nil {
			os.Exit(1)
		}
		return
	}

	// Check Ollama installation.
	status := ollama.CheckInstallation()
	if !status.Installed {
		fmt.Fprintln(os.Stderr, "✗ "+status.ErrorMessage)
		fmt.Fprintln(os.Stderr, ollama.GetInstallInstructions())
		os.Exit(1)
	}

	// Create Ollama client and check health.
	client, err := ollama.NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "✗ "+err.Error())
		os.Exit(1)
	}

	if err := client.CheckHealth(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "✗ Ollama server is not running. Start it with: ollama serve")
		os.Exit(1)
	}

	// Handle --clean: remove all installed models.
	if *wantClean {
		if err := ui.RunCleanup(ctx, client); err != nil {
			fmt.Fprintln(os.Stderr, "✗ "+err.Error())
			os.Exit(1)
		}
		return
	}

	// Handle --benchmark: benchmark all installed models.
	if *wantBenchmark {
		if err := ui.RunBenchmark(ctx, client); err != nil {
			fmt.Fprintln(os.Stderr, "✗ "+err.Error())
			os.Exit(1)
		}
		return
	}

	// ── Non-interactive JSON mode ─────────────────────────────────────
	if *jsonOutput {
		if *useCaseFlag == "" {
			fmt.Fprintln(os.Stderr, "✗ --json requires --use-case (chat|code|vision|embedding|reasoning|image|video|audio|nsfw)")
			os.Exit(1)
		}
		runJSONMode(ctx, client, hardware.UseCase(*useCaseFlag), models.Priority(*priorityFlag))
		return
	}

	// ── Interactive mode ─────────────────────────────────────────────
	specs, err := ui.ScanAndDisplay(ctx)
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ " + err.Error()))
		os.Exit(1)
	}

	// Run questionnaire, defaulting IncludeRemote when no models are installed.
	installedModels, _ := client.ListModels(ctx)
	prefs, err := ui.RunQuestionnaire(specs, len(installedModels) > 0)
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ " + err.Error()))
		os.Exit(1)
	}

	// Refresh model registry from Ollama (best-effort, 5s timeout).
	reg := models.NewRegistry()
	refreshCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := reg.Refresh(refreshCtx); err == nil && reg.DynamicCount() > 0 {
		fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  ✓ Discovered %d additional models from Ollama", reg.DynamicCount())))
		fmt.Println()
	}

	// Get recommendations (uses dynamic catalog if available, else static).
	var catalog []models.ModelEntry
	if reg.DynamicCount() > 0 {
		catalog = reg.FullCatalog()
	}
	// Load or incrementally collect benchmark data (offers speed test for un-benchmarked models).
	benchData := ui.LoadOrPromptBenchmark(ctx, client, installedModels, *jsonOutput)

	recs := models.Recommend(specs, models.RecommendOptions{
		UseCase:       prefs.UseCase,
		SubTask:       prefs.SubTask,
		Priority:      prefs.Priority,
		IncludeRemote: prefs.IncludeRemote,
		MaxResults:    5,
		Catalog:       catalog,
		BenchmarkData: benchData,
	})

	// Display results and get user selection.
	selectedModel, wantInstall, err := ui.DisplayResults(recs, specs, prefs.UseCase, benchData)
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ " + err.Error()))
		os.Exit(1)
	}

	if selectedModel == "" {
		// No model selected — show pipeline context notes as guidance, then exit.
		if prefs.UseCase == hardware.UseCaseVideo {
			ui.ShowVideoPipelineNote()
		}
		if prefs.UseCase == hardware.UseCaseImage {
			ui.ShowImagePipelineNote()
		}
		if prefs.UseCase == hardware.UseCaseAudio {
			ui.ShowAudioPipelineNote()
		}
		fmt.Println(ui.MutedStyle.Render("No model selected. Goodbye!"))
		return
	}

	// Look up the full catalog entry to decide the correct install path.
	entry := models.GetModelByTag(selectedModel)
	if entry != nil && !entry.IsPullable() {
		// External model (diffusion / non-Ollama) — cannot use "ollama pull".
		// Show dedicated pipeline setup instructions instead.
		ui.ShowExternalModelInstructions(*entry)
		// If ComfyUI is running, offer to generate an image right now.
		if entry.Pipeline == "comfyui" {
			ui.OfferGenerate(ctx, *entry)
		}
	} else {
		// Standard Ollama model — pull and run via ollama.
		if wantInstall {
			if err := ui.PullWithProgress(ctx, client, selectedModel); err != nil {
				os.Exit(1)
			}
		} else {
			ui.ShowInstallInstructions(selectedModel)
		}
		if prefs.UseCase == hardware.UseCaseAudio {
			ui.ShowAudioPipelineNote()
		}
	}

	// Show heretic suggestion for unrestricted use case.
	if prefs.UseCase == hardware.UseCaseUnrestricted {
		ui.ShowHereticSuggestion()
	}
}

// runJSONMode handles --json non-interactive mode: detect hardware, recommend, output JSON.
func runJSONMode(ctx context.Context, client *ollama.Client, useCase hardware.UseCase, priority models.Priority) {
	specs, err := hardware.Detect(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "✗ hardware detection failed: "+err.Error())
		os.Exit(1)
	}

	reg := models.NewRegistry()
	refreshCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_ = reg.Refresh(refreshCtx)

	var catalog []models.ModelEntry
	if reg.DynamicCount() > 0 {
		catalog = reg.FullCatalog()
	}

	recs := models.Recommend(specs, models.RecommendOptions{
		UseCase:       useCase,
		Priority:      priority,
		IncludeRemote: true,
		MaxResults:    5,
		Catalog:       catalog,
	})

	type jsonRec struct {
		Rank        int                    `json:"rank"`
		OllamaTag   string                 `json:"ollama_tag"`
		Name        string                 `json:"name"`
		Size        string                 `json:"parameter_size"`
		VRAMNeeded  float64                `json:"vram_needed_gb"`
		FitsInVRAM  bool                   `json:"fits_in_vram"`
		Quality     string                 `json:"quality"`
		Score       float64                `json:"score"`
		Reason      string                 `json:"reason"`
		Performance models.PerformanceEstimate `json:"performance"`
	}

	type jsonOutput struct {
		Hardware        *hardware.HardwareSpecs `json:"hardware"`
		Recommendations []jsonRec               `json:"recommendations"`
	}

	var rJSONs []jsonRec
	for i, r := range recs {
		perf := models.EstimatePerformance(r.Model, specs)
		rJSONs = append(rJSONs, jsonRec{
			Rank:        i + 1,
			OllamaTag:   r.Model.OllamaTag,
			Name:        r.Model.Name,
			Size:        r.Model.ParameterSize,
			VRAMNeeded:  r.Model.EstimatedVRAMGB,
			FitsInVRAM:  r.FitsInVRAM,
			Quality:     perf.QualityRating,
			Score:       r.Score,
			Reason:      r.Reason,
			Performance: perf,
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(jsonOutput{Hardware: specs, Recommendations: rJSONs})
}
