package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/romeo-mz/lokai/internal/ollama"
)

// PullWithProgress downloads a model and shows a progress indicator.
func PullWithProgress(ctx context.Context, client *ollama.Client, modelTag string) error {
	fmt.Println()
	fmt.Println(SubtitleStyle.Render("📥 Downloading " + modelTag + "..."))
	fmt.Println()

	startTime := time.Now()
	var lastStatus string

	err := client.PullModel(ctx, modelTag, func(p ollama.PullProgress) {
		if p.Status != lastStatus {
			if lastStatus != "" {
				fmt.Println() // Newline before new status.
			}
			lastStatus = p.Status
		}

		if p.Total > 0 {
			// Show progress bar.
			pct := p.Percent
			barWidth := 40
			filled := int(pct / 100 * float64(barWidth))
			if filled > barWidth {
				filled = barWidth
			}
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

			elapsed := time.Since(startTime)
			speed := ""
			if elapsed.Seconds() > 1 && p.Completed > 0 {
				mbPerSec := float64(p.Completed) / 1024 / 1024 / elapsed.Seconds()
				speed = fmt.Sprintf(" (%.1f MB/s)", mbPerSec)
			}

			fmt.Printf("\r   %s [%s] %.1f%%%s", p.Status, bar, pct, speed)
		} else {
			fmt.Printf("\r   %s...", p.Status)
		}
	})

	fmt.Println() // Final newline.
	fmt.Println()

	if err != nil {
		fmt.Println(ErrorStyle.Render("✗ Download failed: " + err.Error()))
		return err
	}

	elapsed := time.Since(startTime)
	fmt.Println(SuccessStyle.Render(fmt.Sprintf("✓ Successfully downloaded %s in %s", modelTag, elapsed.Round(time.Second))))
	fmt.Println()

	// Show how to use the model.
	fmt.Println(SubtitleStyle.Render("🚀 Ready to use!"))
	fmt.Println()
	fmt.Printf("   Run your model:     %s\n", ValueStyle.Render("ollama run "+modelTag))
	fmt.Printf("   API endpoint:       %s\n", ValueStyle.Render("http://localhost:11434/api/chat"))
	fmt.Printf("   Stop the model:     %s\n", ValueStyle.Render("ollama stop "+modelTag))
	fmt.Println()

	// Offer to launch an interactive chat session right away.
	var launch bool
	_ = huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Launch "+modelTag+" now?").
			Description("Opens an interactive chat session in this terminal").
			Affirmative("Yes, launch").
			Negative("No thanks").
			Value(&launch),
	)).Run()

	if launch {
		fmt.Println()
		cmd := exec.Command("ollama", "run", modelTag)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}

	return nil
}

// ShowInstallInstructions prints the manual install commands for a model.
func ShowInstallInstructions(modelTag string) {
	fmt.Println()
	fmt.Println(SubtitleStyle.Render("📋 Install Instructions"))
	fmt.Println()
	fmt.Printf("   1. Pull the model:   %s\n", ValueStyle.Render("ollama pull "+modelTag))
	fmt.Printf("   2. Run the model:    %s\n", ValueStyle.Render("ollama run "+modelTag))
	fmt.Printf("   3. API endpoint:     %s\n", ValueStyle.Render("http://localhost:11434/api/chat"))
	fmt.Println()
	fmt.Println(MutedStyle.Render("   The model will be downloaded on first run if not already pulled."))
	fmt.Println()
}
