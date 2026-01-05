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

	aerospacecli "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/focus"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/layout"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-ipc/pkg/client"
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

// AeroSpaceClient implements the AeroSpaceClient interface for interacting with AeroSpaceWM.
//
//revive:disable:exported
type AeroSpaceClient struct {
	ogClient *aerospacecli.AeroSpaceWM
	client   AeroSpaceWMClient // Interface for Windows()/Workspaces() access
	dryRun   bool
}

// ClientOpts defines options for creating a new AeroSpaceClient.
type ClientOpts struct {
	DryRun bool
}

// NewAeroSpaceClient creates a new AeroSpaceClient with the default settings.
func NewAeroSpaceClient(client AeroSpaceWMClient) *AeroSpaceClient {
	// Type assert to get the underlying *AeroSpaceWM for storage
	var ogClient *aerospacecli.AeroSpaceWM
	if realClient, ok := client.(*aerospacecli.AeroSpaceWM); ok {
		ogClient = realClient
	}
	return &AeroSpaceClient{
		ogClient: ogClient,
		client:   client,
		dryRun:   false, // Default dry-run is false
	}
}

// SetOptions the dry-run flag for the AeroSpaceClient.
func (c *AeroSpaceClient) SetOptions(opts ClientOpts) {
	c.dryRun = opts.DryRun
}

// Windows returns the windows service.
func (c *AeroSpaceClient) Windows() *windows.Service {
	return c.client.Windows()
}

// Workspaces returns the workspaces service.
func (c *AeroSpaceClient) Workspaces() *workspaces.Service {
	return c.client.Workspaces()
}

// Focus returns the focus service.
func (c *AeroSpaceClient) Focus() *focus.Service {
	return c.client.Focus()
}

// Layout returns the layout service.
func (c *AeroSpaceClient) Layout() *layout.Service {
	return c.client.Layout()
}

// GetAllWindows retrieves all windows managed by AeroSpaceWM.
func (c *AeroSpaceClient) GetAllWindows() ([]windows.Window, error) {
	return c.client.Windows().GetAllWindows()
}

func (c *AeroSpaceClient) GetAllWindowsByWorkspace(
	workspaceName string,
) ([]windows.Window, error) {
	return c.client.Windows().GetAllWindowsByWorkspace(workspaceName)
}

func (c *AeroSpaceClient) GetFocusedWindow() (*windows.Window, error) {
	return c.client.Windows().GetFocusedWindow()
}

func (c *AeroSpaceClient) SetFocusByWindowID(windowID int) error {
	if c.dryRun {
		fmt.Fprintf(os.Stdout, "[dry-run] SetFocusByWindowID(%d)\n", windowID)
		return nil
	}
	return c.client.Focus().SetFocusByWindowID(windowID)
}

// FocusNextTilingWindow moves focus to the next tiled window in depth-first order, ignoring floating windows.
// Equivalent to: `aerospace focus dfs-next --ignore-floating`.
func (c *AeroSpaceClient) FocusNextTilingWindow() error {
	if c.dryRun {
		fmt.Fprintln(os.Stdout, "[dry-run] FocusNextTilingWindow()")
		return nil
	}
	err := c.client.Focus().SetFocusByDFS("dfs-next", focus.SetFocusOpts{
		IgnoreFloating: true,
	})
	if err != nil {
		// Try dfs-prev if dfs-next fails
		err = c.client.Focus().SetFocusByDFS("dfs-prev", focus.SetFocusOpts{
			IgnoreFloating: true,
		})
		if err != nil {
			return fmt.Errorf("failed to focus next tiling window: %w", err)
		}
	}

	return nil
}

func (c *AeroSpaceClient) GetFocusedWorkspace() (*workspaces.Workspace, error) {
	return c.client.Workspaces().GetFocusedWorkspace()
}

func (c *AeroSpaceClient) MoveWindowToWorkspace(
	windowID int,
	workspaceName string,
) error {
	if c.dryRun {
		fmt.Fprintf(
			os.Stdout,
			"[dry-run] MoveWindowToWorkspace(windowID=%d, workspace=%s)\n",
			windowID,
			workspaceName,
		)
		return nil
	}
	return c.client.Workspaces().MoveWindowToWorkspaceWithOpts(
		workspaces.MoveWindowToWorkspaceArgs{
			WorkspaceName: workspaceName,
		},
		workspaces.MoveWindowToWorkspaceOpts{
			WindowID: &windowID,
		},
	)
}

func (c *AeroSpaceClient) SetLayout(windowID int, layoutName string) error {
	if c.dryRun {
		fmt.Fprintf(
			os.Stdout,
			"[dry-run] SetLayout(windowID=%d, layout=%s)\n",
			windowID,
			layoutName,
		)
		return nil
	}
	return c.client.Layout().SetLayout([]string{layoutName}, layout.SetLayoutOpts{
		WindowID: layout.IntPtr(windowID),
	})
}

func (c *AeroSpaceClient) Connection() client.AeroSpaceConnection {
	if c.client != nil {
		return c.client.Connection()
	}
	return c.ogClient.Connection()
}

func (c *AeroSpaceClient) CloseConnection() error {
	if c.dryRun {
		fmt.Fprintln(os.Stdout, "[dry-run] CloseConnection()")
		return nil
	}
	if c.client != nil {
		if closer, ok := c.client.(interface{ CloseConnection() error }); ok {
			return closer.CloseConnection()
		}
		return nil
	}
	return c.ogClient.CloseConnection()
}

// AeroSpaceWMClient defines the interface for clients that provide Windows(), Workspaces(), Focus(), and Layout() services.
type AeroSpaceWMClient interface {
	Windows() *windows.Service
	Workspaces() *workspaces.Service
	Focus() *focus.Service
	Layout() *layout.Service
	Connection() client.AeroSpaceConnection
}

// GetUnderlyingClient returns the underlying AeroSpaceWM client.
// This is needed for components that need direct access to Windows() and Workspaces() methods.
func (c *AeroSpaceClient) GetUnderlyingClient() AeroSpaceWMClient {
	if c.client != nil {
		return c.client
	}
	return c.ogClient
}

// ExtendedAeroSpaceClient wraps AeroSpaceClient with additional functionality
type ExtendedAeroSpaceClient struct {
	*AeroSpaceClient
}

// NewExtendedAeroSpaceClient creates a new extended client
func NewExtendedAeroSpaceClient(baseClient *AeroSpaceClient) *ExtendedAeroSpaceClient {
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
