//go:build !linux || !cgo
// +build !linux !cgo

package hardware

// enrichNVIDIA is a no-op when NVML/cgo is not available (Windows, static builds, etc.).
func enrichNVIDIA(specs *HardwareSpecs) {
    // intentionally no-op
}
