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
	if est.Measured {
		t.Error("Measured should be false without baselines")
	}
}

func TestEstimatePerformanceWithBaselines(t *testing.T) {
	model := ModelEntry{
		Name: "Test Model", OllamaTag: "test:7b",
		ParameterCount: 7.0, Quality: 60, EstimatedVRAMGB: 6.0,
	}

	specs := &hardware.HardwareSpecs{
		HasGPU:          true,
		AvailableVRAMGB: 12.0,
		GPUs: []hardware.GPUInfo{
			{Vendor: hardware.GPUVendorNVIDIA, VRAMTotalGB: 12.0},
		},
	}

	// Set baselines.
	SetBaselines(map[string]Baseline{
		"test:7b": {TokensPerSecond: 42.5, TimeToFirstToken: 250000000}, // 250ms
	})
	defer SetBaselines(nil)

	est := EstimatePerformance(model, specs)

	if !est.Measured {
		t.Error("Measured should be true when baseline exists")
	}
	if est.TokensPerSecond != 42.5 {
		t.Errorf("TokensPerSecond = %.1f, want 42.5", est.TokensPerSecond)
	}
}

func TestGetBaselineNotFound(t *testing.T) {
	SetBaselines(map[string]Baseline{})
	defer SetBaselines(nil)

	_, ok := getBaseline("nonexistent:model")
	if ok {
		t.Error("getBaseline should return false for missing tag")
	}
}

func TestEstimatePerformanceCPUOnly(t *testing.T) {
	model := ModelEntry{
		Name: "Test Model", ParameterCount: 3.0,
		Quality: 40, EstimatedVRAMGB: 2.0,
	}

	specs := &hardware.HardwareSpecs{
		AvailableVRAMGB: 0,
		RAMTotalGB:      16.0,
		RAMAvailableGB:  12.0,
		CPUCores:        8,
		HasAVX2:         true,
	}

	est := EstimatePerformance(model, specs)

	if est.TokensPerSecond <= 0 {
		t.Errorf("TokensPerSecond should be positive for CPU, got %.1f", est.TokensPerSecond)
	}
	if est.TimeToFirstToken == "" {
		t.Error("TimeToFirstToken should not be empty")
	}
}

func TestInferParamCount(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"8b", 8},
		{"3.8b", 3.8},
		{"70b-instruct", 70},
		{"135m", 0.135},
		{"latest", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := inferParamCount(tt.input)
			if got != tt.want {
				t.Errorf("inferParamCount(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestInferUseCases(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want UseCase
	}{
		{"codellama", "code model", UseCaseCode},
		{"nomic-embed", "embedding model", UseCaseEmbedding},
		{"llava", "vision model", UseCaseVision},
		{"whisper", "audio model", UseCaseAudio},
		{"flux", "image gen", UseCaseImage},
		{"wan", "video gen", UseCaseVideo},
		{"dolphin", "uncensored", UseCaseNSFW},
		{"deepseek-r1", "reasoning model", UseCaseReasoning},
		{"llama3", "general chat", UseCaseChat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferUseCases(tt.name, tt.desc)
			if len(got) == 0 {
				t.Fatal("inferUseCases returned empty")
			}
			if got[0] != tt.want {
				t.Errorf("inferUseCases(%q, %q) = %v, want %v", tt.name, tt.desc, got[0], tt.want)
			}
		})
	}
}

func TestRealWorldTasks(t *testing.T) {
	tasks := RealWorldTasks(UseCaseChat, 30.0)
	if len(tasks) == 0 {
		t.Error("RealWorldTasks should return tasks for Chat")
	}
	for _, task := range tasks {
		if task == "" {
			t.Error("task string should not be empty")
		}
	}

	// Zero tok/sec should return nil.
	if tasks := RealWorldTasks(UseCaseChat, 0); tasks != nil {
		t.Errorf("Expected nil for 0 tok/sec, got %v", tasks)
	}
}

func TestQualityRating(t *testing.T) {
	tests := []struct {
		quality int
		want    string
	}{
		{90, "Excellent"},
		{70, "High"},
		{50, "Medium"},
		{20, "Low"},
	}

	for _, tt := range tests {
		got := qualityRating(tt.quality)
		if got != tt.want {
			t.Errorf("qualityRating(%d) = %q, want %q", tt.quality, got, tt.want)
		}
	}
}
