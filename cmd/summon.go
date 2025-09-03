/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"regexp"
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/ilmars/aerospace-sticky/internal/aerospace"
	"github.com/ilmars/aerospace-sticky/internal/cli"
	"github.com/ilmars/aerospace-sticky/internal/constants"
	"github.com/ilmars/aerospace-sticky/internal/logger"
	"github.com/ilmars/aerospace-sticky/internal/stderr"
	"github.com/spf13/cobra"
)

func SummonCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	var geometry string
	
	summonCmd := &cobra.Command{
		Use:   "summon <pattern>",
		Short: "Summon a window from scratchpad",
		Long: `Summon a window from the scratchpad to the current workspace.

This command brings a window from the scratchpad to the current workspace using a regex to match the window name or title.
It properly searches scratchpad workspaces and handles geometry for consistent sizing across screens.
`,

		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			cli.ValidateAllNonEmpty,
		),

		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			windowNamePattern := strings.TrimSpace(args[0])

			// Use default geometry if not specified
			targetGeometry := geometry
			if targetGeometry == "" {
				targetGeometry = constants.DefaultGeometry
			}

			logger.LogDebug("SUMMON: start command", "pattern", windowNamePattern, "geometry", targetGeometry)

			// Create extended client for geometry support
			extendedClient := aerospace.NewExtendedAeroSpaceClient(aerospaceClient)

			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			if err != nil {
				stderr.Println("Error: unable to get focused workspace")
				return
			}

			// Compile the regex pattern
			windowPattern, err := regexp.Compile(windowNamePattern)
			if err != nil {
				stderr.Println("Error: invalid app-name-pattern")
				return
			}

			// Search for matching windows in scratchpad workspaces first
			matchingWindows, err := findScratchpadWindows(aerospaceClient, windowPattern)
			if err != nil {
				stderr.Printf("Error: unable to find scratchpad windows: %v\n", err)
				return
			}

			logger.LogDebug("SUMMON: found matching windows", "windows", matchingWindows)

			if len(matchingWindows) == 0 {
				stderr.Printf("Error: no scratchpad windows matched the pattern '%s'\n", windowNamePattern)
				return
			}

			// Summon the first matching window
			window := matchingWindows[0]
			err = summonWindowToWorkspace(aerospaceClient, extendedClient, window, focusedWorkspace, targetGeometry)
			if err != nil {
				stderr.Printf("Error: unable to summon window '%+v': %v\n", window, err)
				return
			}

			fmt.Printf("Window '%+v' is summoned\n", window)
		},
	}

	// Add geometry flag
	summonCmd.Flags().StringVarP(&geometry, "geometry", "g", "", fmt.Sprintf("Window geometry (default: %s)", constants.DefaultGeometry))

	return summonCmd
}

// findScratchpadWindows searches for windows matching the pattern in scratchpad workspaces
func findScratchpadWindows(aerospaceClient aerospacecli.AeroSpaceClient, pattern *regexp.Regexp) ([]aerospacecli.Window, error) {
	logger := logger.GetDefaultLogger()
	var matchingWindows []aerospacecli.Window

	// Get all known scratchpad workspace names
	scratchpadWorkspaces := getScratchpadWorkspaces()
	
	for _, workspaceName := range scratchpadWorkspaces {
		logger.LogDebug("SUMMON: searching workspace", "workspace", workspaceName)
		
		windows, err := aerospaceClient.GetAllWindowsByWorkspace(workspaceName)
		if err != nil {
			// Workspace might not exist, continue to next one
			logger.LogDebug("SUMMON: workspace not found", "workspace", workspaceName, "error", err)
			continue
		}

		for _, window := range windows {
			if pattern.MatchString(window.AppName) {
				logger.LogDebug("SUMMON: found matching window", "window", window, "workspace", workspaceName)
				matchingWindows = append(matchingWindows, window)
			}
		}
	}

	return matchingWindows, nil
}

// getScratchpadWorkspaces returns a list of all known scratchpad workspace names
func getScratchpadWorkspaces() []string {
	workspaceSet := make(map[string]bool)
	
	// Add the default scratchpad workspace
	workspaceSet[constants.DefaultScratchpadWorkspaceName] = true
	
	// Add all workspaces from the scratchpad app mappings
	for _, workspace := range constants.DefaultScratchpadAppWorkspaces {
		workspaceSet[workspace] = true
	}
	
	// Convert to slice
	var workspaces []string
	for workspace := range workspaceSet {
		workspaces = append(workspaces, workspace)
	}
	
	return workspaces
}

// summonWindowToWorkspace moves a window to the focused workspace with proper geometry
func summonWindowToWorkspace(
	aerospaceClient aerospacecli.AeroSpaceClient,
	extendedClient *aerospace.ExtendedAeroSpaceClient,
	window aerospacecli.Window,
	focusedWorkspace *aerospacecli.Workspace,
	geometry string,
) error {
	logger := logger.GetDefaultLogger()
	
	// Move window to focused workspace
	err := aerospaceClient.MoveWindowToWorkspace(window.WindowID, focusedWorkspace.Workspace)
	if err != nil {
		return fmt.Errorf("unable to move window to workspace '%s': %w", focusedWorkspace.Workspace, err)
	}

	// Focus the window immediately after moving
	err = aerospaceClient.SetFocusByWindowID(window.WindowID)
	if err != nil {
		return fmt.Errorf("unable to set focus to window: %w", err)
	}

	// Set geometry for proper sizing across screens
	if geometry != "" {
		err = extendedClient.ApplyGeometry(window.WindowID, geometry)
		if err != nil {
			logger.LogDebug("SUMMON: unable to apply geometry", "window", window, "geometry", geometry, "error", err)
			// Don't fail the command for geometry errors, just log them
		}
	}

	logger.LogDebug("SUMMON: successfully summoned window", "window", window, "workspace", focusedWorkspace.Workspace, "geometry", geometry)
	return nil
}
