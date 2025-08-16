package aerospace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
)

// ExtendedAeroSpaceClient wraps the base AeroSpace client with additional functionality
type ExtendedAeroSpaceClient struct {
	aerospacecli.AeroSpaceClient
}

// NewExtendedAeroSpaceClient creates a new extended client
func NewExtendedAeroSpaceClient(baseClient aerospacecli.AeroSpaceClient) *ExtendedAeroSpaceClient {
	return &ExtendedAeroSpaceClient{
		AeroSpaceClient: baseClient,
	}
}

// SetFullscreen sets fullscreen mode for a window
func (c *ExtendedAeroSpaceClient) SetFullscreen(windowID int, enabled bool) error {
	cmd := []string{"aerospace", "fullscreen"}
	if enabled {
		cmd = append(cmd, "on")
	} else {
		cmd = append(cmd, "off")
	}
	cmd = append(cmd, "--window-id", strconv.Itoa(windowID))
	
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set fullscreen: %w, output: %s", err, output)
	}
	return nil
}

// GetScreenDimensions gets the primary screen dimensions
func (c *ExtendedAeroSpaceClient) GetScreenDimensions() (int, int, error) {
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get screen info: %w", err)
	}
	
	// Parse resolution from output
	re := regexp.MustCompile(`Resolution: (\d+) x (\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) != 3 {
		// Fallback to common resolution
		return 1920, 1080, nil
	}
	
	width, err := strconv.Atoi(matches[1])
	if err != nil {
		return 1920, 1080, nil
	}
	
	height, err := strconv.Atoi(matches[2])
	if err != nil {
		return 1920, 1080, nil
	}
	
	return width, height, nil
}

// ResizeToPercentage resizes a window to specific percentage of screen using Swift window manager
func (c *ExtendedAeroSpaceClient) ResizeToPercentage(windowID int, widthPercent, heightPercent int) error {
	// Focus the window first to ensure it's active
	if err := c.SetFocusByWindowID(windowID); err != nil {
		fmt.Printf("Warning: failed to focus window %d before resizing: %v\n", windowID, err)
	}
	
	// Use the Swift window manager binary for reliable window manipulation
	// Try to find window-manager binary in the same directory as the executable
	var windowManagerPath string
	execPath, err := os.Executable()
	if err == nil {
		windowManagerPath = filepath.Join(filepath.Dir(execPath), "window-manager")
		// Check if the file exists
		if _, err := os.Stat(windowManagerPath); os.IsNotExist(err) {
			windowManagerPath = ""
		}
	}
	
	// Fallback to relative path if not found
	if windowManagerPath == "" {
		windowManagerPath = "./window-manager"
		if _, err := os.Stat(windowManagerPath); os.IsNotExist(err) {
			// Try absolute path as last resort
			windowManagerPath = "/Users/ilmars/Dev/private/aerospace-scratchpad-1/window-manager"
		}
	}
	
	cmd := exec.Command(windowManagerPath, "resize", strconv.Itoa(windowID), strconv.Itoa(widthPercent), strconv.Itoa(heightPercent))
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		fmt.Printf("Warning: Swift window manager failed: %v, output: %s\n", err, output)
		// Don't return error - let it continue even if resize fails
	} else {
		fmt.Printf("Window resize output: %s\n", output)
	}
	
	return nil
}

// CenterWindow centers a window using move-mouse command
func (c *ExtendedAeroSpaceClient) CenterWindow(windowID int) error {
	cmd := exec.Command("aerospace", "move-mouse", "window-force-center", "--window-id", strconv.Itoa(windowID))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to center window: %w, output: %s", err, output)
	}
	return nil
}

// GeometrySpec represents window geometry specification
type GeometrySpec struct {
	WidthPercent  int
	HeightPercent int
}

// ParseGeometry parses geometry string in format "60%x90%"
func ParseGeometry(geometry string) (*GeometrySpec, error) {
	re := regexp.MustCompile(`^(\d+)%x(\d+)%$`)
	matches := re.FindStringSubmatch(geometry)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid geometry format: %s, expected format: 60%%x90%%", geometry)
	}
	
	width, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid width percentage: %s", matches[1])
	}
	
	height, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid height percentage: %s", matches[2])
	}
	
	return &GeometrySpec{
		WidthPercent:  width,
		HeightPercent: height,
	}, nil
}

// ApplyGeometry applies geometry to a window
func (c *ExtendedAeroSpaceClient) ApplyGeometry(windowID int, geometry string) error {
	spec, err := ParseGeometry(geometry)
	if err != nil {
		return err
	}
	
	// Try to set floating mode, but don't fail if it doesn't work
	// Some windows (like Arc) might not support floating mode
	if err := c.SetLayout(windowID, "floating"); err != nil {
		fmt.Printf("Info: Could not set floating layout for window %d, continuing anyway\n", windowID)
	}
	
	// Resize and center the window using percentage-based sizing
	// This should work regardless of floating mode
	if err := c.ResizeToPercentage(windowID, spec.WidthPercent, spec.HeightPercent); err != nil {
		return fmt.Errorf("failed to resize window to percentage: %w", err)
	}
	
	return nil
}