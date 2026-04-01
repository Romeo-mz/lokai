package comfyui

import (
	"encoding/json"
	"os"
)

// WriteAPIPromptJSON writes workflow as ComfyUI's /prompt API envelope:
// {"prompt": { node graph }}.
func WriteAPIPromptJSON(path string, workflow map[string]any) error {
	body := map[string]any{"prompt": workflow}
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
