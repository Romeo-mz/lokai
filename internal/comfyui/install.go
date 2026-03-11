package comfyui

import (
	"context"
	"fmt"
	"runtime"
)

// InstallStatus describes the state of ComfyUI availability.
type InstallStatus struct {
	Running      bool
	Addr         string
	ErrorMessage string
}

// CheckInstallation tests whether ComfyUI is reachable at the configured address.
func CheckInstallation(ctx context.Context) InstallStatus {
	c := NewClient()
	status := InstallStatus{Addr: c.addr}
	if err := c.CheckHealth(ctx); err != nil {
		status.ErrorMessage = err.Error()
		return status
	}
	status.Running = true
	return status
}

// GetInstallInstructions returns platform-specific ComfyUI setup instructions.
func GetInstallInstructions() string {
	switch runtime.GOOS {
	case "darwin", "linux":
		return fmt.Sprintf(`ComfyUI is not running. To install and start it:

  git clone https://github.com/comfyanonymous/ComfyUI
  cd ComfyUI
  pip install -r requirements.txt
  python main.py

Then place your model checkpoints in:
  ComfyUI/models/checkpoints/

Once running, lokai will connect at: %s
(override with COMFYUI_HOST env variable)`, defaultAddr)

	case "windows":
		return fmt.Sprintf(`ComfyUI is not running. Download the portable package from:
  https://github.com/comfyanonymous/ComfyUI/releases

Then place your model checkpoints in:
  ComfyUI\models\checkpoints\

Once running, lokai will connect at: %s
(override with COMFYUI_HOST env variable)`, defaultAddr)

	default:
		return fmt.Sprintf(`ComfyUI is not running. Install from:
  https://github.com/comfyanonymous/ComfyUI

lokai will connect at: %s
(override with COMFYUI_HOST env variable)`, defaultAddr)
	}
}
