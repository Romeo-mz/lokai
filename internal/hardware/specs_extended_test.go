package hardware

import (
	"testing"
)

// ─── ComputeAvailableVRAM ────────────────────────────────────────────────────

func TestComputeAvailableVRAM_NvidiaGPU(t *testing.T) {
	s := &HardwareSpecs{
		HasGPU: true,
		GPUs: []GPUInfo{
			{Vendor: GPUVendorNVIDIA, VRAMTotalGB: 12.0, VRAMFreeGB: 10.0},
		},
	}
	s.ComputeAvailableVRAM()
	if s.AvailableVRAMGB != 10.0 {
		t.Errorf("Expected 10.0 GB from free VRAM, got %.1f", s.AvailableVRAMGB)
	}
}

func TestComputeAvailableVRAM_MultiGPU_SumsVendor(t *testing.T) {
	// Two same-vendor GPUs: free VRAM is summed for tensor-parallel inference.
	s := &HardwareSpecs{
		HasGPU: true,
		GPUs: []GPUInfo{
			{Vendor: GPUVendorNVIDIA, VRAMTotalGB: 8.0, VRAMFreeGB: 6.0},
			{Vendor: GPUVendorNVIDIA, VRAMTotalGB: 24.0, VRAMFreeGB: 20.0},
		},
	}
	s.ComputeAvailableVRAM()
	if s.AvailableVRAMGB != 26.0 {
		t.Errorf("Expected summed free VRAM (26 GB), got %.1f", s.AvailableVRAMGB)
	}
}

func TestComputeAvailableVRAM_AppleSilicon16GB(t *testing.T) {
	s := &HardwareSpecs{
		IsAppleSilicon:  true,
		UnifiedMemoryGB: 16.0,
	}
	s.ComputeAvailableVRAM()
	expected := 16.0 - 4.0 // minus OS overhead
	if s.AvailableVRAMGB != expected {
		t.Errorf("Apple Silicon: expected %.1f GB, got %.1f", expected, s.AvailableVRAMGB)
	}
}

func TestComputeAvailableVRAM_AppleSilicon_SmallMemory(t *testing.T) {
	// 6 GB unified memory — 4 GB overhead → 2 GB, should not be negative.
	s := &HardwareSpecs{
		IsAppleSilicon:  true,
		UnifiedMemoryGB: 6.0,
	}
	s.ComputeAvailableVRAM()
	if s.AvailableVRAMGB < 0 {
		t.Errorf("AvailableVRAMGB should never be negative, got %.1f", s.AvailableVRAMGB)
	}
}

func TestComputeAvailableVRAM_AppleSilicon_TinyMemory(t *testing.T) {
	// If unified memory < 4 GB overhead, fallback to 50% of total.
	s := &HardwareSpecs{
		IsAppleSilicon:  true,
		UnifiedMemoryGB: 2.0,
	}
	s.ComputeAvailableVRAM()
	if s.AvailableVRAMGB <= 0 {
		t.Errorf("Tiny Apple Silicon memory should still have a positive budget, got %.1f", s.AvailableVRAMGB)
	}
}

func TestComputeAvailableVRAM_CPUOnly_UsesRAM(t *testing.T) {
	s := &HardwareSpecs{
		RAMAvailableGB: 16.0,
	}
	s.ComputeAvailableVRAM()
	expected := 16.0 * 0.7
	if s.AvailableVRAMGB != expected {
		t.Errorf("CPU-only: expected %.1f (70%% of available RAM), got %.1f", expected, s.AvailableVRAMGB)
	}
}

func TestComputeAvailableVRAM_CPUOnly_FallbackToTotalRAM(t *testing.T) {
	s := &HardwareSpecs{
		RAMTotalGB:     32.0,
		RAMAvailableGB: 0,
	}
	s.ComputeAvailableVRAM()
	expected := 32.0 * 0.5
	if s.AvailableVRAMGB != expected {
		t.Errorf("No available RAM: expected %.1f (50%% of total), got %.1f", expected, s.AvailableVRAMGB)
	}
}

// ─── Summary ─────────────────────────────────────────────────────────────────

func TestSummary_WithGPU(t *testing.T) {
	s := &HardwareSpecs{
		CPUModel:   "Intel Core i9",
		CPUCores:   16,
		RAMTotalGB: 64,
		GPUs: []GPUInfo{
			{Name: "RTX 4090", VRAMTotalGB: 24.0},
		},
	}
	summary := s.Summary()
	if summary == "" {
		t.Error("Summary should not be empty")
	}
	if !containsStr(summary, "RTX 4090") {
		t.Errorf("Summary should mention GPU name, got: %s", summary)
	}
}

func TestSummary_NoGPU(t *testing.T) {
	s := &HardwareSpecs{
		CPUModel:   "AMD Ryzen 9",
		CPUCores:   12,
		RAMTotalGB: 32,
	}
	summary := s.Summary()
	if !containsStr(summary, "no GPU") {
		t.Errorf("Summary should say 'no GPU', got: %s", summary)
	}
}

func TestSummary_AppleSilicon(t *testing.T) {
	s := &HardwareSpecs{
		CPUModel:        "Apple M2 Pro",
		CPUCores:        10,
		RAMTotalGB:      16,
		IsAppleSilicon:  true,
		UnifiedMemoryGB: 16.0,
	}
	summary := s.Summary()
	if !containsStr(summary, "Apple Silicon") {
		t.Errorf("Summary should mention Apple Silicon, got: %s", summary)
	}
}

func TestSummary_GPUNoVRAM(t *testing.T) {
	s := &HardwareSpecs{
		CPUModel: "Intel i7", CPUCores: 8, RAMTotalGB: 16,
		GPUs: []GPUInfo{{Name: "Unknown GPU", VRAMTotalGB: 0}},
	}
	summary := s.Summary()
	if summary == "" {
		t.Error("Summary should not be empty even with unknown GPU VRAM")
	}
}

// ─── Tier ────────────────────────────────────────────────────────────────────

func TestTierAllBuckets(t *testing.T) {
	cases := []struct {
		vram     float64
		expected string
	}{
		{80, "workstation"},
		{48, "workstation"},
		{20, "pro"},
		{16, "pro"},
		{11, "high"},
		{10, "high"},
		{8, "mid"},
		{6, "mid"},
		{5, "low"},
		{4, "low"},
		{2, "minimal"},
		{0, "minimal"},
	}
	for _, c := range cases {
		s := &HardwareSpecs{AvailableVRAMGB: c.vram}
		got := s.Tier()
		if got != c.expected {
			t.Errorf("Tier(%.0f GB) = %q, want %q", c.vram, got, c.expected)
		}
	}
}

// ─── UseCases constants ──────────────────────────────────────────────────────

func TestUseCaseConstants(t *testing.T) {
	all := []UseCase{
		UseCaseChat, UseCaseCode, UseCaseVision, UseCaseEmbedding,
		UseCaseReasoning, UseCaseImage, UseCaseVideo, UseCaseAudio, UseCaseNSFW,
	}
	seen := make(map[UseCase]bool)
	for _, uc := range all {
		if seen[uc] {
			t.Errorf("Duplicate UseCase constant: %s", uc)
		}
		seen[uc] = true
		if string(uc) == "" {
			t.Error("UseCase constant should not be empty string")
		}
	}
}

// ─── GPUVendor constants ─────────────────────────────────────────────────────

func TestGPUVendorConstants(t *testing.T) {
	vendors := []GPUVendor{GPUVendorNVIDIA, GPUVendorAMD, GPUVendorIntel, GPUVendorApple, GPUVendorUnknown}
	for _, v := range vendors {
		if string(v) == "" {
			t.Error("GPUVendor constant should not be empty")
		}
	}
}

// ─── helper ──────────────────────────────────────────────────────────────────

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
