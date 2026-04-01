package comfyui

import (
	"context"
	"strings"
)

// MatchCheckpointForFamily picks a checkpoint whose filename contains the model
// family (case-insensitive). If none match, tries common aliases for sd3, flux,
// sdxl, pixart. Otherwise returns the first list entry, or "" when list is empty.
func MatchCheckpointForFamily(checkpoints []string, family string) string {
	if len(checkpoints) == 0 {
		return ""
	}
	fam := strings.ToLower(strings.TrimSpace(family))
	if fam != "" {
		for _, c := range checkpoints {
			if strings.Contains(strings.ToLower(c), fam) {
				return c
			}
		}
	}
	for _, c := range checkpoints {
		lc := strings.ToLower(c)
		switch fam {
		case "sd3":
			if strings.Contains(lc, "sd3") {
				return c
			}
		case "flux":
			if strings.Contains(lc, "flux") {
				return c
			}
		case "sdxl":
			if strings.Contains(lc, "xl") || strings.Contains(lc, "sdxl") {
				return c
			}
		case "pixart":
			if strings.Contains(lc, "pixart") {
				return c
			}
		}
	}
	return checkpoints[0]
}

// staticCheckpointByTag maps catalog OllamaTag (lowered) to a typical .safetensors
// name for offline export hints. Users should rename to match their on-disk file.
var staticCheckpointByTag = map[string]string{
	"sd3.5-large-turbo": "sd3.5_large_turbo.safetensors",
	"sd3.5-large":       "sd3.5_large.safetensors",
	"flux-schnell":      "flux1-schnell.safetensors",
	"flux-dev":          "flux1-dev.safetensors",
	"pixart-sigma":      "PixArt-Sigma-XL-2-1024-MS.safetensors",
	"sdxl":              "sd_xl_base_1.0.safetensors",
}

// StaticCheckpointHint returns a typical checkpoint filename when ComfyUI is
// unreachable or has no checkpoints listed.
func StaticCheckpointHint(ollamaTag, family string) string {
	key := strings.ToLower(strings.TrimSpace(ollamaTag))
	if s, ok := staticCheckpointByTag[key]; ok {
		return s
	}
	switch strings.ToLower(strings.TrimSpace(family)) {
	case "flux":
		return "flux1-dev.safetensors"
	case "sd3":
		return "sd3.5_large.safetensors"
	case "sdxl":
		return "sd_xl_base_1.0.safetensors"
	case "pixart":
		return "PixArt-Sigma-XL-2-1024-MS.safetensors"
	default:
		return "model.safetensors"
	}
}

// SuggestCheckpoint resolves a checkpoint: explicit override wins; then a live
// ComfyUI list match; then StaticCheckpointHint. client may be nil — only
// ListCheckpoints is used when non-nil.
func SuggestCheckpoint(ctx context.Context, client *Client, family, ollamaTag, override string) string {
	if strings.TrimSpace(override) != "" {
		return override
	}
	var checkpoints []string
	if client != nil {
		checkpoints, _ = client.ListCheckpoints(ctx)
	}
	if c := MatchCheckpointForFamily(checkpoints, family); c != "" {
		return c
	}
	return StaticCheckpointHint(ollamaTag, family)
}
