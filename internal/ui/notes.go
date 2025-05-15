package ui

import "fmt"

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

	// Unrestricted image/video generation note.
	fmt.Println(SubtitleStyle.Render("🖼  Unrestricted Image & Video Generation"))
	fmt.Println()
	fmt.Println("   For unrestricted image/video generation, use Stable Diffusion with Unrestricted checkpoints:")
	fmt.Println()
	fmt.Println("   " + ValueStyle.Render("ComfyUI") + "        — " + MutedStyle.Render("https://github.com/comfyanonymous/ComfyUI"))
	fmt.Println("   " + ValueStyle.Render("A1111 WebUI") + "    — " + MutedStyle.Render("https://github.com/AUTOMATIC1111/stable-diffusion-webui"))
	fmt.Println("   " + ValueStyle.Render("CivitAI") + "        — " + MutedStyle.Render("https://civitai.com (browse Unrestricted checkpoints & LoRAs)"))
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
