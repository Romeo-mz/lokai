package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/romeo-mz/lokai/internal/comfyui"
)

// ShowComfyUIPipelineNote displays ComfyUI status and setup guidance for
// image or video models. It detects ComfyUI, shows its state, provides
// model-specific download instructions, and saves a ready-to-use workflow.
func ShowComfyUIPipelineNote(modelFamily string, useCase string) {
	st := comfyui.Detect()

	// Header.
	title := "🖼  Image Generation — ComfyUI Pipeline"
	if useCase == "video" {
		title = "🎬 Video Generation — ComfyUI Pipeline"
	}
	fmt.Println()
	fmt.Println(SubtitleStyle.Render(title))
	fmt.Println()

	// ComfyUI status.
	if st.Running {
		ver := ""
		if st.Version != "" {
			ver = " v" + st.Version
		}
		fmt.Println("   " + SuccessStyle.Render("✓ ComfyUI is running"+ver+" at "+st.ServerAddr))
	} else if st.Installed {
		fmt.Println("   " + WarningStyle.Render("⚠ ComfyUI found at "+st.InstallPath+" but not running"))
		fmt.Println("   " + MutedStyle.Render("Start it with:  cd "+st.InstallPath+" && python main.py"))
	} else {
		fmt.Println("   " + ErrorStyle.Render("✗ ComfyUI not detected"))
		fmt.Println()
		fmt.Println("   " + ValueStyle.Render("Install ComfyUI:"))
		fmt.Println()
		fmt.Println(comfyui.InstallInstructions())
		fmt.Println()
		fmt.Println("   " + MutedStyle.Render("Or set COMFYUI_PATH to your existing installation"))
	}
	fmt.Println()

	// Model-specific download info.
	source, destSubdir := comfyui.ModelDownloadInfo(modelFamily)
	if source != "" {
		fmt.Println(SubtitleStyle.Render("  📦 Model Weights"))
		fmt.Println()
		fmt.Println("   Download from: " + ValueStyle.Render(source))
		if st.InstallPath != "" {
			dest := filepath.Join(st.InstallPath, destSubdir)
			fmt.Println("   Place in:      " + ValueStyle.Render(dest))
		} else {
			fmt.Println("   Place in:      " + ValueStyle.Render("ComfyUI/"+destSubdir))
		}
		fmt.Println()
	}

	// Generate and save workflow.
	workflow := comfyui.WorkflowForModel(modelFamily, "A beautiful landscape, high quality, detailed")

	home, _ := os.UserHomeDir()
	workflowDir := filepath.Join(home, ".cache", "lokai", "workflows")
	workflowFile := filepath.Join(workflowDir, modelFamily+"_workflow.json")

	if err := comfyui.SaveWorkflow(workflowFile, workflow); err == nil {
		fmt.Println(SubtitleStyle.Render("  🔧 Ready-to-Use Workflow"))
		fmt.Println()
		fmt.Println("   Saved to: " + ValueStyle.Render(workflowFile))
		fmt.Println("   " + MutedStyle.Render("Import this file into ComfyUI (drag & drop or Load button)"))

		// If ComfyUI is running, offer to queue the workflow.
		if st.Running {
			fmt.Println()
			fmt.Println("   " + SuccessStyle.Render("→ ComfyUI is running — you can import and run this workflow now!"))
		}
	}

	fmt.Println()
}

// ShowHereticSuggestion displays a post-install suggestion about heretic,
// a tool for fully automatic censorship removal from GGUF model files.
func ShowHereticSuggestion() {
	fmt.Println()
	fmt.Println(SubtitleStyle.Render("🔓 Want to uncensor ANY model?"))
	fmt.Println()
	fmt.Println("   " + ValueStyle.Render("heretic") + " — Fully automatic censorship removal for language models")
	fmt.Println("   " + MutedStyle.Render("https://github.com/p-e-w/heretic"))
	fmt.Println()
	fmt.Println("   Install:")
	fmt.Println("   " + ValueStyle.Render("go install github.com/p-e-w/heretic@latest"))
	fmt.Println()
	fmt.Println("   Usage:")
	fmt.Println("   " + ValueStyle.Render("heretic model.gguf"))
	fmt.Println()
	fmt.Println(MutedStyle.Render("   Patches any GGUF model file to remove built-in refusal behavior."))
	fmt.Println(MutedStyle.Render("   Works with any model — not just the uncensored ones listed above."))
	fmt.Println()

	// NSFW image/video generation note.
	fmt.Println(SubtitleStyle.Render("🖼  NSFW Image & Video Generation"))
	fmt.Println()
	fmt.Println("   For uncensored image/video generation, use Stable Diffusion with NSFW checkpoints:")
	fmt.Println()
	fmt.Println("   " + ValueStyle.Render("ComfyUI") + "        — " + MutedStyle.Render("https://github.com/comfyanonymous/ComfyUI"))
	fmt.Println("   " + ValueStyle.Render("A1111 WebUI") + "    — " + MutedStyle.Render("https://github.com/AUTOMATIC1111/stable-diffusion-webui"))
	fmt.Println("   " + ValueStyle.Render("CivitAI") + "        — " + MutedStyle.Render("https://civitai.com (browse NSFW checkpoints & LoRAs)"))
	fmt.Println()
	fmt.Println(MutedStyle.Render("   These tools run locally and support full creative freedom."))
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
