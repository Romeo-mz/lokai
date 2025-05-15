package models

import (
	"sort"

	"github.com/romeo-mz/lokai/internal/hardware"
)

// Recommendation represents a recommended model with scoring details.
type Recommendation struct {
	Model       ModelEntry `json:"model"`
	Score       float64    `json:"score"`        // 0-100 composite score
	FitsInVRAM  bool       `json:"fits_in_vram"`
	VRAMUsage   float64    `json:"vram_usage_pct"` // Percentage of available VRAM used
	Reason      string     `json:"reason"`         // Human-readable explanation
}

// RecommendOptions configures the recommendation engine.
type RecommendOptions struct {
	UseCase          UseCase
	Priority         Priority
	IncludeRemote    bool // Include models not yet downloaded
	MaxResults       int  // Max recommendations to return (default: 5)
}

// Recommend returns the best models for the given hardware and preferences.
func Recommend(specs *hardware.HardwareSpecs, opts RecommendOptions) []Recommendation {
	if opts.MaxResults <= 0 {
		opts.MaxResults = 5
	}

	// Get all models for this use case.
	candidates := GetModelsByUseCase(opts.UseCase)
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

		score := computeScore(model, specs, opts.Priority, fits)
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
func computeScore(model ModelEntry, specs *hardware.HardwareSpecs, priority Priority, fits bool) float64 {
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

	return score
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
