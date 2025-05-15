package hardware

import (
	"context"
	"strings"

	"github.com/jaypipes/ghw"
)

// detectGPUs uses ghw to discover GPUs via PCI scan, then enriches with
// vendor-specific VRAM detection (NVIDIA via NVML, AMD via sysfs, etc.).
func detectGPUs(ctx context.Context, specs *HardwareSpecs) error {
	gpuInfo, err := ghw.GPU()
	if err != nil {
		// Not fatal — machine may simply have no discrete GPU.
		return nil
	}

	for _, card := range gpuInfo.GraphicsCards {
		if card.DeviceInfo == nil {
			continue
		}

		gpu := GPUInfo{
			PCIAddress: card.Address,
		}

		dev := card.DeviceInfo

		// Identify vendor.
		if dev.Vendor != nil {
			vendorName := strings.ToUpper(dev.Vendor.Name)
			switch {
			case strings.Contains(vendorName, "NVIDIA"):
				gpu.Vendor = GPUVendorNVIDIA
			case strings.Contains(vendorName, "AMD"),
				strings.Contains(vendorName, "ADVANCED MICRO"):
				gpu.Vendor = GPUVendorAMD
			case strings.Contains(vendorName, "INTEL"):
				gpu.Vendor = GPUVendorIntel
			default:
				gpu.Vendor = GPUVendorUnknown
			}
		}

		// Product name.
		if dev.Product != nil {
			gpu.Name = dev.Product.Name
		}

		// Driver info.
		gpu.DriverVersion = dev.Driver

		specs.GPUs = append(specs.GPUs, gpu)
	}

	// Enrich with vendor-specific VRAM data.
	enrichGPUInfo(specs)

	return nil
}

// enrichGPUInfo adds VRAM data from vendor-specific sources.
// Platform-specific implementations are in gpu_nvidia.go, gpu_amd_linux.go, etc.
func enrichGPUInfo(specs *HardwareSpecs) {
	enrichNVIDIA(specs)
	enrichAMD(specs)
}
