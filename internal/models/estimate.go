// Package models — performance estimation.
//
// Estimates token generation speed, time-to-first-token, and
// human-friendly task durations based on hardware specs.
//
// Sources:
//
//	GPU benchmarks  — https://github.com/XiongjieDai/GPU-Benchmarks-on-LLM-Inference
//	Ollama perf FAQ — https://github.com/ollama/ollama/blob/main/docs/faq.md
package models

import (
	"fmt"
	"strings"

	"github.com/romeo-mz/lokai/internal/hardware"
)

// PerformanceEstimate holds estimated generation speed and timing.
type PerformanceEstimate struct {
	TokensPerSecond  float64 `json:"tokens_per_second"`
	TimeToFirstToken string  `json:"time_to_first_token"` // e.g. "~0.5s"
	GenerationTime   string  `json:"generation_time"`     // e.g. "~12s for 500 tokens"
	QualityRating    string  `json:"quality_rating"`      // "Low", "Medium", "High", "Excellent"
	Notes            string  `json:"notes,omitempty"`
}

// EstimatePerformance predicts generation speed for a model on given hardware.
func EstimatePerformance(model ModelEntry, specs *hardware.HardwareSpecs) PerformanceEstimate {
	tokPerSec := estimateTokensPerSecond(model, specs)

	est := PerformanceEstimate{
		TokensPerSecond:  tokPerSec,
		TimeToFirstToken: estimateTimeToFirstToken(model, specs),
		GenerationTime:   estimateGenerationTime(tokPerSec),
		QualityRating:    qualityRating(model.Quality),
	}

	// Add notes for special cases.
	var notes []string
	if !specs.HasGPU && !specs.IsAppleSilicon {
		notes = append(notes, "CPU-only inference — significantly slower than GPU")
	}
	if specs.HasAVX2 {
		notes = append(notes, "AVX2 detected — optimized CPU inference available")
	}
	if model.EstimatedVRAMGB > specs.AvailableVRAMGB*0.85 {
		notes = append(notes, "Model uses most of your VRAM — consider closing other GPU apps")
	}
	if len(notes) > 0 {
		est.Notes = strings.Join(notes, "; ")
	}

	return est
}

// estimateTokensPerSecond provides a rough token generation speed estimate.
// Based on known benchmarks for common GPU/model combinations.
func estimateTokensPerSecond(model ModelEntry, specs *hardware.HardwareSpecs) float64 {
	params := model.ParameterCount

	// Base speed by GPU tier (tokens/sec for a 7B Q4 model).
	var baseTokPerSec float64

	if specs.IsAppleSilicon {
		// Apple Silicon unified memory bandwidth-based estimates.
		switch {
		case specs.UnifiedMemoryGB >= 64: // M2/M3 Ultra
			baseTokPerSec = 45
		case specs.UnifiedMemoryGB >= 32: // M2/M3 Max
			baseTokPerSec = 35
		case specs.UnifiedMemoryGB >= 16: // M2/M3 Pro or base with 16GB
			baseTokPerSec = 25
		default: // M1/M2 base 8GB
			baseTokPerSec = 18
		}
	} else if specs.HasGPU && len(specs.GPUs) > 0 {
		gpu := specs.GPUs[0]
		switch {
		case gpu.VRAMTotalGB >= 24: // RTX 4090, A6000
			baseTokPerSec = 60
		case gpu.VRAMTotalGB >= 16: // RTX 4080, A5000
			baseTokPerSec = 45
		case gpu.VRAMTotalGB >= 12: // RTX 3080, RTX 4070
			baseTokPerSec = 35
		case gpu.VRAMTotalGB >= 8: // RTX 3060, RTX 4060
			baseTokPerSec = 28
		case gpu.VRAMTotalGB >= 6: // GTX 1660
			baseTokPerSec = 20
		default: // Low-end GPU
			baseTokPerSec = 12
		}
	} else {
		// CPU-only: heavily dependent on core count and AVX.
		baseTokPerSec = float64(specs.CPUCores) * 1.2
		if specs.HasAVX2 {
			baseTokPerSec *= 1.5
		}
		if specs.HasAVX512 {
			baseTokPerSec *= 1.8
		}
		if baseTokPerSec > 15 {
			baseTokPerSec = 15 // CPU inference caps out around here.
		}
	}

	// Scale inversely with parameter count (relative to 7B baseline).
	if params > 0 {
		scaleFactor := 7.0 / params
		baseTokPerSec *= scaleFactor
	}

	// Floor: at least 1 tok/sec.
	if baseTokPerSec < 1 {
		baseTokPerSec = 1
	}

	// Round to 1 decimal.
	return float64(int(baseTokPerSec*10)) / 10
}

// estimateTimeToFirstToken estimates time to first token based on model loading.
func estimateTimeToFirstToken(model ModelEntry, specs *hardware.HardwareSpecs) string {
	if specs.HasGPU || specs.IsAppleSilicon {
		if model.ParameterCount <= 14 {
			return "~0.5s"
		}
		return "~1-2s"
	}
	// CPU inference: slower prompt processing.
	if model.ParameterCount <= 3 {
		return "~1s"
	}
	if model.ParameterCount <= 8 {
		return "~2-3s"
	}
	return "~5-10s"
}

// estimateGenerationTime estimates time to generate a typical response.
func estimateGenerationTime(tokPerSec float64) string {
	if tokPerSec <= 0 {
		return "unknown"
	}

	// Assume typical response is ~200 tokens.
	typicalTokens := 200.0
	seconds := typicalTokens / tokPerSec

	if seconds < 5 {
		return fmt.Sprintf("~%.0fs for %d tokens", seconds, int(typicalTokens))
	}
	if seconds < 60 {
		return fmt.Sprintf("~%.0fs for %d tokens", seconds, int(typicalTokens))
	}
	minutes := seconds / 60
	return fmt.Sprintf("~%.1f min for %d tokens", minutes, int(typicalTokens))
}

// RealWorldTasks returns human-friendly task descriptions with estimated times
// tailored to the selected use case.
func RealWorldTasks(useCase UseCase, tokPerSec float64) []string {
	if tokPerSec <= 0 {
		return nil
	}
	fmtTime := func(tokens float64) string {
		secs := tokens / tokPerSec
		if secs < 1 {
			return "under 1 second"
		}
		if secs < 60 {
			return fmt.Sprintf("~%.0f seconds", secs)
		}
		mins := secs / 60
		if mins < 60 {
			return fmt.Sprintf("~%.1f minutes", mins)
		}
		return fmt.Sprintf("~%.1f hours", mins/60)
	}

	switch useCase {
	case UseCaseChat:
		return []string{
			"Quick Q&A answer: " + fmtTime(80),
			"Write a 500-word email: " + fmtTime(700),
			"Summarize a long article: " + fmtTime(300),
			"Brainstorm 10 ideas: " + fmtTime(400),
		}
	case UseCaseCode:
		return []string{
			"Generate a function (~30 lines): " + fmtTime(200),
			"Build a portfolio HTML page: " + fmtTime(800),
			"Write unit tests for a class: " + fmtTime(500),
			"Refactor & explain legacy code: " + fmtTime(600),
		}
	case UseCaseVision:
		return []string{
			"Describe what's in a photo: " + fmtTime(150),
			"Read text from a screenshot (OCR): " + fmtTime(100),
			"Analyze a chart or diagram: " + fmtTime(200),
		}
	case UseCaseEmbedding:
		return []string{
			"Embed a single document: under 1 second",
			"Index 100 documents for search: " + fmtTime(5000),
			"Build a knowledge base (1K docs): " + fmtTime(50000),
		}
	case UseCaseReasoning:
		return []string{
			"Solve a math equation: " + fmtTime(300),
			"Walk through a logic puzzle: " + fmtTime(500),
			"Plan a multi-step project: " + fmtTime(800),
		}
	case UseCaseImage:
		return []string{
			"Generate one image (512×512): ~10-30 seconds (depends on pipeline)",
			"Generate a high-res image (1024×1024): ~30-90 seconds",
			"Create 4 variations of a concept: ~1-3 minutes",
		}
	case UseCaseVideo:
		return []string{
			"Generate a 2-second clip: ~2-10 minutes (depends on resolution)",
			"Generate a 4-second clip: ~5-20 minutes",
			"Higher resolution = significantly longer",
		}
	case UseCaseAudio:
		return []string{
			"Transcribe 1 minute of audio: ~5-15 seconds",
			"Transcribe a 30-minute meeting: ~3-8 minutes",
			"Generate a spoken sentence (TTS): ~2-5 seconds",
		}
	case UseCaseNSFW:
		return []string{
			"Quick reply: " + fmtTime(80),
			"Generate a short story (~500 words): " + fmtTime(700),
			"Long creative writing (~2000 words): " + fmtTime(2800),
		}
	}
	return nil
}

// qualityRating converts a quality score to a human-readable rating.
func qualityRating(quality int) string {
	switch {
	case quality >= 85:
		return "Excellent"
	case quality >= 65:
		return "High"
	case quality >= 40:
		return "Medium"
	default:
		return "Low"
	}
}
