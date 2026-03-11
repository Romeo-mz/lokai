package hardware

import "testing"

func TestComputeAvailableVRAM_GPU(t *testing.T) {
	specs := &HardwareSpecs{
		HasGPU: true,
		GPUs: []GPUInfo{
			{Vendor: GPUVendorNVIDIA, VRAMTotalGB: 12.0, VRAMFreeGB: 11.5},
		},
	}
	specs.ComputeAvailableVRAM()

	if specs.AvailableVRAMGB != 11.5 {
		t.Errorf("expected 11.5 GB, got %.1f GB", specs.AvailableVRAMGB)
	}
}

func TestComputeAvailableVRAM_MultiGPU(t *testing.T) {
	// Two same-vendor GPUs: free VRAM should be summed (tensor-parallel support).
	specs := &HardwareSpecs{
		HasGPU: true,
		GPUs: []GPUInfo{
			{Vendor: GPUVendorNVIDIA, VRAMTotalGB: 8.0, VRAMFreeGB: 7.0},
			{Vendor: GPUVendorNVIDIA, VRAMTotalGB: 24.0, VRAMFreeGB: 22.0},
		},
	}
	specs.ComputeAvailableVRAM()

	if specs.AvailableVRAMGB != 29.0 {
		t.Errorf("expected summed VRAM across same-vendor GPUs (29.0 GB), got %.1f GB", specs.AvailableVRAMGB)
	}
}

func TestComputeAvailableVRAM_AppleSilicon(t *testing.T) {
	specs := &HardwareSpecs{
		IsAppleSilicon:  true,
		UnifiedMemoryGB: 16.0,
	}
	specs.ComputeAvailableVRAM()

	expected := 12.0 // 16 - 4 overhead
	if specs.AvailableVRAMGB != expected {
		t.Errorf("expected %.1f GB, got %.1f GB", expected, specs.AvailableVRAMGB)
	}
}

func TestComputeAvailableVRAM_CPUOnly(t *testing.T) {
	specs := &HardwareSpecs{
		RAMTotalGB:     16.0,
		RAMAvailableGB: 10.0,
	}
	specs.ComputeAvailableVRAM()

	expected := 7.0 // 10.0 * 0.7
	if specs.AvailableVRAMGB != expected {
		t.Errorf("expected %.1f GB, got %.1f GB", expected, specs.AvailableVRAMGB)
	}
}

func TestTier(t *testing.T) {
	tests := []struct {
		vram     float64
		expected string
	}{
		{0.5, "minimal"},
		{2.0, "minimal"},
		{4.0, "low"},
		{6.0, "mid"},
		{10.0, "high"},
		{16.0, "pro"},
		{48.0, "workstation"},
	}

	for _, tt := range tests {
		specs := &HardwareSpecs{AvailableVRAMGB: tt.vram}
		tier := specs.Tier()
		if tier != tt.expected {
			t.Errorf("Tier(%.1f GB) = %q, want %q", tt.vram, tier, tt.expected)
		}
	}
}
