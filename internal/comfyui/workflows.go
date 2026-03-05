package comfyui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Workflow is a ComfyUI API-format workflow (node graph as JSON object).
type Workflow = map[string]any

// WorkflowForModel returns a ready-to-use ComfyUI workflow for the given model family.
// The workflow can be imported into ComfyUI or queued via the API.
func WorkflowForModel(modelFamily, prompt string) Workflow {
	switch modelFamily {
	case "flux":
		return fluxWorkflow(prompt)
	case "sd3":
		return sd3Workflow(prompt)
	case "sdxl":
		return sdxlWorkflow(prompt)
	case "pixart":
		return pixartWorkflow(prompt)
	case "wan", "ltx", "cogvideo", "hunyuan":
		return videoWorkflow(modelFamily, prompt)
	default:
		return sd3Workflow(prompt) // generic fallback
	}
}

// SaveWorkflow writes a workflow JSON file that can be imported into ComfyUI.
func SaveWorkflow(outputPath string, workflow Workflow) error {
	data, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(outputPath); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(outputPath, data, 0o644)
}

// ──────────────────────────────────────────────────────────────────────
// Image generation workflows
// ──────────────────────────────────────────────────────────────────────

func fluxWorkflow(prompt string) Workflow {
	return Workflow{
		"6": node("CLIPTextEncode", map[string]any{
			"text": prompt,
			"clip": []any{"11", 0},
		}),
		"8": node("VAEDecode", map[string]any{
			"samples": []any{"13", 0},
			"vae":     []any{"10", 0},
		}),
		"9": node("SaveImage", map[string]any{
			"filename_prefix": "lokai_flux",
			"images":          []any{"8", 0},
		}),
		"10": node("VAELoader", map[string]any{
			"vae_name": "ae.safetensors",
		}),
		"11": node("DualCLIPLoader", map[string]any{
			"clip_name1": "t5xxl_fp16.safetensors",
			"clip_name2": "clip_l.safetensors",
			"type":       "flux",
		}),
		"12": node("UNETLoader", map[string]any{
			"unet_name":  "flux1-schnell.safetensors",
			"weight_dtype": "default",
		}),
		"13": node("KSampler", map[string]any{
			"seed":          42,
			"steps":         4,
			"cfg":           1.0,
			"sampler_name":  "euler",
			"scheduler":     "simple",
			"denoise":       1.0,
			"model":         []any{"12", 0},
			"positive":      []any{"6", 0},
			"negative":      []any{"7", 0},
			"latent_image":  []any{"5", 0},
		}),
		"5": node("EmptyLatentImage", map[string]any{
			"width":      1024,
			"height":     1024,
			"batch_size": 1,
		}),
		"7": node("CLIPTextEncode", map[string]any{
			"text": "",
			"clip": []any{"11", 0},
		}),
	}
}

func sd3Workflow(prompt string) Workflow {
	return Workflow{
		"3": node("KSampler", map[string]any{
			"seed":          42,
			"steps":         28,
			"cfg":           7.0,
			"sampler_name":  "euler",
			"scheduler":     "normal",
			"denoise":       1.0,
			"model":         []any{"4", 0},
			"positive":      []any{"6", 0},
			"negative":      []any{"7", 0},
			"latent_image":  []any{"5", 0},
		}),
		"4": node("CheckpointLoaderSimple", map[string]any{
			"ckpt_name": "sd3.5_large.safetensors",
		}),
		"5": node("EmptyLatentImage", map[string]any{
			"width":      1024,
			"height":     1024,
			"batch_size": 1,
		}),
		"6": node("CLIPTextEncode", map[string]any{
			"text": prompt,
			"clip": []any{"4", 1},
		}),
		"7": node("CLIPTextEncode", map[string]any{
			"text": "",
			"clip": []any{"4", 1},
		}),
		"8": node("VAEDecode", map[string]any{
			"samples": []any{"3", 0},
			"vae":     []any{"4", 2},
		}),
		"9": node("SaveImage", map[string]any{
			"filename_prefix": "lokai_sd3",
			"images":          []any{"8", 0},
		}),
	}
}

func sdxlWorkflow(prompt string) Workflow {
	return Workflow{
		"3": node("KSampler", map[string]any{
			"seed":          42,
			"steps":         25,
			"cfg":           7.5,
			"sampler_name":  "euler_ancestral",
			"scheduler":     "normal",
			"denoise":       1.0,
			"model":         []any{"4", 0},
			"positive":      []any{"6", 0},
			"negative":      []any{"7", 0},
			"latent_image":  []any{"5", 0},
		}),
		"4": node("CheckpointLoaderSimple", map[string]any{
			"ckpt_name": "sdxl_base_1.0.safetensors",
		}),
		"5": node("EmptyLatentImage", map[string]any{
			"width":      1024,
			"height":     1024,
			"batch_size": 1,
		}),
		"6": node("CLIPTextEncode", map[string]any{
			"text": prompt,
			"clip": []any{"4", 1},
		}),
		"7": node("CLIPTextEncode", map[string]any{
			"text": "ugly, blurry, low quality, deformed",
			"clip": []any{"4", 1},
		}),
		"8": node("VAEDecode", map[string]any{
			"samples": []any{"3", 0},
			"vae":     []any{"4", 2},
		}),
		"9": node("SaveImage", map[string]any{
			"filename_prefix": "lokai_sdxl",
			"images":          []any{"8", 0},
		}),
	}
}

func pixartWorkflow(prompt string) Workflow {
	return Workflow{
		"3": node("KSampler", map[string]any{
			"seed":          42,
			"steps":         20,
			"cfg":           4.5,
			"sampler_name":  "euler",
			"scheduler":     "normal",
			"denoise":       1.0,
			"model":         []any{"4", 0},
			"positive":      []any{"6", 0},
			"negative":      []any{"7", 0},
			"latent_image":  []any{"5", 0},
		}),
		"4": node("CheckpointLoaderSimple", map[string]any{
			"ckpt_name": "pixart_sigma_xl2.safetensors",
		}),
		"5": node("EmptyLatentImage", map[string]any{
			"width":      1024,
			"height":     1024,
			"batch_size": 1,
		}),
		"6": node("CLIPTextEncode", map[string]any{
			"text": prompt,
			"clip": []any{"4", 1},
		}),
		"7": node("CLIPTextEncode", map[string]any{
			"text": "",
			"clip": []any{"4", 1},
		}),
		"8": node("VAEDecode", map[string]any{
			"samples": []any{"3", 0},
			"vae":     []any{"4", 2},
		}),
		"9": node("SaveImage", map[string]any{
			"filename_prefix": "lokai_pixart",
			"images":          []any{"8", 0},
		}),
	}
}

// ──────────────────────────────────────────────────────────────────────
// Video generation workflows
// ──────────────────────────────────────────────────────────────────────

func videoWorkflow(family, prompt string) Workflow {
	var ckpt string
	var steps int
	switch family {
	case "wan":
		ckpt = "wan2.1_t2v_14b.safetensors"
		steps = 30
	case "ltx":
		ckpt = "ltx-video-2b-v0.9.safetensors"
		steps = 25
	case "cogvideo":
		ckpt = "cogvideox_5b.safetensors"
		steps = 50
	case "hunyuan":
		ckpt = "hunyuan_video.safetensors"
		steps = 30
	default:
		ckpt = fmt.Sprintf("%s.safetensors", family)
		steps = 30
	}

	return Workflow{
		"3": node("KSampler", map[string]any{
			"seed":          42,
			"steps":         steps,
			"cfg":           6.0,
			"sampler_name":  "euler",
			"scheduler":     "normal",
			"denoise":       1.0,
			"model":         []any{"4", 0},
			"positive":      []any{"6", 0},
			"negative":      []any{"7", 0},
			"latent_image":  []any{"5", 0},
		}),
		"4": node("CheckpointLoaderSimple", map[string]any{
			"ckpt_name": ckpt,
		}),
		"5": node("EmptyLatentImage", map[string]any{
			"width":      512,
			"height":     512,
			"batch_size": 16, // frames
		}),
		"6": node("CLIPTextEncode", map[string]any{
			"text": prompt,
			"clip": []any{"4", 1},
		}),
		"7": node("CLIPTextEncode", map[string]any{
			"text": "",
			"clip": []any{"4", 1},
		}),
		"8": node("VAEDecode", map[string]any{
			"samples": []any{"3", 0},
			"vae":     []any{"4", 2},
		}),
		"9": node("SaveImage", map[string]any{
			"filename_prefix": fmt.Sprintf("lokai_%s", family),
			"images":          []any{"8", 0},
		}),
	}
}

// ──────────────────────────────────────────────────────────────────────

func node(classType string, inputs map[string]any) map[string]any {
	return map[string]any{
		"class_type": classType,
		"inputs":     inputs,
	}
}
