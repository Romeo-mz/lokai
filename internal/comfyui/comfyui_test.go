package comfyui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIsComfyUIDir(t *testing.T) {
	t.Run("valid dir with main.py and comfy", func(t *testing.T) {
		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "main.py"), []byte("# comfyui"), 0o644)
		_ = os.MkdirAll(filepath.Join(dir, "comfy"), 0o755)

		if !isComfyUIDir(dir) {
			t.Error("should detect valid ComfyUI dir")
		}
	})

	t.Run("valid dir with main.py and comfy_extras", func(t *testing.T) {
		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "main.py"), []byte("# comfyui"), 0o644)
		_ = os.MkdirAll(filepath.Join(dir, "comfy_extras"), 0o755)

		if !isComfyUIDir(dir) {
			t.Error("should detect dir with comfy_extras")
		}
	})

	t.Run("missing main.py", func(t *testing.T) {
		dir := t.TempDir()
		_ = os.MkdirAll(filepath.Join(dir, "comfy"), 0o755)

		if isComfyUIDir(dir) {
			t.Error("should not detect dir without main.py")
		}
	})

	t.Run("only main.py no comfy dir", func(t *testing.T) {
		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "main.py"), []byte("# not comfy"), 0o644)

		if isComfyUIDir(dir) {
			t.Error("should not detect dir without comfy/ or comfy_extras/")
		}
	})

	t.Run("nonexistent dir", func(t *testing.T) {
		if isComfyUIDir("/nonexistent/path") {
			t.Error("should return false for nonexistent path")
		}
	})
}

func TestModelDownloadInfo(t *testing.T) {
	families := []struct {
		family     string
		wantSource bool
		wantDest   string
	}{
		{"flux", true, "models/unet"},
		{"sd3", true, "models/checkpoints"},
		{"sdxl", true, "models/checkpoints"},
		{"pixart", true, "models/checkpoints"},
		{"wan", true, "models/diffusion_models"},
		{"ltx", true, "models/checkpoints"},
		{"cogvideo", true, "models/diffusion_models"},
		{"hunyuan", true, "models/diffusion_models"},
		{"unknown", false, "models/checkpoints"},
	}

	for _, tt := range families {
		t.Run(tt.family, func(t *testing.T) {
			source, dest := ModelDownloadInfo(tt.family)
			if tt.wantSource && source == "" {
				t.Error("expected non-empty source URL")
			}
			if !tt.wantSource && source != "" {
				t.Errorf("expected empty source for unknown family, got %q", source)
			}
			if dest != tt.wantDest {
				t.Errorf("dest = %q, want %q", dest, tt.wantDest)
			}
		})
	}
}

func TestInstallInstructions(t *testing.T) {
	instructions := InstallInstructions()
	if instructions == "" {
		t.Error("InstallInstructions should not be empty")
	}
	// Should contain at least a numbered step.
	if len(instructions) < 20 {
		t.Error("InstallInstructions seems too short")
	}
}

func TestWorkflowForModel(t *testing.T) {
	families := []string{"flux", "sd3", "sdxl", "pixart", "wan", "ltx", "cogvideo", "hunyuan", "unknown"}

	for _, family := range families {
		t.Run(family, func(t *testing.T) {
			wf := WorkflowForModel(family, "a beautiful sunset")
			if wf == nil {
				t.Fatal("WorkflowForModel returned nil")
			}
			if len(wf) == 0 {
				t.Error("workflow has no nodes")
			}
		})
	}
}

func TestSaveWorkflow(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "subdir", "test-workflow.json")

	wf := WorkflowForModel("flux", "test prompt")
	if err := SaveWorkflow(outPath, wf); err != nil {
		t.Fatalf("SaveWorkflow: %v", err)
	}

	// Verify file exists and is valid JSON.
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(parsed) == 0 {
		t.Error("parsed workflow is empty")
	}
}
