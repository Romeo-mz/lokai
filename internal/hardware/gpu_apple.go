//go:build darwin && arm64

package hardware

import "strings"

// detectAppleSiliconDetails enriches specs with Apple Silicon-specific info.
// On Apple Silicon, the GPU shares unified memory with the CPU.
// We parse the CPU model name to estimate GPU core count.
func detectAppleSiliconDetails(specs *HardwareSpecs) {
	if !specs.IsAppleSilicon {
		return
	}

	// Unified memory is the total system RAM.
	specs.UnifiedMemoryGB = specs.RAMTotalGB

	// Estimate GPU cores from the CPU model name.
	cores := estimateAppleGPUCores(specs.CPUModel)
	if cores > 0 {
		specs.GPUs = append(specs.GPUs, GPUInfo{
			Vendor:       GPUVendorApple,
			Name:         specs.CPUModel + " GPU",
			VRAMTotalGB:  specs.UnifiedMemoryGB,
			VRAMFreeGB:   specs.RAMAvailableGB,
			ComputeCores: cores,
		})
	}
}

// estimateAppleGPUCores returns a rough GPU core count based on the chip name.
// These are approximate and based on public Apple specifications.
func estimateAppleGPUCores(model string) int {
	// Apple Silicon GPU core counts (approximate):
	coreMap := map[string]int{
		// M1 family
		"M1":       8,
		"M1 Pro":   16,
		"M1 Max":   32,
		"M1 Ultra": 64,
		// M2 family
		"M2":       10,
		"M2 Pro":   19,
		"M2 Max":   38,
		"M2 Ultra": 76,
		// M3 family
		"M3":       10,
		"M3 Pro":   18,
		"M3 Max":   40,
		"M3 Ultra": 80,
		// M4 family
		"M4":       10,
		"M4 Pro":   20,
		"M4 Max":   40,
		"M4 Ultra": 80,
	}

	for chip, cores := range coreMap {
		if containsChipName(model, chip) {
			return cores
		}
	}

	return 8 // Fallback for unknown Apple Silicon.
}

// containsChipName checks if the model string contains the chip name.
// Handles cases like "Apple M2 Pro" containing "M2 Pro".
func containsChipName(model, chip string) bool {
	return strings.Contains(model, chip)
}
