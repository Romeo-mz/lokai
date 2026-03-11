// Package comfyui provides a lightweight HTTP client for the ComfyUI REST API.
//
// ComfyUI exposes a REST API on http://localhost:8188 (by default) that accepts
// node-graph workflows as JSON, queues them, and returns generated images.
//
// Workflow:
//  1. CheckHealth — verify the server is reachable
//  2. ListCheckpoints — discover available model checkpoint files
//  3. QueuePrompt — submit a workflow JSON, receive a prompt_id
//  4. WaitForResult — poll /history/{prompt_id} until generation completes
//  5. DownloadImage — fetch image bytes from /view
package comfyui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const defaultAddr = "http://localhost:8188"

// Client communicates with a running ComfyUI instance.
type Client struct {
	addr       string
	httpClient *http.Client
}

// NewClient creates a Client, reading COMFYUI_HOST from the environment and
// falling back to http://localhost:8188.
func NewClient() *Client {
	addr := os.Getenv("COMFYUI_HOST")
	if addr == "" {
		addr = defaultAddr
	}
	return &Client{
		addr:       addr,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Addr returns the base URL this client targets.
func (c *Client) Addr() string { return c.addr }

// CheckHealth returns nil when ComfyUI is reachable and responding.
func (c *Client) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+"/system_stats", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ComfyUI not reachable at %s: %w", c.addr, err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ComfyUI returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// ListCheckpoints returns the filenames of checkpoint models available in ComfyUI.
// Returns nil (no error) when the list cannot be determined.
func (c *Client) ListCheckpoints(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.addr+"/object_info/CheckpointLoaderSimple", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing checkpoints: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("parsing checkpoint list: %w", err)
	}

	// Navigate: .CheckpointLoaderSimple.input.required.ckpt_name[0] (array of names)
	node, ok := raw["CheckpointLoaderSimple"].(map[string]any)
	if !ok {
		return nil, nil
	}
	input, ok := node["input"].(map[string]any)
	if !ok {
		return nil, nil
	}
	required, ok := input["required"].(map[string]any)
	if !ok {
		return nil, nil
	}
	ckptTuple, ok := required["ckpt_name"].([]any)
	if !ok || len(ckptTuple) == 0 {
		return nil, nil
	}
	nameList, ok := ckptTuple[0].([]any)
	if !ok {
		return nil, nil
	}
	result := make([]string, 0, len(nameList))
	for _, n := range nameList {
		if s, ok := n.(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}

// QueuePrompt submits a workflow to ComfyUI and returns the assigned prompt_id.
func (c *Client) QueuePrompt(ctx context.Context, workflow map[string]any) (string, error) {
	body, err := json.Marshal(map[string]any{"prompt": workflow})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.addr+"/prompt", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("queuing prompt: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ComfyUI rejected prompt (HTTP %d): %s", resp.StatusCode, b)
	}

	var out struct {
		PromptID string `json:"prompt_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("parsing queue response: %w", err)
	}
	if out.PromptID == "" {
		return "", fmt.Errorf("ComfyUI returned empty prompt_id")
	}
	return out.PromptID, nil
}

// OutputImage describes a generated image in ComfyUI's output directory.
type OutputImage struct {
	Filename  string
	Subfolder string
	Type      string
}

// WaitForResult polls /history/{promptID} every 2 seconds until the job
// completes or ctx is cancelled.  tickFn is called on each poll tick (may be
// nil) and can be used to update a spinner.
func (c *Client) WaitForResult(ctx context.Context, promptID string, tickFn func()) ([]OutputImage, error) {
	pollClient := &http.Client{Timeout: 5 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}

		if tickFn != nil {
			tickFn()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			c.addr+"/history/"+promptID, nil)
		if err != nil {
			continue
		}
		resp, err := pollClient.Do(req)
		if err != nil {
			continue
		}

		var history map[string]any
		decodeErr := json.NewDecoder(resp.Body).Decode(&history)
		_ = resp.Body.Close()
		if decodeErr != nil {
			continue
		}

		entry, ok := history[promptID].(map[string]any)
		if !ok {
			continue // not done yet
		}

		// Check for explicit completion/error status.
		if statusNode, ok := entry["status"].(map[string]any); ok {
			if statusStr, _ := statusNode["status_str"].(string); statusStr == "error" {
				return nil, fmt.Errorf("ComfyUI generation failed")
			}
			if completed, _ := statusNode["completed"].(bool); !completed {
				continue
			}
		}

		// Collect output images from all nodes.
		outputs, ok := entry["outputs"].(map[string]any)
		if !ok {
			continue
		}

		var images []OutputImage
		for _, nodeOut := range outputs {
			nodeMap, ok := nodeOut.(map[string]any)
			if !ok {
				continue
			}
			imgs, ok := nodeMap["images"].([]any)
			if !ok {
				continue
			}
			for _, img := range imgs {
				imgMap, ok := img.(map[string]any)
				if !ok {
					continue
				}
				filename, _ := imgMap["filename"].(string)
				subfolder, _ := imgMap["subfolder"].(string)
				imgType, _ := imgMap["type"].(string)
				if filename != "" {
					images = append(images, OutputImage{
						Filename:  filename,
						Subfolder: subfolder,
						Type:      imgType,
					})
				}
			}
		}
		if len(images) > 0 {
			return images, nil
		}
	}
}

// DownloadImage fetches image bytes from ComfyUI's /view endpoint.
func (c *Client) DownloadImage(ctx context.Context, img OutputImage) ([]byte, error) {
	params := url.Values{}
	params.Set("filename", img.Filename)
	params.Set("subfolder", img.Subfolder)
	if img.Type != "" {
		params.Set("type", img.Type)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.addr+"/view?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}
