package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startFakeOllama boots a minimal HTTP server that mimics the Ollama REST API
// and returns a *Client already wired to it. The server is shut down when the
// test ends.
func startFakeOllama(t *testing.T, mux *http.ServeMux) *Client {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	t.Setenv("OLLAMA_HOST", srv.URL)

	c, err := NewClient()
	require.NoError(t, err, "NewClient")
	return c
}

// ─── NewClient ───────────────────────────────────────────────────────────────

func TestNewClient_Default(t *testing.T) {
	os.Unsetenv("OLLAMA_HOST")
	c, err := NewClient()
	require.NoError(t, err)
	assert.NotNil(t, c.inner)
}

func TestNewClient_CustomHost(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "http://localhost:9999")
	c, err := NewClient()
	require.NoError(t, err)
	assert.NotNil(t, c)
}

// ─── CheckHealth ─────────────────────────────────────────────────────────────

func TestCheckHealth_ServerUp(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	c := startFakeOllama(t, mux)

	err := c.CheckHealth(context.Background())
	assert.NoError(t, err)
}

func TestCheckHealth_ServerDown(t *testing.T) {
	// Point to an unreachable port — no server listening.
	t.Setenv("OLLAMA_HOST", "http://127.0.0.1:19999")
	c, err := NewClient()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = c.CheckHealth(ctx)
	assert.Error(t, err, "CheckHealth should fail when server is down")
}

// ─── ListModels ──────────────────────────────────────────────────────────────

func ollamaListPayload(tags ...string) []byte {
	type modelItem struct {
		Name    string `json:"name"`
		Model   string `json:"model"`
		Details struct {
			ParameterSize     string `json:"parameter_size"`
			QuantizationLevel string `json:"quantization_level"`
			Family            string `json:"family"`
		} `json:"details"`
		Size int64 `json:"size"`
	}
	type listResp struct {
		Models []modelItem `json:"models"`
	}
	var resp listResp
	for _, tag := range tags {
		var item modelItem
		item.Name = tag
		item.Model = tag
		item.Size = 1_000_000_000
		item.Details.Family = "llama"
		item.Details.ParameterSize = "7B"
		item.Details.QuantizationLevel = "Q4_K_M"
		resp.Models = append(resp.Models, item)
	}
	b, _ := json.Marshal(resp)
	return b
}

func TestListModels_Empty(t *testing.T) {
	mux := http.NewServeMux()
	newOllamaAPIHandler(mux, "/api/tags", ollamaListPayload())
	c := startFakeOllama(t, mux)

	models, err := c.ListModels(context.Background())
	require.NoError(t, err)
	assert.Empty(t, models)
}

func TestListModels_WithModels(t *testing.T) {
	mux := http.NewServeMux()
	newOllamaAPIHandler(mux, "/api/tags", ollamaListPayload("llama3.1:8b", "phi4:latest"))
	c := startFakeOllama(t, mux)

	models, err := c.ListModels(context.Background())
	require.NoError(t, err)
	require.Len(t, models, 2)
	assert.Equal(t, "llama3.1:8b", models[0].Name)
	assert.Equal(t, "phi4:latest", models[1].Name)
}

// ─── IsModelInstalled ────────────────────────────────────────────────────────

func TestIsModelInstalled_True(t *testing.T) {
	mux := http.NewServeMux()
	newOllamaAPIHandler(mux, "/api/tags", ollamaListPayload("llama3.1:8b"))
	c := startFakeOllama(t, mux)

	ok, err := c.IsModelInstalled(context.Background(), "llama3.1:8b")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestIsModelInstalled_False(t *testing.T) {
	mux := http.NewServeMux()
	newOllamaAPIHandler(mux, "/api/tags", ollamaListPayload("phi4:latest"))
	c := startFakeOllama(t, mux)

	ok, err := c.IsModelInstalled(context.Background(), "llama3.1:8b")
	require.NoError(t, err)
	assert.False(t, ok)
}

// ─── PullModel ───────────────────────────────────────────────────────────────

func TestPullModel_Success(t *testing.T) {
	mux := http.NewServeMux()
	// Ollama pull streams NDJSON lines.
	mux.HandleFunc("/api/pull", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		lines := []map[string]interface{}{
			{"status": "pulling manifest"},
			{"status": "downloading", "total": 1000, "completed": 500},
			{"status": "success"},
		}
		for _, l := range lines {
			b, _ := json.Marshal(l)
			w.Write(append(b, '\n'))
		}
	})
	c := startFakeOllama(t, mux)

	var progresses []PullProgress
	err := c.PullModel(context.Background(), "tinyllama:1.1b", func(p PullProgress) {
		progresses = append(progresses, p)
	})
	require.NoError(t, err)
	assert.NotEmpty(t, progresses)
}

func TestPullModel_NoProgressCallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/pull", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		b, _ := json.Marshal(map[string]string{"status": "success"})
		w.Write(append(b, '\n'))
	})
	c := startFakeOllama(t, mux)

	// nil progressFn must not panic.
	err := c.PullModel(context.Background(), "tinyllama:1.1b", nil)
	assert.NoError(t, err)
}

func TestPullModel_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/pull", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})
	c := startFakeOllama(t, mux)

	err := c.PullModel(context.Background(), "some:model", nil)
	assert.Error(t, err)
}

// ─── DeleteModel ─────────────────────────────────────────────────────────────

func TestDeleteModel_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/delete", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	})
	c := startFakeOllama(t, mux)

	err := c.DeleteModel(context.Background(), "llama3.1:8b")
	assert.NoError(t, err)
}

func TestDeleteModel_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/delete", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model not found", http.StatusNotFound)
	})
	c := startFakeOllama(t, mux)

	err := c.DeleteModel(context.Background(), "nonexistent:model")
	assert.Error(t, err)
}

// ─── DeleteAllModels ─────────────────────────────────────────────────────────

func TestDeleteAllModels_Empty(t *testing.T) {
	mux := http.NewServeMux()
	newOllamaAPIHandler(mux, "/api/tags", ollamaListPayload())
	c := startFakeOllama(t, mux)

	deleted, err := c.DeleteAllModels(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, deleted)
}

func TestDeleteAllModels_WithModels(t *testing.T) {
	deleted := 0
	mux := http.NewServeMux()
	newOllamaAPIHandler(mux, "/api/tags", ollamaListPayload("modelA:latest", "modelB:latest"))
	mux.HandleFunc("/api/delete", func(w http.ResponseWriter, _ *http.Request) {
		deleted++
		w.WriteHeader(http.StatusOK)
	})
	c := startFakeOllama(t, mux)

	names, err := c.DeleteAllModels(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, names, 2)
	assert.Equal(t, 2, deleted)
}

func TestDeleteAllModels_ProgressCallback(t *testing.T) {
	mux := http.NewServeMux()
	newOllamaAPIHandler(mux, "/api/tags", ollamaListPayload("m1:latest", "m2:latest", "m3:latest"))
	mux.HandleFunc("/api/delete", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	c := startFakeOllama(t, mux)

	type progressCall struct {
		name           string
		current, total int
	}
	var calls []progressCall
	_, err := c.DeleteAllModels(context.Background(), func(name string, current, total int) {
		calls = append(calls, progressCall{name, current, total})
	})
	require.NoError(t, err)
	require.Len(t, calls, 3)
	assert.Equal(t, 3, calls[2].total)
	assert.Equal(t, 3, calls[2].current)
}

// ─── Generate ────────────────────────────────────────────────────────────────

func TestGenerate_StreamsTokens(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		for _, chunk := range []struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}{
			{"Hello", false},
			{" world", false},
			{"", true},
		} {
			b, _ := json.Marshal(chunk)
			w.Write(append(b, '\n'))
		}
	})
	c := startFakeOllama(t, mux)

	var tokens []string
	var done bool
	err := c.Generate(context.Background(), "tinyllama:1.1b", "Say hello", func(tok string, d bool) {
		tokens = append(tokens, tok)
		if d {
			done = true
		}
	})
	require.NoError(t, err)
	assert.True(t, done)
	assert.Contains(t, tokens, "Hello")
}

// ─── InstallStatus ───────────────────────────────────────────────────────────

func TestCheckInstallation_ContainerEnv(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "http://ollama:11434")
	status := CheckInstallation()
	// In a container, OLLAMA_HOST is set so the check should treat it as installed.
	assert.True(t, status.Installed)
}

func TestGetInstallInstructions_AllPlatforms(t *testing.T) {
	// Just verify the function returns a non-empty string (platform-agnostic).
	instructions := GetInstallInstructions()
	assert.NotEmpty(t, instructions)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// newOllamaAPIHandler registers a handler that always returns body for pattern.
func newOllamaAPIHandler(mux *http.ServeMux, pattern string, body []byte) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})
}
