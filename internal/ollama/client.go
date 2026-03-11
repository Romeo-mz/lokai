package ollama

import (
	"context"
	"fmt"
	"time"

	"github.com/ollama/ollama/api"
)

// Client wraps the Ollama API client with convenience methods.
type Client struct {
	inner *api.Client
}

// ModelInfo holds information about a locally installed model.
type ModelInfo struct {
	Name          string   `json:"name"`
	ParameterSize string   `json:"parameter_size"`
	QuantLevel    string   `json:"quant_level"`
	Family        string   `json:"family"`
	SizeOnDisk    int64    `json:"size_on_disk"`
	Capabilities  []string `json:"capabilities,omitempty"`
}

// PullProgress reports download progress during model pull.
type PullProgress struct {
	Status    string  `json:"status"`
	Total     int64   `json:"total"`
	Completed int64   `json:"completed"`
	Percent   float64 `json:"percent"`
}

// NewClient creates a new Ollama client from environment.
// Reads OLLAMA_HOST env or defaults to http://localhost:11434.
func NewClient() (*Client, error) {
	inner, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}
	return &Client{inner: inner}, nil
}

// CheckHealth verifies that the Ollama server is running and responsive.
func (c *Client) CheckHealth(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := c.inner.Heartbeat(ctx)
	if err != nil {
		return fmt.Errorf("ollama server not responding: %w", err)
	}
	return nil
}

// ListModels returns all locally installed models.
func (c *Client) ListModels(ctx context.Context) ([]ModelInfo, error) {
	resp, err := c.inner.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var models []ModelInfo
	for _, m := range resp.Models {
		info := ModelInfo{
			Name:          m.Name,
			ParameterSize: m.Details.ParameterSize,
			QuantLevel:    m.Details.QuantizationLevel,
			Family:        m.Details.Family,
			SizeOnDisk:    m.Size,
		}
		models = append(models, info)
	}
	return models, nil
}

// ShowModel returns detailed information about a specific model.
func (c *Client) ShowModel(ctx context.Context, name string) (*ModelInfo, error) {
	resp, err := c.inner.Show(ctx, &api.ShowRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to show model %q: %w", name, err)
	}

	caps := make([]string, 0)
	for _, cap := range resp.Capabilities {
		caps = append(caps, string(cap))
	}

	return &ModelInfo{
		Name:          name,
		ParameterSize: resp.Details.ParameterSize,
		QuantLevel:    resp.Details.QuantizationLevel,
		Family:        resp.Details.Family,
		Capabilities:  caps,
	}, nil
}

// PullModel downloads a model from the Ollama registry with progress reporting.
func (c *Client) PullModel(ctx context.Context, name string, progressFn func(PullProgress)) error {
	req := &api.PullRequest{
		Name: name,
	}

	err := c.inner.Pull(ctx, req, func(resp api.ProgressResponse) error {
		if progressFn != nil {
			progress := PullProgress{
				Status:    resp.Status,
				Total:     resp.Total,
				Completed: resp.Completed,
			}
			if resp.Total > 0 {
				progress.Percent = float64(resp.Completed) / float64(resp.Total) * 100
			}
			progressFn(progress)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to pull model %q: %w", name, err)
	}
	return nil
}

// IsModelInstalled checks if a specific model is already downloaded locally.
func (c *Client) IsModelInstalled(ctx context.Context, name string) (bool, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return false, err
	}
	for _, m := range models {
		if m.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// GenerateStats captures the authoritative generation metrics returned by Ollama
// in the final streaming message. Pass a non-nil pointer to receive them.
type GenerateStats struct {
	EvalCount          int
	EvalDuration       time.Duration
	LoadDuration       time.Duration
	PromptEvalCount    int
	PromptEvalDuration time.Duration
}

// Generate runs a text generation request and streams tokens via callback.
// If stats is non-nil, Ollama's native eval metrics are written to it on completion.
func (c *Client) Generate(ctx context.Context, model, prompt string, tokenFn func(token string, done bool), stats *GenerateStats) error {
	stream := true
	req := &api.GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: &stream,
	}

	err := c.inner.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if tokenFn != nil {
			tokenFn(resp.Response, resp.Done)
		}
		if resp.Done && stats != nil {
			stats.EvalCount = resp.EvalCount
			stats.EvalDuration = resp.EvalDuration
			stats.LoadDuration = resp.LoadDuration
			stats.PromptEvalCount = resp.PromptEvalCount
			stats.PromptEvalDuration = resp.PromptEvalDuration
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("generation failed for %q: %w", model, err)
	}
	return nil
}

// DeleteModel removes a single model from the local Ollama installation.
func (c *Client) DeleteModel(ctx context.Context, name string) error {
	err := c.inner.Delete(ctx, &api.DeleteRequest{Model: name})
	if err != nil {
		return fmt.Errorf("failed to delete model %q: %w", name, err)
	}
	return nil
}

// DeleteAllModels removes all locally installed models.
// Returns the list of deleted model names and any error.
func (c *Client) DeleteAllModels(ctx context.Context, progressFn func(name string, current, total int)) ([]string, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		return nil, nil
	}

	var deleted []string
	for i, m := range models {
		if progressFn != nil {
			progressFn(m.Name, i+1, len(models))
		}
		if err := c.DeleteModel(ctx, m.Name); err != nil {
			return deleted, fmt.Errorf("failed at model %q: %w", m.Name, err)
		}
		deleted = append(deleted, m.Name)
	}

	return deleted, nil
}
