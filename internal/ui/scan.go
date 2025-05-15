package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/romeo-mz/lokai/internal/hardware"
)

// ScanAndDisplay runs hardware detection with a spinner, then displays results.
func ScanAndDisplay(ctx context.Context) (*hardware.HardwareSpecs, error) {
	fmt.Println(SubtitleStyle.Render("⚙ Scanning hardware..."))
	fmt.Println()

	specs, err := hardware.Detect(ctx)
	if err != nil {
		return nil, fmt.Errorf("hardware detection failed: %w", err)
	}

	displaySpecs(specs)
	return specs, nil
}

// displaySpecs renders detected hardware in a formatted table.
func displaySpecs(specs *hardware.HardwareSpecs) {
	// System info.
	rows := [][]string{
		{"Platform", fmt.Sprintf("%s/%s (%s)", specs.OS, specs.Arch, specs.Platform)},
		{"CPU", specs.CPUModel},
		{"Cores / Threads", fmt.Sprintf("%d / %d", specs.CPUCores, specs.CPUThreads)},
		{"CPU Frequency", fmt.Sprintf("%.0f MHz", specs.CPUFreqMHz)},
		{"RAM Total", fmt.Sprintf("%.1f GB", specs.RAMTotalGB)},
		{"RAM Available", fmt.Sprintf("%.1f GB", specs.RAMAvailableGB)},
	}

	// CPU features.
	features := ""
	if specs.HasAVX2 {
		features += "AVX2 "
	}
	if specs.HasAVX512 {
		features += "AVX-512 "
	}
	if features == "" {
		features = "none detected"
	}
	rows = append(rows, []string{"CPU Features", features})

	// GPU info.
	if len(specs.GPUs) > 0 {
		for i, gpu := range specs.GPUs {
			label := "GPU"
			if len(specs.GPUs) > 1 {
				label = fmt.Sprintf("GPU %d", i+1)
			}
			gpuDesc := fmt.Sprintf("%s %s", gpu.Vendor, gpu.Name)
			if gpu.VRAMTotalGB > 0 {
				gpuDesc += fmt.Sprintf(" (%.1f GB VRAM, %.1f GB free)", gpu.VRAMTotalGB, gpu.VRAMFreeGB)
			}
			if gpu.CUDACapability != "" {
				gpuDesc += fmt.Sprintf(" [CUDA %s]", gpu.CUDACapability)
			}
			rows = append(rows, []string{label, gpuDesc})
		}
	} else if specs.IsAppleSilicon {
		rows = append(rows, []string{"GPU", fmt.Sprintf("Apple Silicon (%.0f GB unified memory)", specs.UnifiedMemoryGB)})
	} else {
		rows = append(rows, []string{"GPU", WarningStyle.Render("No discrete GPU detected")})
	}

	// VRAM budget.
	budgetLabel := "VRAM Budget"
	if !specs.HasGPU {
		budgetLabel = "RAM Budget"
	}
	rows = append(rows, []string{budgetLabel, fmt.Sprintf("%.1f GB available for models", specs.AvailableVRAMGB)})

	// Hardware tier.
	rows = append(rows, []string{"Hardware Tier", fmt.Sprintf("%s", specs.Tier())})

	// Virtualization.
	if specs.IsVirtualized {
		rows = append(rows, []string{"Virtualization", WarningStyle.Render(specs.VirtualizationSystem)})
	}

	// Render table.
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorSecondary)).
		Headers("Component", "Details").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().
					Bold(true).
					Foreground(ColorSecondary).
					Padding(0, 1)
			}
			if col == 0 {
				return LabelStyle.Padding(0, 1)
			}
			return ValueStyle.Padding(0, 1)
		})

	fmt.Println(t)
	fmt.Println()
}
