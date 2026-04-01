package comfyui

import "testing"

func TestMatchCheckpointForFamily(t *testing.T) {
	list := []string{"other.safetensors", "flux1-schnell.safetensors"}
	if got := MatchCheckpointForFamily(list, "flux"); got != "flux1-schnell.safetensors" {
		t.Fatalf("got %q", got)
	}
	if got := MatchCheckpointForFamily(list, "nosuch"); got != "other.safetensors" {
		t.Fatalf("fallback got %q", got)
	}
	if got := MatchCheckpointForFamily(nil, "flux"); got != "" {
		t.Fatalf("empty list got %q", got)
	}
}

func TestStaticCheckpointHint(t *testing.T) {
	if got := StaticCheckpointHint("flux-schnell", ""); got != "flux1-schnell.safetensors" {
		t.Fatalf("got %q", got)
	}
	if got := StaticCheckpointHint("unknown", "sdxl"); got != "sd_xl_base_1.0.safetensors" {
		t.Fatalf("got %q", got)
	}
}
