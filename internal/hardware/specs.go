package hardware

import "fmt"

// UseCase represents the type of AI task the user wants to run.
type UseCase string

const (
	UseCaseChat      UseCase = "chat"
	UseCaseCode      UseCase = "code"
	UseCaseVision    UseCase = "vision"
	UseCaseEmbedding UseCase = "embedding"
	UseCaseReasoning UseCase = "reasoning"
	UseCaseImage     UseCase = "image"
	UseCaseVideo     UseCase = "video"
	UseCaseAudio     UseCase = "audio"
	UseCaseUnrestricted      UseCase = "unrestricted"
)

// GPUVendor identifies the GPU manufacturer.
type GPUVendor string

const (
	GPUVendorNVIDIA  GPUVendor = "NVIDIA"
	GPUVendorAMD     GPUVendor = "AMD"
	GPUVendorIntel   GPUVendor = "Intel"
	GPUVendorApple   GPUVendor = "Apple"
	GPUVendorUnknown GPUVendor = "Unknown"
)

// GPUInfo holds detected information about a single GPU.
type GPUInfo struct {
	Vendor         GPUVendor `json:"vendor"`
	Name           string    `json:"name"`
	VRAMTotalGB    float64   `json:"vram_total_gb"`
	VRAMFreeGB     float64   `json:"vram_free_gb"`
	CUDACapability string    `json:"cuda_capability,omitempty"`
	PCIAddress     string    `json:"pci_address,omitempty"`
	DriverVersion  string    `json:"driver_version,omitempty"`
	ComputeCores   int       `json:"compute_cores,omitempty"`
}

// HardwareSpecs holds all detected system hardware information.
type HardwareSpecs struct {
	// Platform
	OS       string `json:"os"`       // "linux", "darwin", "windows"
	Arch     string `json:"arch"`     // "amd64", "arm64"
	Hostname string `json:"hostname"` // machine hostname
	Platform string `json:"platform"` // "ubuntu", "arch", "darwin", etc.

	// CPU
	CPUModel   string   `json:"cpu_model"`    // e.g. "Intel Core i9-13900K", "Apple M2 Pro"
	CPUVendor  string   `json:"cpu_vendor"`   // "GenuineIntel", "AuthenticAMD", "Apple"
	CPUCores   int      `json:"cpu_cores"`    // physical cores
	CPUThreads int      `json:"cpu_threads"`  // logical threads
	CPUFreqMHz float64  `json:"cpu_freq_mhz"` // base clock speed
	CPUFlags   []string `json:"cpu_flags"`    // instruction set flags
	HasAVX2    bool     `json:"has_avx2"`     // AVX2 support (important for CPU inference)
	HasAVX512  bool     `json:"has_avx512"`   // AVX-512 support

	// Memory
	RAMTotalGB     float64 `json:"ram_total_gb"`
	RAMAvailableGB float64 `json:"ram_available_gb"`

	// GPU
	GPUs     []GPUInfo `json:"gpus"`
	HasGPU   bool      `json:"has_gpu"`
	GPUCount int       `json:"gpu_count"`

	// Apple Silicon
	IsAppleSilicon  bool    `json:"is_apple_silicon"`
	UnifiedMemoryGB float64 `json:"unified_memory_gb,omitempty"`

	// Computed budget — how much VRAM/RAM is available for model loading
	AvailableVRAMGB float64 `json:"available_vram_gb"`

	// Virtualization
	IsVirtualized        bool   `json:"is_virtualized"`
	VirtualizationSystem string `json:"virtualization_system,omitempty"`
}

// ComputeAvailableVRAM calculates the effective VRAM budget for model loading.
// Priority: discrete GPU VRAM > Apple unified memory > system RAM fallback.
// When multiple GPUs share the same vendor, their free VRAM is summed to support
// tensor-parallel inference (e.g. two RTX 3090s, NVLink setups).
func (s *HardwareSpecs) ComputeAvailableVRAM() {
	// Sum free VRAM per vendor — same-vendor multi-GPU can run models in parallel.
	vendorFree := make(map[GPUVendor]float64)
	for _, gpu := range s.GPUs {
		if gpu.VRAMFreeGB > 0 {
			vendorFree[gpu.Vendor] += gpu.VRAMFreeGB
		}
	}

	var maxVRAM float64
	for _, vram := range vendorFree {
		if vram > maxVRAM {
			maxVRAM = vram
		}
	}

	if maxVRAM > 0 {
		s.AvailableVRAMGB = maxVRAM
		return
	}

	// Apple Silicon: unified memory minus OS overhead (~4 GB).
	if s.IsAppleSilicon && s.UnifiedMemoryGB > 0 {
		overhead := 4.0
		s.AvailableVRAMGB = s.UnifiedMemoryGB - overhead
		if s.AvailableVRAMGB < 0 {
			s.AvailableVRAMGB = s.UnifiedMemoryGB * 0.5
		}
		return
	}

	// CPU-only fallback: use 70% of available RAM.
	if s.RAMAvailableGB > 0 {
		s.AvailableVRAMGB = s.RAMAvailableGB * 0.7
	} else {
		s.AvailableVRAMGB = s.RAMTotalGB * 0.5
	}
}

// Summary returns a human-readable one-line summary of the hardware.
func (s *HardwareSpecs) Summary() string {
	gpuDesc := "no GPU"
	if len(s.GPUs) > 0 {
		g := s.GPUs[0]
		if g.VRAMTotalGB > 0 {
			gpuDesc = fmt.Sprintf("%s (%.1f GB VRAM)", g.Name, g.VRAMTotalGB)
		} else {
			gpuDesc = g.Name
		}
	} else if s.IsAppleSilicon {
		gpuDesc = fmt.Sprintf("Apple Silicon (%.0f GB unified)", s.UnifiedMemoryGB)
	}

	return fmt.Sprintf("%s | %d cores | %.0f GB RAM | %s",
		s.CPUModel, s.CPUCores, s.RAMTotalGB, gpuDesc)
}

// Tier returns a hardware tier string for recommendation matching.
func (s *HardwareSpecs) Tier() string {
	vram := s.AvailableVRAMGB
	switch {
	case vram >= 48:
		return "workstation"
	case vram >= 16:
		return "pro"
	case vram >= 10:
		return "high"
	case vram >= 6:
		return "mid"
	case vram >= 4:
		return "low"
	default:
		return "minimal"
	}
}
