//go:build !cgo

package hardware

// enrichNVIDIA is a no-op when CGO is disabled (e.g. Docker/alpine builds).
// NVML requires CGO; without it we simply skip NVIDIA VRAM enrichment.
func enrichNVIDIA(_ *HardwareSpecs) {}
