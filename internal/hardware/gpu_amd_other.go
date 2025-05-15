//go:build !linux

package hardware

// enrichAMD is a no-op on non-Linux platforms.
// AMD VRAM detection via sysfs is only available on Linux.
func enrichAMD(specs *HardwareSpecs) {
	// No-op: AMD VRAM detection requires Linux sysfs.
}
