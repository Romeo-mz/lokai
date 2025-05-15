package hardware

import (
	"context"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	hostinfo "github.com/shirou/gopsutil/v4/host"
)

// detectPlatform populates platform/host info in specs.
func detectPlatform(ctx context.Context, specs *HardwareSpecs) error {
	info, err := hostinfo.InfoWithContext(ctx)
	if err != nil {
		return err
	}

	specs.Hostname = info.Hostname
	specs.Platform = info.Platform
	specs.IsVirtualized = info.VirtualizationRole == "guest"
	specs.VirtualizationSystem = info.VirtualizationSystem

	return nil
}

// detectCPU populates CPU info in specs using gopsutil.
func detectCPU(ctx context.Context, specs *HardwareSpecs) error {
	infos, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return err
	}

	if len(infos) > 0 {
		first := infos[0]
		specs.CPUModel = first.ModelName
		specs.CPUVendor = first.VendorID
		specs.CPUFreqMHz = first.Mhz
		specs.CPUFlags = first.Flags

		// Check for AVX2 and AVX-512 support.
		for _, flag := range first.Flags {
			switch strings.ToLower(flag) {
			case "avx2":
				specs.HasAVX2 = true
			case "avx512f", "avx512_vnni":
				specs.HasAVX512 = true
			}
		}
	}

	// Physical cores.
	physical, err := cpu.CountsWithContext(ctx, false)
	if err == nil {
		specs.CPUCores = physical
	}

	// Logical threads.
	logical, err := cpu.CountsWithContext(ctx, true)
	if err == nil {
		specs.CPUThreads = logical
	}

	return nil
}
