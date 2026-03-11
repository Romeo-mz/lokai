package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/romeo-mz/lokai/internal/comfyui"
	"github.com/romeo-mz/lokai/internal/models"
)

// GenerateOptions configures a direct image generation run.
type GenerateOptions struct {
	Prompt     string  // Positive text prompt; empty = ask interactively
	Checkpoint string  // ComfyUI checkpoint filename; empty = ask interactively
	Steps      int     // Sampling steps; 0 = auto-select based on model family
	Width      int     // Output width in pixels; 0 = 1024
	Height     int     // Output height in pixels; 0 = 1024
	CFG        float64 // Classifier-free guidance scale; 0 = auto-select
	Seed       int64   // RNG seed; -1 = random
	OutputDir  string  // Directory to save images; "" = current directory
}

// RunGenerate performs a complete text-to-image generation via ComfyUI.
// It checks connectivity, collects any missing parameters interactively,
// queues the workflow, polls for completion, and saves the resulting image.
// entry may be nil when called from --generate without a --model flag.
func RunGenerate(ctx context.Context, entry *models.ModelEntry, opts GenerateOptions) error {
	client := comfyui.NewClient()

	fmt.Println()
	fmt.Println(SubtitleStyle.Render("🖼  Connecting to ComfyUI..."))

	if err := client.CheckHealth(ctx); err != nil {
		fmt.Println(ErrorStyle.Render("✗ " + err.Error()))
		fmt.Println()
		fmt.Println(MutedStyle.Render(comfyui.GetInstallInstructions()))
		return err
	}
	fmt.Println(SuccessStyle.Render("✓ ComfyUI is running at " + client.Addr()))
	fmt.Println()

	// Fetch available checkpoints from the live ComfyUI instance.
	checkpoints, _ := client.ListCheckpoints(ctx)

	// Collect missing parameters interactively.
	if opts.Prompt == "" || opts.Checkpoint == "" {
		if err := promptGenerateOptions(entry, checkpoints, &opts); err != nil {
			return fmt.Errorf("generation cancelled: %w", err)
		}
	} else if opts.Checkpoint != "" && !checkpointInList(checkpoints, opts.Checkpoint) && len(checkpoints) > 0 {
		fmt.Println(WarningStyle.Render("⚠  Checkpoint \"" + opts.Checkpoint + "\" not found in ComfyUI."))
		fmt.Println("   Available checkpoints:")
		for _, c := range checkpoints {
			fmt.Println("      " + MutedStyle.Render(c))
		}
		return fmt.Errorf("checkpoint %q not available in ComfyUI", opts.Checkpoint)
	}

	// Apply model-family defaults for unset parameters.
	steps, cfg := resolveStepsCFG(entry, opts)
	if opts.Width == 0 {
		opts.Width = 1024
	}
	if opts.Height == 0 {
		opts.Height = 1024
	}

	// Build and queue the workflow.
	workflow := comfyui.BuildTxt2ImgWorkflow(
		opts.Checkpoint, opts.Prompt, "",
		steps, opts.Width, opts.Height, cfg, opts.Seed,
	)

	fmt.Println(SubtitleStyle.Render("🚀 Queuing generation..."))
	promptID, err := client.QueuePrompt(ctx, workflow)
	if err != nil {
		fmt.Println(ErrorStyle.Render("✗ Failed to queue: " + err.Error()))
		return err
	}
	fmt.Println(MutedStyle.Render("   prompt_id: " + promptID))
	fmt.Println()

	// Poll with a spinner until done.
	spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	tick := 0
	fmt.Print(SubtitleStyle.Render("⏳ Generating image") + "  ")

	images, err := client.WaitForResult(ctx, promptID, func() {
		fmt.Printf("\r%s  %s", SubtitleStyle.Render("⏳ Generating image"), spinChars[tick%len(spinChars)])
		tick++
	})
	fmt.Println()

	if err != nil {
		fmt.Println(ErrorStyle.Render("✗ Generation failed: " + err.Error()))
		return err
	}
	fmt.Println(SuccessStyle.Render("✓ Generation complete!"))
	fmt.Println()

	// Download and save each output image.
	outDir := opts.OutputDir
	if outDir == "" {
		outDir = "."
	}
	for i, img := range images {
		data, err := client.DownloadImage(ctx, img)
		if err != nil {
			fmt.Println(ErrorStyle.Render("✗ Could not download image: " + err.Error()))
			continue
		}
		ts := time.Now().Format("20060102-150405")
		name := fmt.Sprintf("lokai-%s-%d.png", ts, i+1)
		outPath := filepath.Join(outDir, name)
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			fmt.Println(ErrorStyle.Render("✗ Could not save image: " + err.Error()))
			continue
		}
		abs, _ := filepath.Abs(outPath)
		fmt.Printf("   %s  %s\n", SuccessStyle.Render("✓"), ValueStyle.Render(abs))
	}
	fmt.Println()
	return nil
}

// OfferGenerate is called after ShowExternalModelInstructions when the selected
// model uses the comfyui pipeline. It silently checks whether ComfyUI is
// running, and if so, offers an inline generation prompt.
func OfferGenerate(ctx context.Context, entry models.ModelEntry) {
	client := comfyui.NewClient()
	if err := client.CheckHealth(ctx); err != nil {
		return // ComfyUI not running — silently skip
	}

	fmt.Println(SuccessStyle.Render("✓ ComfyUI detected at " + client.Addr()))
	fmt.Println()

	var wantGen bool
	_ = huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Generate an image with "+entry.Name+" now?").
			Description("ComfyUI is running — queue a generation directly from lokai").
			Affirmative("Yes, generate").
			Negative("No thanks").
			Value(&wantGen),
	)).Run()

	if !wantGen {
		return
	}

	genCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	_ = RunGenerate(genCtx, &entry, GenerateOptions{Seed: -1})
}

// promptGenerateOptions collects prompt and checkpoint from the user interactively.
func promptGenerateOptions(entry *models.ModelEntry, checkpoints []string, opts *GenerateOptions) error {
	var fields []huh.Field

	// Prompt input (multi-line text area).
	if opts.Prompt == "" {
		fields = append(fields, huh.NewText().
			Title("Image prompt").
			Description("Describe the image you want to generate").
			Placeholder("a red fox sitting in a snowy forest, cinematic lighting, 8k").
			CharLimit(1000).
			Value(&opts.Prompt),
		)
	}

	// Checkpoint selection.
	if opts.Checkpoint == "" {
		if len(checkpoints) > 0 {
			// Pre-select best match if a model tag was provided.
			selected := checkpoints[0]
			if entry != nil {
				for _, c := range checkpoints {
					if strings.Contains(strings.ToLower(c), strings.ToLower(entry.Family)) {
						selected = c
						break
					}
				}
			}
			opts.Checkpoint = selected

			options := make([]huh.Option[string], len(checkpoints))
			for i, c := range checkpoints {
				options[i] = huh.NewOption(c, c)
			}
			fields = append(fields, huh.NewSelect[string]().
				Title("Checkpoint").
				Description("Model checkpoint to use for generation").
				Options(options...).
				Value(&opts.Checkpoint),
			)
		} else {
			// ComfyUI returned no checkpoints — fall back to free-text input.
			fields = append(fields, huh.NewInput().
				Title("Checkpoint filename").
				Description("Must be placed in ComfyUI/models/checkpoints/ (e.g. flux1-dev.safetensors)").
				Value(&opts.Checkpoint),
			)
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return huh.NewForm(huh.NewGroup(fields...)).Run()
}

// resolveStepsCFG returns sampling steps and CFG scale, applying model-family
// defaults when the caller left them at zero.
func resolveStepsCFG(entry *models.ModelEntry, opts GenerateOptions) (steps int, cfg float64) {
	steps = opts.Steps
	cfg = opts.CFG

	if entry != nil {
		switch entry.Family {
		case "flux":
			if steps == 0 {
				steps = 4
			}
			if cfg == 0 {
				cfg = 1.0
			}
		case "sd3":
			if steps == 0 {
				steps = 4
			}
			if cfg == 0 {
				cfg = 5.0
			}
		case "sdxl":
			if steps == 0 {
				steps = 20
			}
			if cfg == 0 {
				cfg = 7.0
			}
		}
	}
	if steps == 0 {
		steps = 20
	}
	if cfg == 0 {
		cfg = 7.0
	}
	return
}

// checkpointInList checks whether name appears in list (case-insensitive).
func checkpointInList(list []string, name string) bool {
	lower := strings.ToLower(name)
	for _, c := range list {
		if strings.ToLower(c) == lower {
			return true
		}
	}
	return false
}
