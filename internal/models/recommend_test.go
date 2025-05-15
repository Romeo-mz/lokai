package models

import (
	"testing"

	"github.com/romeo-mz/lokai/internal/hardware"
)

func TestGetModelsByUseCase(t *testing.T) {
	tests := []struct {
		useCase  UseCase
		minCount int
	}{
		{UseCaseChat, 10},
		{UseCaseCode, 8},
		{UseCaseVision, 5},
		{UseCaseEmbedding, 3},
		{UseCaseReasoning, 5},
		{UseCaseImage, 3},
		{UseCaseVideo, 3},
		{UseCaseAudio, 3},
		{UseCaseNSFW, 3},
	}

	for _, tt := range tests {
		t.Run(string(tt.useCase), func(t *testing.T) {
			models := GetModelsByUseCase(tt.useCase)
			if len(models) < tt.minCount {
				t.Errorf("GetModelsByUseCase(%s) returned %d models, want at least %d", tt.useCase, len(models), tt.minCount)
			}
		})
	}
}

func TestGetModelByTag(t *testing.T) {
	m := GetModelByTag("llama3.1:8b")
	if m == nil {
		t.Fatal("GetModelByTag('llama3.1:8b') returned nil")
	}
	if m.Family != "llama" {
		t.Errorf("expected family 'llama', got %q", m.Family)
	}

	m = GetModelByTag("nonexistent:model")
	if m != nil {
		t.Errorf("GetModelByTag('nonexistent:model') should return nil, got %+v", m)
	}
}

func TestCatalogNoDuplicateTags(t *testing.T) {
	seen := make(map[string]bool)
	for _, m := range Catalog {
		if seen[m.OllamaTag] {
			t.Errorf("duplicate OllamaTag: %s", m.OllamaTag)
		}
		seen[m.OllamaTag] = true
	}
}

func TestCatalogValidFields(t *testing.T) {
	for _, m := range Catalog {
		if m.Name == "" {
			t.Errorf("model with tag %q has empty Name", m.OllamaTag)
		}
		if m.OllamaTag == "" {
			t.Errorf("model %q has empty OllamaTag", m.Name)
		}
		if m.Quality < 1 || m.Quality > 100 {
			t.Errorf("model %q has Quality=%d, want 1-100", m.OllamaTag, m.Quality)
		}
		if m.ParameterCount <= 0 {
			t.Errorf("model %q has ParameterCount=%.3f, want > 0", m.OllamaTag, m.ParameterCount)
		}
		if m.EstimatedVRAMGB <= 0 {
			t.Errorf("model %q has EstimatedVRAMGB=%.1f, want > 0", m.OllamaTag, m.EstimatedVRAMGB)
		}
		if len(m.UseCases) == 0 {
			t.Errorf("model %q has no UseCases", m.OllamaTag)
		}
	}
}

func TestRecommendBasic(t *testing.T) {
	specs := &hardware.HardwareSpecs{
		AvailableVRAMGB: 12.0,
		HasGPU:          true,
		GPUs: []hardware.GPUInfo{
			{Vendor: hardware.GPUVendorNVIDIA, VRAMTotalGB: 12.0, VRAMFreeGB: 11.5},
		},
	}

	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseCode,
		Priority:      PriorityBalanced,
		IncludeRemote: true,
		MaxResults:    5,
	})

	if len(recs) == 0 {
		t.Fatal("Recommend returned no results for 12GB VRAM + code use case")
	}

	// Top recommendation should fit in VRAM.
	if !recs[0].FitsInVRAM {
		t.Error("Top recommendation should fit in VRAM")
	}

	// Score should be positive.
	if recs[0].Score <= 0 {
		t.Errorf("Top recommendation score should be positive, got %.1f", recs[0].Score)
	}

	// Results should be sorted by score descending.
	for i := 1; i < len(recs); i++ {
		if recs[i].Score > recs[i-1].Score {
			t.Errorf("Results not sorted: recs[%d].Score=%.1f > recs[%d].Score=%.1f", i, recs[i].Score, i-1, recs[i-1].Score)
		}
	}
}

func TestRecommendTinyHardware(t *testing.T) {
	specs := &hardware.HardwareSpecs{
		AvailableVRAMGB: 1.5,
		RAMTotalGB:      2.0,
		RAMAvailableGB:  1.5,
	}

	recs := Recommend(specs, RecommendOptions{
		UseCase:    UseCaseChat,
		Priority:   PrioritySpeed,
		MaxResults: 5,
	})

	if len(recs) == 0 {
		t.Fatal("Recommend should return results even for tiny hardware")
	}

	// Every returned model should fit.
	for _, r := range recs {
		if !r.FitsInVRAM {
			t.Errorf("Model %s doesn't fit but was returned without IncludeRemote", r.Model.OllamaTag)
		}
	}
}

func TestRecommendMaxResults(t *testing.T) {
	specs := &hardware.HardwareSpecs{
		AvailableVRAMGB: 48.0,
		HasGPU:          true,
	}

	recs := Recommend(specs, RecommendOptions{
		UseCase:       UseCaseChat,
		Priority:      PriorityQuality,
		IncludeRemote: true,
		MaxResults:    3,
	})

	if len(recs) > 3 {
		t.Errorf("MaxResults=3 but got %d results", len(recs))
	}
}

func TestRecommendNoModels(t *testing.T) {
	specs := &hardware.HardwareSpecs{AvailableVRAMGB: 100.0}

	// Use a use case that has no models (shouldn't exist but test the path).
	recs := Recommend(specs, RecommendOptions{
		UseCase:    UseCase("nonexistent"),
		Priority:   PriorityBalanced,
		MaxResults: 5,
	})

	if len(recs) != 0 {
		t.Errorf("Expected 0 results for nonexistent use case, got %d", len(recs))
	}
}

func TestEstimatePerformance(t *testing.T) {
	model := ModelEntry{
		Name: "Test Model", ParameterCount: 7.0,
		Quality: 60, EstimatedVRAMGB: 6.0,
	}

	specs := &hardware.HardwareSpecs{
		HasGPU:          true,
		AvailableVRAMGB: 12.0,
		GPUs: []hardware.GPUInfo{
			{Vendor: hardware.GPUVendorNVIDIA, VRAMTotalGB: 12.0},
		},
	}

	est := EstimatePerformance(model, specs)

	if est.TokensPerSecond <= 0 {
		t.Errorf("TokensPerSecond should be positive, got %.1f", est.TokensPerSecond)
	}
	if est.QualityRating == "" {
		t.Error("QualityRating should not be empty")
	}
	if est.TimeToFirstToken == "" {
		t.Error("TimeToFirstToken should not be empty")
	}
}
