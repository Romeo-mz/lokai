package hardware

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// enrichNVIDIA uses NVML to get accurate VRAM info for NVIDIA GPUs.
func enrichNVIDIA(specs *HardwareSpecs) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		// NVIDIA driver not available — not an error, just skip.
		return
	}

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return
	}

	const bytesPerGB = 1024 * 1024 * 1024

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}

		name, ret := device.GetName()
		if ret != nvml.SUCCESS {
			name = "Unknown NVIDIA GPU"
		}

		memInfo, ret := device.GetMemoryInfo()
		if ret != nvml.SUCCESS {
			continue
		}

		major, minor, ret := device.GetCudaComputeCapability()
		cudaCap := ""
		if ret == nvml.SUCCESS {
			cudaCap = fmt.Sprintf("%d.%d", major, minor)
		}

		cores, _ := device.GetNumGpuCores()

		// Match to existing GPU entries from PCI scan, or append new.
		matched := false
		for j := range specs.GPUs {
			if specs.GPUs[j].Vendor == GPUVendorNVIDIA && specs.GPUs[j].VRAMTotalGB == 0 {
				specs.GPUs[j].Name = name
				specs.GPUs[j].VRAMTotalGB = float64(memInfo.Total) / bytesPerGB
				specs.GPUs[j].VRAMFreeGB = float64(memInfo.Free) / bytesPerGB
				specs.GPUs[j].CUDACapability = cudaCap
				specs.GPUs[j].ComputeCores = int(cores)
				matched = true
				break
			}
		}

		if !matched {
			specs.GPUs = append(specs.GPUs, GPUInfo{
				Vendor:         GPUVendorNVIDIA,
				Name:           name,
				VRAMTotalGB:    float64(memInfo.Total) / bytesPerGB,
				VRAMFreeGB:     float64(memInfo.Free) / bytesPerGB,
				CUDACapability: cudaCap,
				ComputeCores:   int(cores),
			})
		}
	}
}
