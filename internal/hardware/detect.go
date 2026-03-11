package hardware

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
)

// Detect scans system hardware and returns a populated HardwareSpecs.
// All sub-detectors run concurrently for speed.
func Detect(ctx context.Context) (*HardwareSpecs, error) {
	specs := &HardwareSpecs{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	addErr := func(err error) {
		mu.Lock()
		errs = append(errs, err)
		mu.Unlock()
	}

	// Detect platform info.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := detectPlatform(ctx, specs); err != nil {
			addErr(fmt.Errorf("platform: %w", err))
		}
	}()

	// Detect CPU.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := detectCPU(ctx, specs); err != nil {
			addErr(fmt.Errorf("cpu: %w", err))
		}
	}()

	// Detect memory.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := detectMemory(ctx, specs); err != nil {
			addErr(fmt.Errorf("memory: %w", err))
		}
	}()

	// Detect GPUs.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := detectGPUs(ctx, specs); err != nil {
			addErr(fmt.Errorf("gpu: %w", err))
		}
	}()

	wg.Wait()

	// Post-processing.
	detectAppleSilicon(specs)
	specs.ComputeAvailableVRAM()

	specs.HasGPU = len(specs.GPUs) > 0 || specs.IsAppleSilicon
	specs.GPUCount = len(specs.GPUs)

	// Return any non-critical errors as warnings (detection still succeeds).
	if len(errs) > 0 {
		// Log warnings but don't fail — partial detection is still useful.
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  ⚠ %s\n", e)
		}
	}

	return specs, nil
}

// detectAppleSilicon checks if the system is an Apple Silicon Mac.
func detectAppleSilicon(specs *HardwareSpecs) {
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		specs.IsAppleSilicon = true
		specs.UnifiedMemoryGB = specs.RAMTotalGB
	}
}
