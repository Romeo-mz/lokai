//go:build linux

package hardware

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// enrichAMD reads VRAM info from sysfs for AMD GPUs on Linux.
func enrichAMD(specs *HardwareSpecs) {
	hasAMD := false
	for _, gpu := range specs.GPUs {
		if gpu.Vendor == GPUVendorAMD {
			hasAMD = true
			break
		}
	}
	if !hasAMD {
		return
	}

	const bytesPerGB = 1024 * 1024 * 1024

	// Scan /sys/class/drm/card*/device/mem_info_vram_total
	cards, err := filepath.Glob("/sys/class/drm/card[0-9]*/device/mem_info_vram_total")
	if err != nil || len(cards) == 0 {
		return
	}

	for _, cardPath := range cards {
		totalBytes, err := readSysfsBytes(cardPath)
		if err != nil {
			continue
		}

		usedPath := strings.Replace(cardPath, "mem_info_vram_total", "mem_info_vram_used", 1)
		usedBytes, _ := readSysfsBytes(usedPath)

		freeBytes := totalBytes - usedBytes
		if freeBytes < 0 {
			freeBytes = 0
		}

		// Match to the next AMD GPU entry that has no VRAM data yet.
		// sysfs card order is stable within a boot, matching by sequential index.
		for j := range specs.GPUs {
			if specs.GPUs[j].Vendor == GPUVendorAMD && specs.GPUs[j].VRAMTotalGB == 0 {
				specs.GPUs[j].VRAMTotalGB = float64(totalBytes) / bytesPerGB
				specs.GPUs[j].VRAMFreeGB = float64(freeBytes) / bytesPerGB
				break
			}
		}
	}
}

// readSysfsBytes reads a sysfs file containing a decimal byte count.
func readSysfsBytes(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}
