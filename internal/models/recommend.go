// Package models — recommendation engine.
//
// Sources:
//
//	Static catalog  — internal/models/database.go (hand-curated)
//	Ollama Registry — https://registry.ollama.ai/v2/library (dynamic)
//	Ollama Library  — https://ollama.com/search (dynamic)
//	GitHub API      — https://api.github.com/search/repositories (dynamic)
package models

import (
	"sort"
	"strings"

	"github.com/romeo-mz/lokai/internal/hardware"
)

// SubTask represents a specific task within a use case for finer recommendations.
type SubTask string

const (
	// Chat sub-tasks.
	SubTaskQuickQA       SubTask = "quick_qa"
	SubTaskWriting       SubTask = "writing"
	SubTaskCreative      SubTask = "creative"
	SubTaskTranslation   SubTask = "translation"
	SubTaskSummarization SubTask = "summarization"

	// Code sub-tasks.
	SubTaskAutocomplete  SubTask = "autocomplete"
	SubTaskProjectGen    SubTask = "project_gen"
	SubTaskCodeReview    SubTask = "code_review"
	SubTaskDebugging     SubTask = "debugging"
	SubTaskDocumentation SubTask = "documentation"

	// Vision sub-tasks.
	SubTaskPhotoAnalysis SubTask = "photo_analysis"
	SubTaskOCR           SubTask = "ocr"
	SubTaskChartAnalysis SubTask = "chart_analysis"

	// Reasoning sub-tasks.
	SubTaskMath     SubTask = "math"
	SubTaskLogic    SubTask = "logic"
	SubTaskPlanning SubTask = "planning"
	SubTaskResearch SubTask = "research"

	// Image gen sub-tasks.
	SubTaskPhotorealistic SubTask = "photorealistic"
	SubTaskArtistic       SubTask = "artistic"
	SubTaskFastDrafts     SubTask = "fast_drafts"

	// Audio sub-tasks.
	SubTaskTranscription SubTask = "transcription"
	SubTaskTTS           SubTask = "tts"
)

// Recommendation represents a recommended model with scoring details.
type Recommendation struct {
	Model      ModelEntry `json:"model"`
	Score      float64    `json:"score"` // 0-100 composite score
	FitsInVRAM bool       `json:"fits_in_vram"`
	VRAMUsage  float64    `json:"vram_usage_pct"` // Percentage of available VRAM used
	Reason     string     `json:"reason"`         // Human-readable explanation
}

// RecommendOptions configures the recommendation engine.
type RecommendOptions struct {
	UseCase       UseCase
	SubTask       SubTask
	Priority      Priority
	IncludeRemote bool         // Include models not yet downloaded
	MaxResults    int          // Max recommendations to return (default: 5)
	Catalog       []ModelEntry // Optional dynamic catalog (from Registry); nil = static
}

// Recommend returns the best models for the given hardware and preferences.
func Recommend(specs *hardware.HardwareSpecs, opts RecommendOptions) []Recommendation {
	if opts.MaxResults <= 0 {
		opts.MaxResults = 5
	}

	// Use provided dynamic catalog or fall back to the static one.
	var candidates []ModelEntry
	if len(opts.Catalog) > 0 {
		for _, m := range opts.Catalog {
			for _, uc := range m.UseCases {
				if uc == opts.UseCase {
					candidates = append(candidates, m)
					break
				}
			}
		}
	} else {
		candidates = GetModelsByUseCase(opts.UseCase)
	}
	if len(candidates) == 0 {
		return nil
	}

	availableVRAM := specs.AvailableVRAMGB
	// Safety margin: use 90% of available VRAM.
	vramBudget := availableVRAM * 0.9

	var recommendations []Recommendation
	for _, model := range candidates {
		fits := model.EstimatedVRAMGB <= vramBudget
		if !fits && !opts.IncludeRemote {
			// Skip models that don't fit if user only wants what works.
			// But still include if they opted to see all options.
			continue
		}

		score := computeScore(model, specs, opts.Priority, opts.SubTask, fits)
		usage := 0.0
		if availableVRAM > 0 {
			usage = (model.EstimatedVRAMGB / availableVRAM) * 100
		}

		reason := buildReason(model, specs, fits)

		recommendations = append(recommendations, Recommendation{
			Model:      model,
			Score:      score,
			FitsInVRAM: fits,
			VRAMUsage:  usage,
			Reason:     reason,
		})
	}

	// Sort by score descending.
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	// Limit results.
	if len(recommendations) > opts.MaxResults {
		recommendations = recommendations[:opts.MaxResults]
	}

	return recommendations
}

// computeScore assigns a composite score based on model quality, fit, and priority.
func computeScore(model ModelEntry, specs *hardware.HardwareSpecs, priority Priority, subTask SubTask, fits bool) float64 {
	score := float64(model.Quality) // Base: quality score (0-100)

	if !fits {
		// Heavy penalty for models that don't fit in VRAM.
		score *= 0.3
	}

	switch priority {
	case PriorityQuality:
		// Favor larger, higher-quality models.
		score *= 1.0

	case PrioritySpeed:
		// Favor smaller models (faster inference).
		// Bonus for models that use less VRAM (more headroom = faster).
		if specs.AvailableVRAMGB > 0 {
			usageRatio := model.EstimatedVRAMGB / specs.AvailableVRAMGB
			speedBonus := (1.0 - usageRatio) * 30 // Up to 30 points for low usage
			score += speedBonus
		}
		// Penalty for very large models.
		if model.ParameterCount > 30 {
			score *= 0.7
		}

	case PriorityBalanced:
		// Slight bonus for models using 40-70% of VRAM (sweet spot).
		if specs.AvailableVRAMGB > 0 {
			usageRatio := model.EstimatedVRAMGB / specs.AvailableVRAMGB
			if usageRatio >= 0.4 && usageRatio <= 0.7 {
				score += 10
			}
		}
	}

	// AVX2 bonus for CPU inference with larger models.
	if specs.HasAVX2 && !specs.HasGPU && model.ParameterCount <= 14 {
		score += 5
	}

	// Sub-task bonuses: tailor scoring to specific user needs.
	score += subTaskBonus(model, subTask)

	return score
}

// subTaskBonus returns a score adjustment based on how well a model fits a sub-task.
func subTaskBonus(model ModelEntry, subTask SubTask) float64 {
	if subTask == "" {
		return 0
	}

	nameLower := strings.ToLower(model.Name)
	familyLower := strings.ToLower(model.Family)
	descLower := strings.ToLower(model.Description)
	hasCap := func(cap string) bool {
		for _, c := range model.Capabilities {
			if c == cap {
				return true
			}
		}
		return false
	}

	switch subTask {
	// ── Chat sub-tasks ──
	case SubTaskQuickQA:
		// Prefer small, fast models.
		if model.ParameterCount <= 3 {
			return 15
		}
		if model.ParameterCount <= 8 {
			return 5
		}
		if model.ParameterCount > 30 {
			return -10
		}

	case SubTaskWriting:
		// Prefer mid-to-large models with good language quality.
		if model.ParameterCount >= 8 && model.ParameterCount <= 14 {
			return 12
		}
		if model.ParameterCount >= 14 {
			return 8
		}

	case SubTaskCreative:
		// Prefer larger models, bonus for uncensored/creative families.
		if model.ParameterCount >= 14 {
			return 10
		}
		if model.Quality >= 60 {
			return 8
		}

	case SubTaskTranslation:
		// Prefer multilingual models (Qwen, Gemma).
		if familyLower == "qwen" || familyLower == "gemma" {
			return 15
		}
		if model.ParameterCount >= 8 {
			return 5
		}

	case SubTaskSummarization:
		// Prefer models with good context handling.
		if hasCap("tools") {
			return 5
		}
		if model.ParameterCount >= 8 && model.ParameterCount <= 14 {
			return 10
		}

	// ── Code sub-tasks ──
	case SubTaskAutocomplete:
		// Prefer small, fast code models.
		if model.ParameterCount <= 3 {
			return 18
		}
		if model.ParameterCount <= 8 {
			return 8
		}
		if model.ParameterCount > 14 {
			return -8
		}

	case SubTaskProjectGen:
		// Prefer large, high-quality code models.
		if model.ParameterCount >= 14 {
			return 15
		}
		if model.Quality >= 70 {
			return 10
		}
		if model.ParameterCount <= 3 {
			return -10
		}

	case SubTaskCodeReview:
		// Prefer models with reasoning and tools capability.
		if hasCap("thinking") {
			return 12
		}
		if model.ParameterCount >= 8 {
			return 8
		}

	case SubTaskDebugging:
		// Prefer mid-large reasoning-capable code models.
		if hasCap("thinking") {
			return 10
		}
		if model.ParameterCount >= 8 && model.ParameterCount <= 35 {
			return 8
		}

	case SubTaskDocumentation:
		// Balanced models work well for doc-gen.
		if model.ParameterCount >= 7 && model.ParameterCount <= 14 {
			return 10
		}

	// ── Vision sub-tasks ──
	case SubTaskPhotoAnalysis:
		// Prefer vision-capable models with high quality.
		if model.Quality >= 70 {
			return 10
		}

	case SubTaskOCR:
		// Prefer smaller, faster vision models.
		if model.ParameterCount <= 8 {
			return 12
		}

	case SubTaskChartAnalysis:
		// Prefer larger vision models with better reasoning.
		if model.ParameterCount >= 12 {
			return 12
		}
		if hasCap("thinking") {
			return 8
		}

	// ── Reasoning sub-tasks ──
	case SubTaskMath:
		// Prefer models with thinking capability.
		if hasCap("thinking") {
			return 15
		}
		if strings.Contains(nameLower, "qwq") || strings.Contains(nameLower, "phi-4") {
			return 10
		}

	case SubTaskLogic:
		// Similar to math but any reasoning model works.
		if hasCap("thinking") {
			return 12
		}

	case SubTaskPlanning:
		// Prefer larger context models.
		if model.ParameterCount >= 14 {
			return 12
		}
		if hasCap("tools") {
			return 8
		}

	case SubTaskResearch:
		// Prefer high-quality, large models.
		if model.Quality >= 70 {
			return 15
		}
		if model.ParameterCount >= 14 {
			return 8
		}

	// ── Image gen sub-tasks ──
	case SubTaskPhotorealistic:
		// FLUX and SD 3.5 are better for photorealism.
		if strings.Contains(nameLower, "flux") || strings.Contains(nameLower, "sd 3.5") {
			return 15
		}

	case SubTaskArtistic:
		// SDXL and PixArt are good for artistic styles.
		if strings.Contains(nameLower, "sdxl") || strings.Contains(nameLower, "pixart") {
			return 12
		}
		if strings.Contains(nameLower, "flux") {
			return 8
		}

	case SubTaskFastDrafts:
		// Prefer smaller/faster image models.
		if strings.Contains(descLower, "fast") || strings.Contains(nameLower, "schnell") {
			return 15
		}
		if model.EstimatedVRAMGB <= 8 {
			return 8
		}

	// ── Audio sub-tasks ──
	case SubTaskTranscription:
		// Prefer Whisper models.
		if strings.Contains(nameLower, "whisper") {
			return 15
		}

	case SubTaskTTS:
		// Prefer Bark / TTS models.
		if strings.Contains(nameLower, "bark") {
			return 15
		}
	}

	return 0
}

// buildReason generates a human-readable explanation for the recommendation.
func buildReason(model ModelEntry, specs *hardware.HardwareSpecs, fits bool) string {
	if !fits {
		return "⚠ Exceeds available VRAM — may run slowly or fail to load"
	}

	if specs.AvailableVRAMGB > 0 {
		usage := (model.EstimatedVRAMGB / specs.AvailableVRAMGB) * 100
		switch {
		case usage < 40:
			return "✓ Lightweight — fast inference with VRAM to spare"
		case usage < 70:
			return "✓ Great fit — good balance of quality and performance"
		case usage < 90:
			return "✓ Tight fit — uses most of your VRAM for best quality"
		default:
			return "✓ Maximum quality — uses nearly all available VRAM"
		}
	}

	return "✓ Compatible with your hardware"
}
