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
	SubTask       models.SubTask
	Priority      models.Priority
	IncludeRemote bool
}

// RunQuestionnaire displays the interactive use-case selection form.
// hasInstalledModels controls whether the "include remote" question is shown;
// when false (empty library) it is skipped and remote models are always included.
func RunQuestionnaire(specs *hardware.HardwareSpecs, hasInstalledModels bool) (*UserPreferences, error) {
	prefs := &UserPreferences{}

	var useCase string
	var subTask string

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
					huh.NewOption("🔓 Unrestricted — Models without content filters (Unrestricted)", "unrestricted"),
				).
				Value(&useCase),
		),
	)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("questionnaire cancelled: %w", err)
	}

	// Page 2: Sub-task selection (context-sensitive based on use case).
	subTaskOptions := subTasksForUseCase(hardware.UseCase(useCase))
	if len(subTaskOptions) > 0 {
		subForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What will you mainly use it for?").
					Description("This helps us pick the perfect model for your needs").
					Options(subTaskOptions...).
					Value(&subTask),
			),
		)
		if err := subForm.Run(); err != nil {
			return nil, fmt.Errorf("questionnaire cancelled: %w", err)
		}
	}

	// Page 3: Priority + optionally include remote.
	priority := "balanced"
	var includeRemote bool

	priorityForm := huh.NewForm(
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
	)
	if err := priorityForm.Run(); err != nil {
		return nil, fmt.Errorf("questionnaire cancelled: %w", err)
	}

	// Skip the remote-models question when the library is empty: always include all.
	if hasInstalledModels {
		remoteForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Include models not yet downloaded?").
					Description("Show all compatible models, not just locally installed ones").
					Affirmative("Yes, show all").
					Negative("No, only installed").
					Value(&includeRemote),
			),
		)
		if err := remoteForm.Run(); err != nil {
			return nil, fmt.Errorf("questionnaire cancelled: %w", err)
		}
	} else {
		// Nothing is installed yet — always show all models so results aren’t empty.
		includeRemote = true
	}

	prefs.UseCase = hardware.UseCase(useCase)
	prefs.SubTask = models.SubTask(subTask)
	prefs.Priority = models.Priority(priority)
	prefs.IncludeRemote = includeRemote

	return prefs, nil
}

// subTasksForUseCase returns the sub-task options for a given use case.
func subTasksForUseCase(uc hardware.UseCase) []huh.Option[string] {
	switch uc {
	case hardware.UseCaseChat:
		return []huh.Option[string]{
			huh.NewOption("💬 Quick Q&A — Short answers, lookups", string(models.SubTaskQuickQA)),
			huh.NewOption("✉️  Writing — Emails, essays, articles", string(models.SubTaskWriting)),
			huh.NewOption("🎨 Creative — Stories, poetry, role-play", string(models.SubTaskCreative)),
			huh.NewOption("🌍 Translation — Between languages", string(models.SubTaskTranslation)),
			huh.NewOption("📚 Summarization — Condense long texts", string(models.SubTaskSummarization)),
		}
	case hardware.UseCaseCode:
		return []huh.Option[string]{
			huh.NewOption("⚡ Autocomplete — Fast inline suggestions", string(models.SubTaskAutocomplete)),
			huh.NewOption("🏗  Project Gen — Build full pages/components", string(models.SubTaskProjectGen)),
			huh.NewOption("🔍 Code Review — Review & refactor existing code", string(models.SubTaskCodeReview)),
			huh.NewOption("🐛 Debugging — Find and fix bugs", string(models.SubTaskDebugging)),
			huh.NewOption("📝 Documentation — Generate docs & comments", string(models.SubTaskDocumentation)),
		}
	case hardware.UseCaseVision:
		return []huh.Option[string]{
			huh.NewOption("📸 Photo Analysis — Describe images & scenes", string(models.SubTaskPhotoAnalysis)),
			huh.NewOption("📄 OCR / Document — Read text from images", string(models.SubTaskOCR)),
			huh.NewOption("📊 Charts & Diagrams — Interpret visual data", string(models.SubTaskChartAnalysis)),
		}
	case hardware.UseCaseReasoning:
		return []huh.Option[string]{
			huh.NewOption("🔢 Math — Equations, calculations, proofs", string(models.SubTaskMath)),
			huh.NewOption("🧩 Logic — Puzzles, word problems, deductions", string(models.SubTaskLogic)),
			huh.NewOption("📋 Planning — Step-by-step plans, strategies", string(models.SubTaskPlanning)),
			huh.NewOption("🔬 Research — Analysis, comparisons, deep-dives", string(models.SubTaskResearch)),
		}
	case hardware.UseCaseImage:
		return []huh.Option[string]{
			huh.NewOption("📷 Photorealistic — Lifelike photos", string(models.SubTaskPhotorealistic)),
			huh.NewOption("🎨 Artistic — Stylized, illustrations, concept art", string(models.SubTaskArtistic)),
			huh.NewOption("⚡ Fast Drafts — Quick iterations, low res", string(models.SubTaskFastDrafts)),
		}
	case hardware.UseCaseAudio:
		return []huh.Option[string]{
			huh.NewOption("🎤 Transcription — Speech to text", string(models.SubTaskTranscription)),
			huh.NewOption("🔊 Text to Speech — Generate spoken audio", string(models.SubTaskTTS)),
		}
	default:
		return nil
	}
}
