package ollama

import (
	"os/exec"
	"runtime"
	"os"
)

// InstallStatus represents the current state of the Ollama installation.
type InstallStatus struct {
	Installed    bool   `json:"installed"`
	ServerUp     bool   `json:"server_up"`
	BinaryPath   string `json:"binary_path,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// CheckInstallation verifies if Ollama is installed and running.
func CheckInstallation() InstallStatus {
	status := InstallStatus{}

	// Check if ollama binary is in PATH.
	path, err := exec.LookPath("ollama")
	if err != nil {
		// Check if running in Docker container via environment variable
		if os.Getenv("OLLAMA_HOST") != "" {
			status.Installed = true
			status.BinaryPath = "ollama (container)"
			return status
		}
	}
	if err != nil {
		status.ErrorMessage = "Ollama is not installed"
		return status
	}

	status.Installed = true
	status.BinaryPath = path

	return status
}

// GetInstallInstructions returns platform-specific install instructions.
func GetInstallInstructions() string {
	switch runtime.GOOS {
	case "darwin":
		return `Ollama is not installed. Install it with:

  brew install ollama

Or download from: https://ollama.com/download/mac

Then start the server:

  ollama serve`

	case "linux":
		return `Ollama is not installed. Install it with:

  curl -fsSL https://ollama.com/install.sh | sh

Then start the server:

  ollama serve

Or run as a systemd service:

  sudo systemctl start ollama`

	case "windows":
		return 	`Ollama is not installed. Install it with:

  winget install Ollama.Ollama

Or download from: https://ollama.com/download/windows

The server starts automatically after installation.`

	default:
		return `Ollama is not installed. Download from: https://ollama.com/download`
	}
}
