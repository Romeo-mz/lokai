package models

import (
	"fmt"
	"strings"

	"github.com/romeo-mz/lokai/internal/hardware"
)

// PerformanceEstimate holds estimated generation speed and timing.
type PerformanceEstimate struct {
	TokensPerSecond float64 `json:"tokens_per_second"`
	TimeToFirstToken string `json:"time_to_first_token"` // e.g. "~0.5s"
	GenerationTime   string `json:"generation_time"`     // e.g. "~12s for 500 tokens"
	QualityRating    string `json:"quality_rating"`      // "Low", "Medium", "High", "Excellent"
	Notes            string `json:"notes,omitempty"`
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
