package ui

import (
	"fmt"
	"strings"

	"github.com/romeo-mz/lokai/internal/models"
)

// comfyUIVideoExampleDocsURL returns official ComfyUI example docs for video families in the catalog.
func comfyUIVideoExampleDocsURL(family string) string {
	switch strings.ToLower(strings.TrimSpace(family)) {
	case "wan":
		return "https://comfyanonymous.github.io/ComfyUI_examples/wan/"
	case "ltx":
		return "https://comfyanonymous.github.io/ComfyUI_examples/ltxv/"
	case "cogvideo":
		return "https://comfyanonymous.github.io/ComfyUI_examples/video/"
	case "hunyuan":
		return "https://comfyanonymous.github.io/ComfyUI_examples/hunyuan_video/"
	default:
		return "https://comfyanonymous.github.io/ComfyUI_examples/video/"
	}
}

// ShowVideoPipelineNote displays an informational box explaining that video
// generation models require a separate diffusion pipeline.
func ShowVideoPipelineNote() {
	note := SubtitleStyle.Render("🎬 Video Generation — Pipeline Required") + "\n\n" +
		"   Video generation models do NOT run as standard Ollama chat models.\n" +
		"   They require a diffusion pipeline such as:\n\n" +
		"   " + ValueStyle.Render("ComfyUI") + "        — " + MutedStyle.Render("https://github.com/comfyanonymous/ComfyUI") + "\n" +
		"   " + ValueStyle.Render("A1111 WebUI") + "    — " + MutedStyle.Render("https://github.com/AUTOMATIC1111/stable-diffusion-webui") + "\n" +
		"   " + ValueStyle.Render("Forge WebUI") + "    — " + MutedStyle.Render("https://github.com/lllyasviel/stable-diffusion-webui-forge") + "\n\n" +
		MutedStyle.Render("   Ollama pulls the model weights; you then load them in your pipeline of choice.")

	fmt.Println()
	fmt.Println(WarningBoxStyle.Render(note))
	fmt.Println()
}

// ShowHereticSuggestion displays a post-install suggestion about heretic,
// a tool for fully automatic filter removal from GGUF model files.
func ShowHereticSuggestion() {
	fmt.Println()
	fmt.Println(SubtitleStyle.Render("🔓 Want to remove filters from ANY model?"))
	fmt.Println()
	fmt.Println("   " + ValueStyle.Render("heretic") + " — Fully automatic filter removal for language models")
	fmt.Println("   " + MutedStyle.Render("https://github.com/p-e-w/heretic"))
	fmt.Println()
	fmt.Println("   Install:")
	fmt.Println("   " + ValueStyle.Render("go install github.com/p-e-w/heretic@latest"))
	fmt.Println()
	fmt.Println("   Usage:")
	fmt.Println("   " + ValueStyle.Render("heretic model.gguf"))
	fmt.Println()
	fmt.Println(MutedStyle.Render("   Patches any GGUF model file to remove built-in refusal behavior."))
	fmt.Println(MutedStyle.Render("   Works with any model — not just the unrestricted ones listed above."))
	fmt.Println()

	// Uncensored image/video generation note.
	fmt.Println(SubtitleStyle.Render("🖼  Unrestricted Image & Video Generation"))
	fmt.Println()
	fmt.Println("   For unrestricted image/video generation, use Stable Diffusion with open checkpoints:")
	fmt.Println()
	fmt.Println("   " + ValueStyle.Render("ComfyUI") + "        — " + MutedStyle.Render("https://github.com/comfyanonymous/ComfyUI"))
	fmt.Println("   " + ValueStyle.Render("A1111 WebUI") + "    — " + MutedStyle.Render("https://github.com/AUTOMATIC1111/stable-diffusion-webui"))
	fmt.Println("   " + ValueStyle.Render("CivitAI") + "        — " + MutedStyle.Render("https://civitai.com (browse open checkpoints & LoRAs)"))
	fmt.Println()
	fmt.Println(MutedStyle.Render("   These tools run locally and support full creative freedom."))
	fmt.Println()
}

// ShowImagePipelineNote displays an informational box explaining that image
// generation models require a separate diffusion pipeline.
func ShowImagePipelineNote() {
	note := SubtitleStyle.Render("🖼  Image Generation — Pipeline Required") + "\n\n" +
		"   Image generation models do NOT run as standard Ollama chat models.\n" +
		"   They require a diffusion pipeline such as:\n\n" +
		"   " + ValueStyle.Render("ComfyUI") + "        — " + MutedStyle.Render("https://github.com/comfyanonymous/ComfyUI") + "\n" +
		"   " + ValueStyle.Render("A1111 WebUI") + "    — " + MutedStyle.Render("https://github.com/AUTOMATIC1111/stable-diffusion-webui") + "\n" +
		"   " + ValueStyle.Render("Forge WebUI") + "    — " + MutedStyle.Render("https://github.com/lllyasviel/stable-diffusion-webui-forge") + "\n\n" +
		MutedStyle.Render("   Ollama pulls the model weights; you then load them in your pipeline of choice.")

	fmt.Println()
	fmt.Println(WarningBoxStyle.Render(note))
	fmt.Println()
}

// ShowAudioPipelineNote displays an informational box explaining that audio
// models may require dedicated tools beyond Ollama.
func ShowAudioPipelineNote() {
	note := SubtitleStyle.Render("🎙  Audio Models — Dedicated Tools") + "\n\n" +
		"   Some audio models require dedicated inference tools:\n\n" +
		"   " + ValueStyle.Render("whisper.cpp") + "    — " + MutedStyle.Render("https://github.com/ggerganov/whisper.cpp") + "  (speech-to-text)\n" +
		"   " + ValueStyle.Render("Piper TTS") + "      — " + MutedStyle.Render("https://github.com/rhasspy/piper") + "            (text-to-speech)\n" +
		"   " + ValueStyle.Render("Bark") + "           — " + MutedStyle.Render("https://github.com/suno-ai/bark") + "            (TTS + sound effects)\n\n" +
		MutedStyle.Render("   Ollama can serve some audio models; others need the tools above.")

	fmt.Println()
	fmt.Println(WarningBoxStyle.Render(note))
	fmt.Println()
}

// ShowExternalModelInstructions prints step-by-step setup guidance for models
// that require a dedicated inference pipeline (e.g. ComfyUI for diffusion models)
// instead of the standard "ollama pull" workflow.
func ShowExternalModelInstructions(entry models.ModelEntry) {
	fmt.Println()

	// Determine whether this is a video or image generation model.
	isVideo := false
	for _, cap := range entry.Capabilities {
		if cap == "video-generation" {
			isVideo = true
			break
		}
	}

	var title string
	switch entry.Pipeline {
	case "comfyui":
		if isVideo {
			title = "🎬 " + entry.Name + " — Video Generation Pipeline"
		} else {
			title = "🖼  " + entry.Name + " — Image Generation Pipeline"
		}
	default:
		title = "⚙  " + entry.Name + " — External Pipeline Required"
	}

	note := SubtitleStyle.Render(title) + "\n\n" +
		"   This model cannot be installed via \"ollama pull\".\n" +
		"   It requires a dedicated diffusion pipeline to run.\n\n"

	if entry.Pipeline == "comfyui" {
		note += "   " + SubtitleStyle.Render("Step 1 — Download the model weights") + "\n" +
			"   " + ValueStyle.Render(entry.ExternalURL) + "\n\n" +
			"   " + SubtitleStyle.Render("Step 2 — Install a pipeline") + "\n" +
			"   " + ValueStyle.Render("ComfyUI") + "        — " + MutedStyle.Render("https://github.com/comfyanonymous/ComfyUI") + "\n" +
			"   " + ValueStyle.Render("A1111 WebUI") + "    — " + MutedStyle.Render("https://github.com/AUTOMATIC1111/stable-diffusion-webui") + "\n" +
			"   " + ValueStyle.Render("Forge WebUI") + "    — " + MutedStyle.Render("https://github.com/lllyasviel/stable-diffusion-webui-forge") + "\n\n" +
			"   " + SubtitleStyle.Render("Step 3 — Load the model in your pipeline") + "\n" +
			MutedStyle.Render("   Place the downloaded weights in the pipeline's models/checkpoints directory\n") +
			MutedStyle.Render("   and select it from the UI. ComfyUI also supports drag-and-drop workflows.")
		if isVideo {
			videoURL := comfyUIVideoExampleDocsURL(entry.Family)
			note += "\n\n" +
				"   " + SubtitleStyle.Render("Step 4 — Example ComfyUI workflow (video)") + "\n" +
				"   Official examples and node graphs differ per model; start from:\n" +
				"   " + ValueStyle.Render(videoURL) + "\n" +
				MutedStyle.Render("   Install any extra custom nodes the example requires, then match weights paths.")
		}
	} else if entry.ExternalURL != "" {
		note += "   " + ValueStyle.Render(entry.ExternalURL) + "\n"
	}

	fmt.Println(WarningBoxStyle.Render(note))
	fmt.Println()
}
