// Package comfyui provides ComfyUI detection, API integration, and
// workflow generation for image/video diffusion models.
//
// ComfyUI is the standard local pipeline for running Stable Diffusion,
// FLUX, and video generation models.
//
// Sources:
//   ComfyUI repo    — https://github.com/comfyanonymous/ComfyUI
//   ComfyUI API     — https://github.com/comfyanonymous/ComfyUI/blob/master/server.py
package comfyui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const defaultAddr = "http://127.0.0.1:8188"

// Status holds the detected state of ComfyUI on the system.
type Status struct {
	Installed    bool   `json:"installed"`
	Running      bool   `json:"running"`
	InstallPath  string `json:"install_path,omitempty"`
	ServerAddr   string `json:"server_addr"`
	Version      string `json:"version,omitempty"`
	ErrorMessage string `json:"error,omitempty"`
}

// SystemStats mirrors the subset of ComfyUI /system_stats we need.
type SystemStats struct {
	System struct {
		OS             string `json:"os"`
		PythonVersion  string `json:"python_version"`
		ComfyUIVersion string `json:"comfyui_version"`
	} `json:"system"`
	Devices []struct {
		Name      string `json:"name"`
		Type      string `json:"type"`
		VRAMTotal int64  `json:"vram_total"`
		VRAMFree  int64  `json:"vram_free"`
	} `json:"devices"`
}

// Detect checks whether ComfyUI is installed and/or running.
func Detect() Status {
	st := Status{ServerAddr: defaultAddr}

	// Check if ComfyUI server is running.
	if stats, err := fetchSystemStats(defaultAddr); err == nil {
		st.Running = true
		st.Installed = true
		st.Version = stats.System.ComfyUIVersion
		return st
	}

	// Check common installation paths.
	st.InstallPath = findInstallPath()
	st.Installed = st.InstallPath != ""

	return st
}

// IsRunning pings the ComfyUI server and returns true if it responds.
func IsRunning() bool {
	_, err := fetchSystemStats(defaultAddr)
	return err == nil
}

// GetSystemStats returns system information from the running ComfyUI server.
func GetSystemStats() (*SystemStats, error) {
	return fetchSystemStats(defaultAddr)
}

// QueuePrompt sends a workflow to ComfyUI for execution. Returns the prompt ID.
func QueuePrompt(ctx context.Context, workflow map[string]any) (string, error) {
	body := map[string]any{"prompt": workflow}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		defaultAddr+"/prompt", strings.NewReader(string(data)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ComfyUI not reachable: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ComfyUI returned HTTP %d", resp.StatusCode)
	}

	var result struct {
		PromptID string `json:"prompt_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.PromptID, nil
}

// ListModels returns model checkpoint filenames available in ComfyUI.
func ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		defaultAddr+"/object_info/CheckpointLoaderSimple", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ComfyUI not reachable: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Parse the nested response to extract checkpoint names.
	var info map[string]struct {
		Input struct {
			Required map[string]any `json:"required"`
		} `json:"input"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&info); err != nil {
		return nil, err
	}

	node, ok := info["CheckpointLoaderSimple"]
	if !ok {
		return nil, nil
	}
	ckptField, ok := node.Input.Required["ckpt_name"]
	if !ok {
		return nil, nil
	}

	// ckpt_name field is [["model1.safetensors", "model2.safetensors"]]
	if arr, ok := ckptField.([]any); ok && len(arr) > 0 {
		if names, ok := arr[0].([]any); ok {
			var out []string
			for _, n := range names {
				if s, ok := n.(string); ok {
					out = append(out, s)
				}
			}
			return out, nil
		}
	}
	return nil, nil
}

// InstallInstructions returns platform-specific ComfyUI installation steps.
func InstallInstructions() string {
	switch runtime.GOOS {
	case "darwin":
		return `  1. Install Python 3.10+:  brew install python@3.12
  2. Clone ComfyUI:         git clone https://github.com/comfyanonymous/ComfyUI.git
  3. Install dependencies:  cd ComfyUI && pip install -r requirements.txt
  4. Start:                 python main.py`
	case "windows":
		return `  1. Download the portable package from:
     https://github.com/comfyanonymous/ComfyUI/releases
  2. Extract and run:       run_nvidia_gpu.bat`
	default: // linux
		return `  1. Install Python 3.10+:  sudo apt install python3 python3-pip python3-venv
  2. Clone ComfyUI:         git clone https://github.com/comfyanonymous/ComfyUI.git
  3. Create venv:           cd ComfyUI && python3 -m venv venv && source venv/bin/activate
  4. Install dependencies:  pip install -r requirements.txt
  5. Install PyTorch+CUDA:  pip install torch torchvision --index-url https://download.pytorch.org/whl/cu124
  6. Start:                 python main.py`
	}
}

// ModelDownloadInfo returns where to download a model and where to place it in ComfyUI.
func ModelDownloadInfo(modelFamily string) (source, destSubdir string) {
	switch strings.ToLower(modelFamily) {
	case "flux":
		return "https://huggingface.co/black-forest-labs/FLUX.1-schnell", "models/unet"
	case "sd3":
		return "https://huggingface.co/stabilityai/stable-diffusion-3.5-large", "models/checkpoints"
	case "sdxl":
		return "https://huggingface.co/stabilityai/stable-diffusion-xl-base-1.0", "models/checkpoints"
	case "pixart":
		return "https://huggingface.co/PixArt-alpha/PixArt-Sigma-XL-2-1024-MS", "models/checkpoints"
	case "wan":
		return "https://huggingface.co/Wan-AI/Wan2.1-T2V-14B-Diffusers", "models/diffusion_models"
	case "ltx":
		return "https://huggingface.co/Lightricks/LTX-Video", "models/checkpoints"
	case "cogvideo":
		return "https://huggingface.co/THUDM/CogVideoX-5b", "models/diffusion_models"
	case "hunyuan":
		return "https://huggingface.co/tencent/HunyuanVideo", "models/diffusion_models"
	default:
		return "", "models/checkpoints"
	}
}

// ──────────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────────

func fetchSystemStats(addr string) (*SystemStats, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(addr + "/system_stats")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var stats SystemStats
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

func findInstallPath() string {
	// Check environment variable first.
	if p := os.Getenv("COMFYUI_PATH"); p != "" {
		if isComfyUIDir(p) {
			return p
		}
	}

	// Check if 'comfyui' is on PATH (portable/pip install).
	if p, err := exec.LookPath("comfyui"); err == nil {
		return filepath.Dir(p)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	candidates := []string{
		filepath.Join(home, "ComfyUI"),
		filepath.Join(home, "comfyui"),
		filepath.Join(home, "Documents", "ComfyUI"),
		filepath.Join(home, ".local", "share", "ComfyUI"),
		"/opt/ComfyUI",
	}

	for _, c := range candidates {
		if isComfyUIDir(c) {
			return c
		}
	}
	return ""
}

func isComfyUIDir(path string) bool {
	// ComfyUI always has main.py and a comfy/ or comfy_extras/ directory.
	if _, err := os.Stat(filepath.Join(path, "main.py")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "comfy")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(path, "comfy_extras")); err == nil {
		return true
	}
	return false
}
