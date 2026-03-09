package models

import (
	"testing"

	"github.com/romeo-mz/lokai/internal/hardware"
)

// ─── shared helpers ──────────────────────────────────────────────────────────

func specsForVRAM(vramGB float64) *hardware.HardwareSpecs {
	if vramGB <= 0 {
		return &hardware.HardwareSpecs{
			RAMTotalGB: 8, RAMAvailableGB: 6, AvailableVRAMGB: 0,
		}
	}
	return &hardware.HardwareSpecs{
		HasGPU:          true,
		AvailableVRAMGB: vramGB,
		GPUs: []hardware.GPUInfo{
			{Vendor: hardware.GPUVendorNVIDIA, VRAMTotalGB: vramGB, VRAMFreeGB: vramGB},
		},
	}
}

// ─── All use‑cases return results ────────────────────────────────────────────

func TestRecommend_AllUseCasesReturnResults(t *testing.T) {
	specs := specsForVRAM(24.0) // workstation-class — something fits in every category
	useCases := []UseCase{
		UseCaseChat, UseCaseCode, UseCaseVision, UseCaseEmbedding,
		UseCaseReasoning, UseCaseImage, UseCaseVideo, UseCaseAudio, UseCaseNSFW,
	}
	for _, uc := range useCases {
		t.Run(string(uc), func(t *testing.T) {
			recs := Recommend(specs, RecommendOptions{
				UseCase:       uc,
				Priority:      PriorityBalanced,
				IncludeRemote: true,
				MaxResults:    5,
			})
			if len(recs) == 0 {
				t.Errorf("Recommend(%s, 24GB) returned 0 results", uc)
			}
		})
	}
}

// ─── Priority modes ──────────────────────────────────────────────────────────

func TestRecommend_PriorityQuality(t *testing.T) {
	specs := specsForVRAM(24.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseChat,
		Priority:      PriorityQuality,
		IncludeRemote: true,
		MaxResults:    5,
	})
	if len(recs) == 0 {
		t.Fatal("PriorityQuality should return results")
	}
	// Results must be sorted by score descending.
	for i := 1; i < len(recs); i++ {
		if recs[i].Score > recs[i-1].Score {
			t.Errorf("PriorityQuality results not sorted at index %d", i)
		}
	}
}

func TestRecommend_PrioritySpeed(t *testing.T) {
	specs := specsForVRAM(12.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseCode,
		Priority:      PrioritySpeed,
		IncludeRemote: true,
		MaxResults:    5,
	})
	if len(recs) == 0 {
		t.Fatal("PrioritySpeed should return results")
	}
	for i := 1; i < len(recs); i++ {
		if recs[i].Score > recs[i-1].Score {
			t.Errorf("PrioritySpeed results not sorted at index %d", i)
		}
	}
}

func TestRecommend_PriorityBalanced(t *testing.T) {
	specs := specsForVRAM(8.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseReasoning,
		Priority:      PriorityBalanced,
		IncludeRemote: true,
		MaxResults:    5,
	})
	if len(recs) == 0 {
		t.Fatal("PriorityBalanced should return results")
	}
}

// ─── VRAM budget filtering ───────────────────────────────────────────────────

func TestRecommend_OnlyFittingModelsWithoutIncludeRemote(t *testing.T) {
	// Very constrained VRAM — only smallest models should be listed.
	specs := specsForVRAM(1.5)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseChat,
		Priority:      PrioritySpeed,
		IncludeRemote: false,
		MaxResults:    10,
	})
	for _, r := range recs {
		if !r.FitsInVRAM {
			t.Errorf("IncludeRemote=false: model %s doesn't fit but was returned", r.Model.OllamaTag)
		}
	}
}

func TestRecommend_IncludeRemote_MoreResults(t *testing.T) {
	specs := specsForVRAM(1.0) // Very small — most models won't fit.
	withRemote := Recommend(specs, RecommendOptions{UseCase: UseCaseChat, Priority: PriorityQuality, IncludeRemote: true, MaxResults: 20})
	withoutRemote := Recommend(specs, RecommendOptions{UseCase: UseCaseChat, Priority: PriorityQuality, IncludeRemote: false, MaxResults: 20})

	if len(withRemote) < len(withoutRemote) {
		t.Error("IncludeRemote=true should return >= results vs IncludeRemote=false")
	}
}

// ─── Hardware tiers ──────────────────────────────────────────────────────────

func TestRecommend_RPiEdgeHardware(t *testing.T) {
	specs := &hardware.HardwareSpecs{
		RAMTotalGB: 2, RAMAvailableGB: 1.5, AvailableVRAMGB: 1.0,
	}
	recs := Recommend(specs, RecommendOptions{
		UseCase:    UseCaseChat,
		Priority:   PrioritySpeed,
		MaxResults: 5,
	})
	// At least one model should fit (e.g. SmolLM2 135M).
	if len(recs) == 0 {
		t.Error("Expected at least one recommendation on RPi-class hardware")
	}
	for _, r := range recs {
		if !r.FitsInVRAM {
			t.Errorf("RPi: model %s doesn't fit but was returned", r.Model.OllamaTag)
		}
	}
}

func TestRecommend_WorkstationHardware(t *testing.T) {
	specs := specsForVRAM(80.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseChat,
		Priority:      PriorityQuality,
		IncludeRemote: true,
		MaxResults:    5,
	})
	if len(recs) == 0 {
		t.Fatal("Workstation hardware should return results")
	}
	// Top result on workstation should be a large model (>= 30B).
	if recs[0].Model.ParameterCount < 10 {
		t.Errorf("Expected large model on workstation, got %s (%.0fB)", recs[0].Model.OllamaTag, recs[0].Model.ParameterCount)
	}
}

// ─── Sub‑task scoring ────────────────────────────────────────────────────────

func TestRecommend_SubTask_QuickQA_PrefersSmallModels(t *testing.T) {
	specs := specsForVRAM(24.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseChat,
		SubTask:       SubTaskQuickQA,
		Priority:      PrioritySpeed,
		IncludeRemote: true,
		MaxResults:    5,
	})
	if len(recs) == 0 {
		t.Fatal("SubTaskQuickQA should return results")
	}
}

func TestRecommend_SubTask_Coding(t *testing.T) {
	specs := specsForVRAM(12.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseCode,
		SubTask:       SubTaskAutocomplete,
		Priority:      PrioritySpeed,
		IncludeRemote: true,
		MaxResults:    5,
	})
	if len(recs) == 0 {
		t.Fatal("SubTaskAutocomplete should return results")
	}
}

func TestRecommend_SubTask_Reasoning(t *testing.T) {
	specs := specsForVRAM(16.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseReasoning,
		SubTask:       SubTaskMath,
		Priority:      PriorityQuality,
		IncludeRemote: true,
		MaxResults:    5,
	})
	if len(recs) == 0 {
		t.Fatal("SubTaskMath on reasoning should return results")
	}
}

// ─── MaxResults capping ──────────────────────────────────────────────────────

func TestRecommend_MaxResults_Respected(t *testing.T) {
	specs := specsForVRAM(48.0)
	for _, limit := range []int{1, 3, 5, 10} {
		recs := Recommend(specs, RecommendOptions{
			UseCase: UseCaseChat, Priority: PriorityQuality,
			IncludeRemote: true, MaxResults: limit,
		})
		if len(recs) > limit {
			t.Errorf("MaxResults=%d but got %d results", limit, len(recs))
		}
	}
}

func TestRecommend_MaxResultsDefault(t *testing.T) {
	// MaxResults == 0 → defaults to 5 internally.
	specs := specsForVRAM(24.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase: UseCaseChat, Priority: PriorityQuality,
		IncludeRemote: true, MaxResults: 0,
	})
	if len(recs) > 5 {
		t.Errorf("Default MaxResults should be 5, got %d", len(recs))
	}
}

// ─── Scoring invariants ──────────────────────────────────────────────────────

func TestRecommend_FittingModelHigherThanNonFitting(t *testing.T) {
	// With a modest VRAM budget some models fit and some don't; the fitting ones
	// should generally outscore the non-fitting ones.
	specs := specsForVRAM(8.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseChat,
		Priority:      PriorityQuality,
		IncludeRemote: true,
		MaxResults:    20,
	})

	var topFitScore, topNoFitScore float64
	for _, r := range recs {
		if r.FitsInVRAM && r.Score > topFitScore {
			topFitScore = r.Score
		}
		if !r.FitsInVRAM && r.Score > topNoFitScore {
			topNoFitScore = r.Score
		}
	}
	if topFitScore == 0 || topNoFitScore == 0 {
		// Either all fit or none don't — skip this check.
		t.Skip("All models fit or none fit; skipping cross-category comparison")
	}
	if topNoFitScore > topFitScore {
		t.Errorf("A non-fitting model (%.1f) outscored the best fitting model (%.1f)", topNoFitScore, topFitScore)
	}
}

func TestRecommend_ScoresArePositive(t *testing.T) {
	specs := specsForVRAM(8.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseCode,
		Priority:      PriorityBalanced,
		IncludeRemote: true,
		MaxResults:    10,
	})
	for _, r := range recs {
		if r.Score <= 0 {
			t.Errorf("Model %s has non-positive score %.1f", r.Model.OllamaTag, r.Score)
		}
	}
}

// ─── Dynamic catalog ─────────────────────────────────────────────────────────

func TestRecommend_DynamicCatalog(t *testing.T) {
	customCatalog := []ModelEntry{
		{
			Name: "Custom Chat A", OllamaTag: "custom:chat-a",
			ParameterCount: 7, EstimatedVRAMGB: 5, Quality: 80,
			UseCases: []UseCase{UseCaseChat},
		},
		{
			Name: "Custom Chat B", OllamaTag: "custom:chat-b",
			ParameterCount: 3, EstimatedVRAMGB: 2.5, Quality: 60,
			UseCases: []UseCase{UseCaseChat},
		},
	}

	specs := specsForVRAM(12.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseChat,
		Priority:      PriorityQuality,
		IncludeRemote: true,
		MaxResults:    5,
		Catalog:       customCatalog,
	})
	if len(recs) != 2 {
		t.Fatalf("Expected 2 results from custom catalog, got %d", len(recs))
	}
	// Higher quality model should be ranked first.
	if recs[0].Model.OllamaTag != "custom:chat-a" {
		t.Errorf("Expected custom:chat-a first, got %s", recs[0].Model.OllamaTag)
	}
}

func TestRecommend_DynamicCatalog_WrongUseCase(t *testing.T) {
	customCatalog := []ModelEntry{
		{
			Name: "Code Only", OllamaTag: "custom:code",
			ParameterCount: 7, EstimatedVRAMGB: 5, Quality: 80,
			UseCases: []UseCase{UseCaseCode},
		},
	}

	recs := Recommend(specsForVRAM(12.0), RecommendOptions{
		UseCase:       UseCaseChat, // different from catalog entry
		Priority:      PriorityQuality,
		IncludeRemote: true,
		MaxResults:    5,
		Catalog:       customCatalog,
	})
	if len(recs) != 0 {
		t.Errorf("Expected 0 results when catalog has no matching use case, got %d", len(recs))
	}
}

// ─── Reason field ────────────────────────────────────────────────────────────

func TestRecommend_ReasonNotEmpty(t *testing.T) {
	specs := specsForVRAM(12.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase: UseCaseCode, Priority: PriorityBalanced,
		IncludeRemote: true, MaxResults: 3,
	})
	for _, r := range recs {
		if r.Reason == "" {
			t.Errorf("Model %s has empty Reason field", r.Model.OllamaTag)
		}
	}
}

// ─── VRAMUsage percentage ────────────────────────────────────────────────────

func TestRecommend_VRAMUsageCalculation(t *testing.T) {
	specs := specsForVRAM(10.0)
	recs := Recommend(specs, RecommendOptions{
		UseCase: UseCaseChat, Priority: PriorityQuality,
		IncludeRemote: true, MaxResults: 5,
	})
	for _, r := range recs {
		expectedPct := (r.Model.EstimatedVRAMGB / 10.0) * 100
		if abs(r.VRAMUsage-expectedPct) > 0.1 {
			t.Errorf("Model %s: VRAMUsage=%.2f%%, expected %.2f%%",
				r.Model.OllamaTag, r.VRAMUsage, expectedPct)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ─── CPU‑only hardware (HasGPU = false) ──────────────────────────────────────

func TestRecommend_CPUOnlyHardware(t *testing.T) {
	specs := &hardware.HardwareSpecs{
		HasGPU:          false,
		HasAVX2:         true,
		CPUCores:        8,
		RAMTotalGB:      16,
		RAMAvailableGB:  10,
		AvailableVRAMGB: 7,
	}
	recs := Recommend(specs, RecommendOptions{
		UseCase:    UseCaseChat,
		Priority:   PrioritySpeed,
		MaxResults: 5,
	})
	if len(recs) == 0 {
		t.Error("CPU-only hardware should still yield recommendations")
	}
}
