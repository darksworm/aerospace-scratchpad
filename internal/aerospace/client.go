package aerospace

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
)

//go:embed window-manager
var windowManagerBinary []byte

// Cache for the extracted binary path
var (
	windowManagerPath string
	extractOnce       sync.Once
)

// extractWindowManagerBinary extracts the embedded binary to a temporary location
func extractWindowManagerBinary() (string, error) {
	var err error
	extractOnce.Do(func() {
		// Create a temporary file
		tmpFile, tmpErr := os.CreateTemp("", "window-manager-*")
		if tmpErr != nil {
			err = fmt.Errorf("failed to create temp file: %w", tmpErr)
			return
		}
		defer tmpFile.Close()

		// Write the embedded binary data
		if _, tmpErr := tmpFile.Write(windowManagerBinary); tmpErr != nil {
			err = fmt.Errorf("failed to write binary data: %w", tmpErr)
			return
		}

		// Make it executable
		if tmpErr := tmpFile.Chmod(0755); tmpErr != nil {
			err = fmt.Errorf("failed to make binary executable: %w", tmpErr)
			return
		}

		windowManagerPath = tmpFile.Name()
	})

	if err != nil {
		return "", err
	}
	return windowManagerPath, nil
}

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
	return c.ResizeToPercentageWithPosition(windowID, widthPercent, heightPercent, "center")
}

// ResizeToPercentageWithPosition resizes and positions a window using Swift window manager
func (c *ExtendedAeroSpaceClient) ResizeToPercentageWithPosition(windowID int, widthPercent, heightPercent int, position string) error {
	// Focus the window first to ensure it's active
	if err := c.SetFocusByWindowID(windowID); err != nil {
		fmt.Printf("Warning: failed to focus window %d before resizing: %v\n", windowID, err)
	}
	
	// Extract the embedded Swift window manager binary
	windowManagerPath, err := extractWindowManagerBinary()
	if err != nil {
		return fmt.Errorf("failed to extract window manager binary: %w", err)
	}
	
	cmd := exec.Command(windowManagerPath, "resize", strconv.Itoa(windowID), strconv.Itoa(widthPercent), strconv.Itoa(heightPercent), position)
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
	Position      string
}

// ParseGeometry parses geometry string in format "60%x90%" or "60%x90%@position"
func ParseGeometry(geometry string) (*GeometrySpec, error) {
	// Split geometry and position if @ is present
	parts := strings.Split(geometry, "@")
	geometryPart := parts[0]
	position := "center" // default position
	
	if len(parts) == 2 {
		position = strings.TrimSpace(parts[1])
	} else if len(parts) > 2 {
		return nil, fmt.Errorf("invalid geometry format: %s, expected format: 60%%x90%%@position", geometry)
	}
	
	re := regexp.MustCompile(`^(\d+)%x(\d+)%$`)
	matches := re.FindStringSubmatch(geometryPart)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid geometry format: %s, expected format: 60%%x90%%[@position]", geometry)
	}
	
	width, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid width percentage: %s", matches[1])
	}
	
	height, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid height percentage: %s", matches[2])
	}
	
	// Validate position
	validPositions := []string{"center", "top", "bottom", "left", "right"}
	isValidPosition := false
	for _, validPos := range validPositions {
		if strings.ToLower(position) == validPos {
			isValidPosition = true
			break
		}
	}
	if !isValidPosition {
		return nil, fmt.Errorf("invalid position: %s, valid positions: %s", position, strings.Join(validPositions, ", "))
	}
	
	return &GeometrySpec{
		WidthPercent:  width,
		HeightPercent: height,
		Position:      strings.ToLower(position),
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
	
	// Resize and position the window using percentage-based sizing
	// This should work regardless of floating mode
	if err := c.ResizeToPercentageWithPosition(windowID, spec.WidthPercent, spec.HeightPercent, spec.Position); err != nil {
		return fmt.Errorf("failed to resize window to percentage: %w", err)
	}
	
	return nil
}