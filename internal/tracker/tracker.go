package tracker

import (
	"fmt"
	"regexp"
	"time"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/ilmars/aerospace-sticky/internal/logger"
	"github.com/ilmars/aerospace-sticky/internal/registry"
)

type WindowTracker interface {
	// GetMatchingWindows returns all windows that match any of the sticky patterns
	GetMatchingWindows(patterns []string) ([]aerospacecli.Window, error)
	
	// MoveWindowsToWorkspace moves all specified windows to the target workspace
	MoveWindowsToWorkspace(windows []aerospacecli.Window, workspace string) error
	
	// GetCurrentWorkspace returns the currently focused workspace
	GetCurrentWorkspace() (string, error)
	
	// FollowWorkspaceChanges continuously monitors for workspace changes and moves sticky windows
	FollowWorkspaceChanges(reg registry.Registry) error
}

type AerospaceTracker struct {
	client aerospacecli.AeroSpaceClient
}

func NewAerospaceTracker(client aerospacecli.AeroSpaceClient) *AerospaceTracker {
	return &AerospaceTracker{
		client: client,
	}
}

func (t *AerospaceTracker) GetMatchingWindows(patterns []string) ([]aerospacecli.Window, error) {
	if len(patterns) == 0 {
		return []aerospacecli.Window{}, nil
	}
	
	// Get all windows
	allWindows, err := t.client.GetAllWindows()
	if err != nil {
		return nil, fmt.Errorf("unable to get all windows: %w", err)
	}
	
	var matchingWindows []aerospacecli.Window
	
	// Compile all patterns
	var compiledPatterns []*regexp.Regexp
	for _, pattern := range patterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			logger.GetDefaultLogger().LogError("Invalid pattern", "pattern", pattern, "error", err)
			continue
		}
		compiledPatterns = append(compiledPatterns, regex)
	}
	
	// Find matching windows
	for _, window := range allWindows {
		for _, regex := range compiledPatterns {
			if regex.MatchString(window.AppName) {
				matchingWindows = append(matchingWindows, window)
				break // Don't add the same window multiple times
			}
		}
	}
	
	return matchingWindows, nil
}

func (t *AerospaceTracker) MoveWindowsToWorkspace(windows []aerospacecli.Window, workspace string) error {
	for _, window := range windows {
		// Skip if window is already in the target workspace
		if window.Workspace == workspace {
			continue
		}
		
		err := t.client.MoveWindowToWorkspace(window.WindowID, workspace)
		if err != nil {
			logger.GetDefaultLogger().LogError(
				"Failed to move window to workspace",
				"window", window,
				"workspace", workspace,
				"error", err,
			)
			// Continue with other windows instead of failing completely
			continue
		}
		
		logger.GetDefaultLogger().LogDebug(
			"Moved sticky window to workspace",
			"window", window,
			"workspace", workspace,
		)
	}
	
	return nil
}

func (t *AerospaceTracker) GetCurrentWorkspace() (string, error) {
	workspace, err := t.client.GetFocusedWorkspace()
	if err != nil {
		return "", fmt.Errorf("unable to get focused workspace: %w", err)
	}
	
	return workspace.Workspace, nil
}

func (t *AerospaceTracker) FollowWorkspaceChanges(reg registry.Registry) error {
	logger := logger.GetDefaultLogger()
	logger.LogInfo("Starting workspace change monitoring")
	
	var lastWorkspace string
	
	for {
		// Exit if no more sticky patterns
		if reg.IsEmpty() {
			logger.LogInfo("No sticky patterns found, exiting follower")
			break
		}
		
		// Get current workspace
		currentWorkspace, err := t.GetCurrentWorkspace()
		if err != nil {
			logger.LogError("Failed to get current workspace", "error", err)
			continue
		}
		
		// Check if workspace changed
		if currentWorkspace != lastWorkspace && lastWorkspace != "" {
			logger.LogDebug("Workspace changed", "from", lastWorkspace, "to", currentWorkspace)
			
			// Get sticky windows
			patterns := reg.GetPatterns()
			stickyWindows, err := t.GetMatchingWindows(patterns)
			if err != nil {
				logger.LogError("Failed to get matching windows", "error", err)
				continue
			}
			
			// Move sticky windows to current workspace
			if len(stickyWindows) > 0 {
				err = t.MoveWindowsToWorkspace(stickyWindows, currentWorkspace)
				if err != nil {
					logger.LogError("Failed to move sticky windows", "error", err)
				} else {
					logger.LogInfo("Moved sticky windows to new workspace", 
						"count", len(stickyWindows), 
						"workspace", currentWorkspace)
				}
			}
		}
		
		lastWorkspace = currentWorkspace
		
		// Poll every 500ms - could be made configurable
		// Note: In a real implementation, we might want to use aerospace callbacks
		// or a more efficient change detection mechanism
		time.Sleep(500 * time.Millisecond)
	}
	
	return nil
}

// Simple timer helper to avoid importing time in the interface
func timer(milliseconds int) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		// Simple polling implementation
		// In production, this could be replaced with proper async mechanisms
		for i := 0; i < milliseconds; i++ {
			// Basic busy wait - not efficient but works for proof of concept
		}
	}()
	return ch
}