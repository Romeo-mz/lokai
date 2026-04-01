package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/romeo-mz/lokai/internal/comfyui"
	"github.com/romeo-mz/lokai/internal/models"
)

// EntryIsImageComfyUI is true when the catalog entry is an image-generation model
// meant to run under ComfyUI.
func EntryIsImageComfyUI(entry models.ModelEntry) bool {
	if entry.Pipeline != "comfyui" {
		return false
	}
	for _, c := range entry.Capabilities {
		if c == "image-generation" {
			return true
		}
	}
	return false
}

// ExportComfyWorkflowOptions configures writing an example API workflow JSON.
type ExportComfyWorkflowOptions struct {
	OutPath    string
	Checkpoint string
	Prompt     string
	Steps      int
	Width      int
	Height     int
	CFG        float64
	Seed       int64
}

// RunExportComfyWorkflow writes a ComfyUI /prompt-compatible JSON file using
// the same txt2img graph as RunGenerate. entry must be non-nil.
func RunExportComfyWorkflow(ctx context.Context, entry *models.ModelEntry, opts ExportComfyWorkflowOptions) error {
	if entry == nil {
		return fmt.Errorf("model entry required")
	}
	var client *comfyui.Client
	checkpoint := comfyui.SuggestCheckpoint(ctx, client, entry.Family, entry.OllamaTag, opts.Checkpoint)

	genOpts := GenerateOptions{
		Prompt:     opts.Prompt,
		Checkpoint: checkpoint,
		Steps:      opts.Steps,
		Width:      opts.Width,
		Height:     opts.Height,
		CFG:        opts.CFG,
		Seed:       opts.Seed,
	}
	steps, cfg := resolveStepsCFG(entry, genOpts)
	width, height := opts.Width, opts.Height
	if width == 0 {
		width = 1024
	}
	if height == 0 {
		height = 1024
	}
	prompt := strings.TrimSpace(opts.Prompt)
	if prompt == "" {
		prompt = "lokai example — describe your scene here"
	}

	workflow := comfyui.BuildTxt2ImgWorkflow(
		checkpoint, prompt, "",
		steps, width, height, cfg, opts.Seed,
	)

	outPath := opts.OutPath
	if outPath == "" {
		return fmt.Errorf("output path required")
	}
	if err := comfyui.WriteAPIPromptJSON(outPath, workflow); err != nil {
		return err
	}

	abs, _ := filepath.Abs(outPath)
	fmt.Println()
	fmt.Println(SuccessStyle.Render("✓ Example ComfyUI workflow saved"))
	fmt.Println(MutedStyle.Render("   checkpoint: " + checkpoint))
	fmt.Println(ValueStyle.Render(abs))
	fmt.Println()
	fmt.Println(MutedStyle.Render("   Tip: ensure this file matches a checkpoint under ComfyUI/models/checkpoints/,"))
	fmt.Println(MutedStyle.Render("   or edit the ckpt_name field in node \"4\" before loading."))
	fmt.Println()
	return nil
}

// OfferComfyWorkflowExport asks whether to save an API workflow JSON next to
// the current directory with a default filename based on the model tag.
func OfferComfyWorkflowExport(ctx context.Context, entry models.ModelEntry) {
	fmt.Println()

	var want bool
	if err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Save example ComfyUI workflow JSON?").
			Description("Writes an API-format workflow you can POST to ComfyUI or use as a template").
			Affirmative("Yes, save").
			Negative("No thanks").
			Value(&want),
	)).Run(); err != nil || !want {
		return
	}

	tag := strings.ReplaceAll(entry.OllamaTag, ":", "-")
	defaultPath := filepath.Join(".", fmt.Sprintf("lokai-%s-txt2img.json", tag))

	path := defaultPath
	_ = huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Output file path").
			Description("JSON path for the workflow (ComfyUI API prompt format)").
			Value(&path),
	)).Run()

	if strings.TrimSpace(path) == "" {
		path = defaultPath
	}

	exportCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := RunExportComfyWorkflow(exportCtx, &entry, ExportComfyWorkflowOptions{OutPath: path, Seed: -1}); err != nil {
		fmt.Println(ErrorStyle.Render("✗ " + err.Error()))
		return
	}
}
