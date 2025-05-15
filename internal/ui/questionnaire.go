package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/romeo-mz/lokai/internal/hardware"
	"github.com/romeo-mz/lokai/internal/models"
)

// UserPreferences holds the user's selections from the questionnaire.
type UserPreferences struct {
	UseCase       hardware.UseCase
	Priority      models.Priority
	IncludeRemote bool
}

// RunQuestionnaire displays the interactive use-case selection form.
func RunQuestionnaire(specs *hardware.HardwareSpecs) (*UserPreferences, error) {
	prefs := &UserPreferences{}

	var useCase string
	var priority string
	var includeRemote bool

	form := huh.NewForm(
		// Page 1: Use case selection.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What do you need a local AI model for?").
				Description("Select the primary task for your model").
				Options(
					huh.NewOption("💬 Chat — General conversation & Q&A", "chat"),
					huh.NewOption("💻 Code — Code generation, completion & review", "code"),
					huh.NewOption("👁 Vision — Image understanding & analysis", "vision"),
					huh.NewOption("📐 Embedding — Text embeddings for RAG / search", "embedding"),
					huh.NewOption("🧠 Reasoning — Complex problem-solving & math", "reasoning"),
				huh.NewOption("🖼  Image Gen — Image generation (Stable Diffusion, FLUX)", "image"),
				huh.NewOption("🎬 Video — Video generation (requires ComfyUI pipeline)", "video"),
				huh.NewOption("🎙  Audio — Speech-to-text & text-to-speech", "audio"),
				huh.NewOption("🔓 Uncensored — Models without content filters (NSFW)", "nsfw"),				).
				Value(&useCase),
		),

		// Page 2: Priority.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What's your priority?").
				Description("This affects which model size we recommend").
				Options(
					huh.NewOption("⚡ Speed — Fastest possible response time", "speed"),
					huh.NewOption("⚖️  Balanced — Good quality with reasonable speed", "balanced"),
					huh.NewOption("🏆 Quality — Best output quality (larger model)", "quality"),
				).
				Value(&priority),
		),

		// Page 3: Include remote models.
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include models not yet downloaded?").
				Description("Show all compatible models, not just locally installed ones").
				Affirmative("Yes, show all").
				Negative("No, only installed").
				Value(&includeRemote),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, fmt.Errorf("questionnaire cancelled: %w", err)
	}

	prefs.UseCase = hardware.UseCase(useCase)
	prefs.Priority = models.Priority(priority)
	prefs.IncludeRemote = includeRemote

	return prefs, nil
}
