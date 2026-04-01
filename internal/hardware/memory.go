package hardware

import (
	"context"

	"github.com/shirou/gopsutil/v4/mem"
)

// detectMemory populates RAM info in specs using gopsutil.
func detectMemory(ctx context.Context, specs *HardwareSpecs) error {
	var vm *mem.VirtualMemoryStat
	var err error
	vm, err = mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}

	const bytesPerGB = 1024 * 1024 * 1024

	specs.RAMTotalGB = float64(vm.Total) / bytesPerGB
	specs.RAMAvailableGB = float64(vm.Available) / bytesPerGB

	return nil
}
