package comfyui

import (
	"crypto/rand"
	"time"
)

// BuildTxt2ImgWorkflow constructs a minimal ComfyUI node-graph workflow for
// text-to-image generation, compatible with SD 1.x / 2.x, SDXL, SD3, and
// FLUX model checkpoints.
//
// The returned map can be passed directly to Client.QueuePrompt.
//
// Parameters:
//   - checkpoint: filename as shown in ComfyUI (e.g. "flux1-dev.safetensors")
//   - prompt:     positive text description
//   - negative:   negative prompt; empty string uses a built-in default
//   - steps:      sampling steps (4 for FLUX/turbo, 20–30 for SD)
//   - width/height: output resolution in pixels
//   - cfg:        classifier-free guidance scale (1.0 for FLUX, 7.0 for SD)
//   - seed:       deterministic seed; values < 0 generate a random seed
func BuildTxt2ImgWorkflow(
	checkpoint, prompt, negative string,
	steps, width, height int,
	cfg float64,
	seed int64,
) map[string]any {
	if negative == "" {
		negative = "text, watermark, blurry, low quality, deformed"
	}
	if seed < 0 {
		seed = randomSeed()
	}

	return map[string]any{
		// Node 4: load checkpoint (model + CLIP + VAE)
		"4": map[string]any{
			"class_type": "CheckpointLoaderSimple",
			"inputs": map[string]any{
				"ckpt_name": checkpoint,
			},
		},
		// Node 6: positive CLIP text encoding
		"6": map[string]any{
			"class_type": "CLIPTextEncode",
			"inputs": map[string]any{
				"text": prompt,
				"clip": []any{"4", 1},
			},
		},
		// Node 7: negative CLIP text encoding
		"7": map[string]any{
			"class_type": "CLIPTextEncode",
			"inputs": map[string]any{
				"text": negative,
				"clip": []any{"4", 1},
			},
		},
		// Node 5: empty latent image (canvas)
		"5": map[string]any{
			"class_type": "EmptyLatentImage",
			"inputs": map[string]any{
				"width":      width,
				"height":     height,
				"batch_size": 1,
			},
		},
		// Node 3: KSampler (denoising)
		"3": map[string]any{
			"class_type": "KSampler",
			"inputs": map[string]any{
				"model":        []any{"4", 0},
				"positive":     []any{"6", 0},
				"negative":     []any{"7", 0},
				"latent_image": []any{"5", 0},
				"seed":         seed,
				"steps":        steps,
				"cfg":          cfg,
				"sampler_name": "euler",
				"scheduler":    "normal",
				"denoise":      1.0,
			},
		},
		// Node 8: VAE decode (latent → pixel space)
		"8": map[string]any{
			"class_type": "VAEDecode",
			"inputs": map[string]any{
				"samples": []any{"3", 0},
				"vae":     []any{"4", 2},
			},
		},
		// Node 9: save image to output directory
		"9": map[string]any{
			"class_type": "SaveImage",
			"inputs": map[string]any{
				"images":           []any{"8", 0},
				"filename_prefix": "lokai",
			},
		},
	}
}

// randomSeed generates a random int64 seed using crypto/rand, falling back to
// a time-based seed on error.
func randomSeed() int64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UnixNano()
	}
	return int64(b[0])<<56 | int64(b[1])<<48 | int64(b[2])<<40 | int64(b[3])<<32 |
		int64(b[4])<<24 | int64(b[5])<<16 | int64(b[6])<<8 | int64(b[7])
}
