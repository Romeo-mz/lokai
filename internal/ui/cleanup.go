package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/romeo-mz/lokai/internal/ollama"
)

// RunCleanup lists all installed models, asks for confirmation, then deletes them all.
func RunCleanup(ctx context.Context, client *ollama.Client) error {
	fmt.Println(SubtitleStyle.Render("🧹 Clean Up — Remove All Installed Models"))
	fmt.Println()

	// List current models.
	models, err := client.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		fmt.Println(SuccessStyle.Render("✓ No models installed — nothing to clean."))
		fmt.Println()
		return nil
	}

	// Show what will be deleted.
	fmt.Printf("   Found %s installed model(s):\n\n", ValueStyle.Render(fmt.Sprintf("%d", len(models))))
	var totalSizeGB float64
	for _, m := range models {
		sizeGB := float64(m.SizeOnDisk) / (1024 * 1024 * 1024)
		totalSizeGB += sizeGB
		fmt.Printf("   • %s  %s  %s\n",
			ValueStyle.Render(m.Name),
			MutedStyle.Render(m.ParameterSize),
			MutedStyle.Render(fmt.Sprintf("(%.1f GB)", sizeGB)),
		)
	}
	fmt.Println()
	fmt.Printf("   Total disk usage: %s\n", WarningStyle.Render(fmt.Sprintf("%.1f GB", totalSizeGB)))
	fmt.Println()

	// Confirmation.
	var confirmed bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Delete ALL installed models?").
				Description("This action cannot be undone. You will need to re-download any models you want to use.").
				Affirmative("Yes, delete everything").
				Negative("No, cancel").
				Value(&confirmed),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	if !confirmed {
		fmt.Println(MutedStyle.Render("Cleanup cancelled."))
		return nil
	}

	fmt.Println()

	// Delete all models with progress.
	deleted, err := client.DeleteAllModels(ctx, func(name string, current, total int) {
		fmt.Printf("   [%d/%d] Deleting %s...\n", current, total, ValueStyle.Render(name))
	})

	if err != nil {
		fmt.Println()
		fmt.Println(ErrorStyle.Render(fmt.Sprintf("✗ Error after deleting %d model(s): %s", len(deleted), err.Error())))
		return err
	}

	fmt.Println()
	fmt.Println(SuccessStyle.Render(fmt.Sprintf("✓ Successfully deleted %d model(s) — freed ~%.1f GB of disk space.", len(deleted), totalSizeGB)))
	fmt.Println()

	return nil
}
