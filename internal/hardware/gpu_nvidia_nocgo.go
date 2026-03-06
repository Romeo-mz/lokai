//go:build !cgo

package hardware

// enrichNVIDIA is a no-op when CGO is disabled (cross-compilation).
func enrichNVIDIA(_ *HardwareSpecs) {}
