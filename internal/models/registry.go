// Package models — dynamic model registry.
//
// The registry merges the static built-in catalog with models discovered
// at runtime from the Ollama library and GitHub.  Results are cached to
// disk so subsequent runs don't need network access.
//
// Sources:
//
//	Ollama Registry — https://registry.ollama.ai/v2/library
//	Ollama Library  — https://ollama.com/search
//	GitHub API      — https://api.github.com/search/repositories
package models

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/romeo-mz/lokai/internal/cache"
)

const registryCacheTTL = 24 * time.Hour

const registryCacheKey = "registry"

// Registry manages a dynamic model catalog that merges the static
// built-in entries with models discovered from online sources.
type Registry struct {
	mu          sync.RWMutex
	dynamic     []ModelEntry
	lastRefresh time.Time
	store       *cache.Store
}

// NewRegistry creates a Registry, loading any previously cached data.
func NewRegistry() *Registry {
	store, _ := cache.New()
	r := &Registry{store: store}
	r.loadCache()
	return r
}

// Refresh fetches fresh model data from Ollama and GitHub, merges it
// with the static catalog, and writes the result to the disk cache.
// If the cache is still fresh (< 24 h), this is a no-op.
func (r *Registry) Refresh(ctx context.Context) error {
	r.mu.RLock()
	fresh := time.Since(r.lastRefresh) < registryCacheTTL && len(r.dynamic) > 0
	r.mu.RUnlock()
	if fresh {
		return nil
	}

	var discovered []DiscoveredModel

	ollamaModels, err := FetchOllamaModels(ctx)
	if err == nil {
		discovered = append(discovered, ollamaModels...)
	}

	githubModels, _ := FetchGitHubModels(ctx) // best-effort
	discovered = append(discovered, githubModels...)

	if len(discovered) == 0 {
		return fmt.Errorf("no models discovered from any source")
	}

	entries := mergeDiscovered(discovered)

	r.mu.Lock()
	r.dynamic = entries
	r.lastRefresh = time.Now()
	r.mu.Unlock()

	r.saveCache()
	return err
}

// DynamicCount returns how many extra models were found beyond the static catalog.
func (r *Registry) DynamicCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.dynamic)
}

// FullCatalog returns every known model: static catalog + dynamic discoveries.
func (r *Registry) FullCatalog() []ModelEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]ModelEntry, 0, len(Catalog)+len(r.dynamic))
	all = append(all, Catalog...)
	all = append(all, r.dynamic...)
	return all
}

// GetModelsByUseCase returns models from the full catalog matching a use case.
func (r *Registry) GetModelsByUseCase(uc UseCase) []ModelEntry {
	catalog := r.FullCatalog()
	var out []ModelEntry
	for _, m := range catalog {
		for _, supported := range m.UseCases {
			if supported == uc {
				out = append(out, m)
				break
			}
		}
	}
	return out
}

// ──────────────────────────────────────────────────────────────────────
// Merge & inference
// ──────────────────────────────────────────────────────────────────────

// mergeDiscovered converts discovered models to ModelEntry, skipping
// models already in the static catalog.
func mergeDiscovered(discovered []DiscoveredModel) []ModelEntry {
	staticTags := make(map[string]bool)
	for _, m := range Catalog {
		staticTags[m.OllamaTag] = true
	}

	var entries []ModelEntry
	seen := make(map[string]bool)

	for _, d := range discovered {
		// GitHub repos are informational only — not directly Ollama-pullable.
		if d.Source == "github" {
			continue
		}

		if len(d.Tags) > 0 {
			for _, tag := range d.Tags {
				ollamaTag := d.Name + ":" + tag
				if staticTags[ollamaTag] || seen[ollamaTag] {
					continue
				}
				seen[ollamaTag] = true
				if e := inferModelEntry(d.Name, tag, d.Description); e != nil {
					entries = append(entries, *e)
				}
			}
		} else {
			ollamaTag := d.Name + ":latest"
			if staticTags[ollamaTag] || seen[ollamaTag] {
				continue
			}
			seen[ollamaTag] = true
			if e := inferModelEntry(d.Name, "latest", d.Description); e != nil {
				entries = append(entries, *e)
			}
		}
	}
	return entries
}

// inferModelEntry builds a ModelEntry by guessing metadata from the name/tag.
func inferModelEntry(name, tag, description string) *ModelEntry {
	paramCount := inferParamCount(tag)
	if paramCount == 0 {
		paramCount = inferParamCount(name) // e.g. "smollm2" won't help, but "phi3:3.8b" will
	}
	if paramCount == 0 {
		paramCount = 7 // reasonable default
	}

	useCases := inferUseCases(name, description)
	if len(useCases) == 0 {
		return nil
	}

	vram := estimateVRAMFromParams(paramCount)
	quality := inferQuality(paramCount)

	displayName := titleCase(name) + " " + tag

	return &ModelEntry{
		Name:            displayName + " (discovered)",
		OllamaTag:       name + ":" + tag,
		Family:          name,
		ParameterSize:   formatParamSize(paramCount),
		ParameterCount:  paramCount,
		QuantLevel:      "Q4_K_M",
		DiskSizeGB:      vram * 0.55,
		EstimatedVRAMGB: vram,
		UseCases:        useCases,
		Quality:         quality,
		Description:     description,
	}
}

// inferParamCount extracts a parameter count (in billions) from a string
// like "8b", "70b-instruct", "135m", etc.
func inferParamCount(s string) float64 {
	lower := strings.ToLower(s)
	// Billions: "8b", "3.8b"
	reB := regexp.MustCompile(`(\d+\.?\d*)b`)
	if m := reB.FindStringSubmatch(lower); len(m) > 1 {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			return v
		}
	}
	// Millions: "135m", "360m"
	reM := regexp.MustCompile(`(\d+)m`)
	if m := reM.FindStringSubmatch(lower); len(m) > 1 {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			return v / 1000
		}
	}
	return 0
}

// inferUseCases guesses use cases from model name and description.
func inferUseCases(name, description string) []UseCase {
	lower := strings.ToLower(name + " " + description)

	switch {
	case containsAny(lower, "code", "coder", "starcoder", "codellama", "granite-code", "yi-coder", "deepseek-coder"):
		return []UseCase{UseCaseCode}
	case containsAny(lower, "embed", "embedding", "nomic", "mxbai", "snowflake", "arctic-embed"):
		return []UseCase{UseCaseEmbedding}
	case containsAny(lower, "llava", "vision", "internvl", "pixtral", "molmo"):
		return []UseCase{UseCaseVision}
	case containsAny(lower, "whisper", "bark", "audio", "speech"):
		return []UseCase{UseCaseAudio}
	case containsAny(lower, "stable-diffusion", "flux", "sdxl", "pixart"):
		return []UseCase{UseCaseImage}
	case containsAny(lower, "wan", "cogvideo", "ltx-video", "hunyuan-video"):
		return []UseCase{UseCaseVideo}
	case containsAny(lower, "dolphin", "wizard-vicuna", "nous-hermes", "unrestricted", "unrestricted"):
		return []UseCase{UseCaseUnrestricted}
	case containsAny(lower, "qwq", "deepseek-r1", "phi4-reasoning", "reasoning"):
		return []UseCase{UseCaseReasoning}
	default:
		return []UseCase{UseCaseChat}
	}
}

func containsAny(s string, terms ...string) bool {
	for _, t := range terms {
		if strings.Contains(s, t) {
			return true
		}
	}
	return false
}

// estimateVRAMFromParams gives a rough VRAM estimate for a Q4-quantised model.
func estimateVRAMFromParams(params float64) float64 {
	return params*0.7 + 1.0
}

// inferQuality assigns a quality score based on parameter count.
func inferQuality(params float64) int {
	switch {
	case params >= 70:
		return 85
	case params >= 30:
		return 75
	case params >= 14:
		return 65
	case params >= 7:
		return 50
	case params >= 3:
		return 35
	case params >= 1:
		return 20
	default:
		return 10
	}
}

func formatParamSize(params float64) string {
	if params < 1 {
		return fmt.Sprintf("%.0fM", params*1000)
	}
	if params == float64(int(params)) {
		return fmt.Sprintf("%.0fB", params)
	}
	return fmt.Sprintf("%.1fB", params)
}

// titleCase capitalises the first letter of each word.
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// ──────────────────────────────────────────────────────────────────────
// Disk cache (via shared cache.Store)
// ──────────────────────────────────────────────────────────────────────

type registryCachePayload struct {
	Models      []ModelEntry `json:"models"`
	LastRefresh time.Time    `json:"last_refresh"`
}

func (r *Registry) loadCache() {
	if r.store == nil {
		return
	}
	var c registryCachePayload
	if !r.store.Get(registryCacheKey, &c) {
		return
	}
	r.mu.Lock()
	r.dynamic = c.Models
	r.lastRefresh = c.LastRefresh
	r.mu.Unlock()
}

func (r *Registry) saveCache() {
	if r.store == nil {
		return
	}
	r.mu.RLock()
	c := registryCachePayload{
		Models:      r.dynamic,
		LastRefresh: r.lastRefresh,
	}
	r.mu.RUnlock()

	_ = r.store.Set(registryCacheKey, c, registryCacheTTL)
}
