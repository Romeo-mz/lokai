package models

import (
	"strings"
	"testing"

	"github.com/romeo-mz/lokai/internal/hardware"
)

// ─── EstimatePerformance ─────────────────────────────────────────────────────

func model7B() ModelEntry {
	return ModelEntry{
		Name: "Test 7B", OllamaTag: "test:7b", Family: "llama",
		ParameterSize: "7B", ParameterCount: 7.0,
		Quality: 60, EstimatedVRAMGB: 5.5,
	}
}

func specsNVIDIA12GB() *hardware.HardwareSpecs {
	return &hardware.HardwareSpecs{
		HasGPU: true, AvailableVRAMGB: 12.0,
		GPUs: []hardware.GPUInfo{
			{Vendor: hardware.GPUVendorNVIDIA, VRAMTotalGB: 12.0, VRAMFreeGB: 12.0},
		},
	}
}

func TestEstimatePerformance_NVIDIA(t *testing.T) {
	est := EstimatePerformance(model7B(), specsNVIDIA12GB())

	if est.TokensPerSecond <= 0 {
		t.Errorf("TokensPerSecond should be > 0, got %.1f", est.TokensPerSecond)
	}
	if est.QualityRating == "" {
		t.Error("QualityRating should not be empty")
	}
	if est.TimeToFirstToken == "" {
		t.Error("TimeToFirstToken should not be empty")
	}
	if est.GenerationTime == "" {
		t.Error("GenerationTime should not be empty")
	}
}

func TestEstimatePerformance_AppleSilicon(t *testing.T) {
	specs := &hardware.HardwareSpecs{
		IsAppleSilicon:  true,
		UnifiedMemoryGB: 16.0,
		AvailableVRAMGB: 12.0,
	}

	est := EstimatePerformance(model7B(), specs)
	if est.TokensPerSecond <= 0 {
		t.Errorf("Apple Silicon estimate should be > 0, got %.1f", est.TokensPerSecond)
	}
	// Apple Silicon does NOT carry a "CPU-only" note.
	if strings.Contains(est.Notes, "CPU-only") {
		t.Error("Apple Silicon should not carry CPU-only note")
	}
}

func TestEstimatePerformance_CPUOnly(t *testing.T) {
	specs := &hardware.HardwareSpecs{
		HasGPU:          false,
		CPUCores:        8,
		HasAVX2:         true,
		RAMTotalGB:      32,
		RAMAvailableGB:  20,
		AvailableVRAMGB: 14,
	}

	est := EstimatePerformance(model7B(), specs)
	if est.TokensPerSecond <= 0 {
		t.Errorf("CPU-only estimate should be > 0, got %.1f", est.TokensPerSecond)
	}
	if !strings.Contains(est.Notes, "CPU-only") {
		t.Error("CPU-only inference should carry a CPU-only note")
	}
}

func TestEstimatePerformance_HighVRAMPressure(t *testing.T) {
	// A model that uses >85% of available VRAM should trigger a note.
	bigModel := ModelEntry{
		Name: "Big Model", ParameterCount: 30, Quality: 80,
		EstimatedVRAMGB: 11.5, // 11.5 / 12.0 ≈ 95 % → note expected
	}
	est := EstimatePerformance(bigModel, specsNVIDIA12GB())
	if !strings.Contains(est.Notes, "most of your VRAM") {
		t.Error("High VRAM usage should produce a warning note")
	}
}

func TestEstimatePerformance_AVX2Note(t *testing.T) {
	specs := &hardware.HardwareSpecs{HasAVX2: true, CPUCores: 8, AvailableVRAMGB: 14}
	est := EstimatePerformance(model7B(), specs)
	if !strings.Contains(est.Notes, "AVX2") {
		t.Error("AVX2-capable machine should carry an AVX2 note")
	}
}

// ─── estimateTokensPerSecond tiers ───────────────────────────────────────────

func TestEstimateTokensPerSecond_GPUTiers(t *testing.T) {
	tests := []struct {
		name        string
		vramGB      float64
		minExpected float64
	}{
		{"4090 tier (24GB)", 24.0, 5},
		{"4080 tier (16GB)", 16.0, 5},
		{"3080 tier (12GB)", 12.0, 5},
		{"3060 tier (8GB)", 8.0, 5},
		{"1660 tier (6GB)", 6.0, 3},
		{"low-end (4GB)", 4.0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs := &hardware.HardwareSpecs{
				HasGPU:          true,
				AvailableVRAMGB: tt.vramGB,
				GPUs: []hardware.GPUInfo{
					{Vendor: hardware.GPUVendorNVIDIA, VRAMTotalGB: tt.vramGB, VRAMFreeGB: tt.vramGB},
				},
			}
			est := EstimatePerformance(model7B(), specs)
			if est.TokensPerSecond < tt.minExpected {
				t.Errorf("%s: got %.1f tok/s, want >= %.1f", tt.name, est.TokensPerSecond, tt.minExpected)
			}
		})
	}
}

func TestEstimateTokensPerSecond_AppleSiliconTiers(t *testing.T) {
	tests := []struct {
		name    string
		memGB   float64
		minToks float64
	}{
		{"Ultra (64GB)", 64, 10},
		{"Max (32GB)", 32, 8},
		{"Pro (16GB)", 16, 5},
		{"Base (8GB)", 8, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs := &hardware.HardwareSpecs{
				IsAppleSilicon:  true,
				UnifiedMemoryGB: tt.memGB,
				AvailableVRAMGB: tt.memGB - 4,
			}
			est := EstimatePerformance(model7B(), specs)
			if est.TokensPerSecond < tt.minToks {
				t.Errorf("%s: got %.1f tok/s, want >= %.1f", tt.name, est.TokensPerSecond, tt.minToks)
			}
		})
	}
}

func TestEstimateTokensPerSecond_LargeModelSlower(t *testing.T) {
	smallModel := ModelEntry{ParameterCount: 3, Quality: 50, EstimatedVRAMGB: 2}
	largeModel := ModelEntry{ParameterCount: 70, Quality: 90, EstimatedVRAMGB: 40}
	specs := specsNVIDIA12GB()

	smallEst := EstimatePerformance(smallModel, specs)
	largeEst := EstimatePerformance(largeModel, specs)

	if largeEst.TokensPerSecond >= smallEst.TokensPerSecond {
		t.Errorf("Large model (70B) should be slower than small (3B): %.1f vs %.1f",
			largeEst.TokensPerSecond, smallEst.TokensPerSecond)
	}
}

// ─── qualityRating ───────────────────────────────────────────────────────────

func TestQualityRating(t *testing.T) {
	cases := []struct {
		quality  int
		expected string
	}{
		{90, "Excellent"},
		{85, "Excellent"},
		{70, "High"},
		{65, "High"},
		{50, "Medium"},
		{40, "Medium"},
		{20, "Low"},
		{1, "Low"},
	}
	for _, c := range cases {
		got := qualityRating(c.quality)
		if got != c.expected {
			t.Errorf("qualityRating(%d) = %q, want %q", c.quality, got, c.expected)
		}
	}
}

// ─── RealWorldTasks ──────────────────────────────────────────────────────────

func TestRealWorldTasks_AllUseCases(t *testing.T) {
	useCases := []UseCase{
		UseCaseChat, UseCaseCode, UseCaseVision, UseCaseEmbedding,
		UseCaseReasoning, UseCaseImage, UseCaseVideo, UseCaseAudio, UseCaseNSFW,
	}
	for _, uc := range useCases {
		t.Run(string(uc), func(t *testing.T) {
			tasks := RealWorldTasks(uc, 30.0)
			if len(tasks) == 0 {
				t.Errorf("RealWorldTasks(%s, 30.0) returned no tasks", uc)
			}
			for _, task := range tasks {
				if task == "" {
					t.Errorf("RealWorldTasks(%s) has an empty task string", uc)
				}
			}
		})
	}
}

func TestRealWorldTasks_ZeroTokPerSec(t *testing.T) {
	tasks := RealWorldTasks(UseCaseChat, 0)
	if tasks != nil {
		t.Errorf("RealWorldTasks with 0 tok/s should return nil, got %v", tasks)
	}
}

func TestRealWorldTasks_FastHardware(t *testing.T) {
	tasks := RealWorldTasks(UseCaseChat, 100.0)
	// With 100 tok/s, quick Q&A should take well under 1 second.
	found := false
	for _, task := range tasks {
		if strings.Contains(task, "Quick Q&A") {
			found = true
			// Should mention "second" (fast).
			if !strings.Contains(task, "second") {
				t.Errorf("Expected 'second' in task time, got: %s", task)
			}
		}
	}
	if !found {
		t.Error("Expected 'Quick Q&A' task for UseCaseChat")
	}
}

// ─── estimateGenerationTime ──────────────────────────────────────────────────

func TestEstimateGenerationTime_ZeroToks(t *testing.T) {
	result := estimateGenerationTime(0)
	if result != "unknown" {
		t.Errorf("0 tok/s should return 'unknown', got %q", result)
	}
}

func TestEstimateGenerationTime_Fast(t *testing.T) {
	// 200 tokens at 100 tok/s = 2s → should contain "s".
	result := estimateGenerationTime(100)
	if !strings.Contains(result, "s") {
		t.Errorf("Fast generation should show seconds, got %q", result)
	}
}

func TestEstimateGenerationTime_Slow(t *testing.T) {
	// 200 tokens at 0.5 tok/s = 400s = 6.67 min → should contain "min".
	result := estimateGenerationTime(0.5)
	if !strings.Contains(result, "min") {
		t.Errorf("Slow generation should show minutes, got %q", result)
	}
}
