/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"regexp"
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/ilmars/aerospace-sticky/internal/aerospace"
	"github.com/ilmars/aerospace-sticky/internal/constants"
	"github.com/ilmars/aerospace-sticky/internal/logger"
	"github.com/ilmars/aerospace-sticky/internal/stderr"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
func ShowCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	var workspace string
	var geometry string
	
	showCmd := &cobra.Command{
		Use:   "show <pattern>",
		Short: "Show a window from scratchpad or specified workspace",
		Long: `Show a window from the scratchpad (or custom workspace) in the current workspace.
By default, it will set the window to floating and focus on it.

Similar to I3/Sway WM, it will toggle show/hide the window if called multiple times.

Examples:
  # Show from default scratchpad
  aerospace-scratchpad show Terminal
  
  # Show from custom workspace with geometry and position
  aerospace-scratchpad show --workspace "dev" --geometry "80%x60%@bottom" Finder
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			logger.LogDebug("SHOW: start command", "args", args, "workspace", workspace, "geometry", geometry)
			
			// Create extended client for additional functionality
			extendedClient := aerospace.NewExtendedAeroSpaceClient(aerospaceClient)
			
			// Determine source workspace
			sourceWorkspace := workspace
			if sourceWorkspace == "" {
				sourceWorkspace = constants.DefaultScratchpadWorkspaceName
			}
			
			// Use default geometry if not specified
			targetGeometry := geometry
			if targetGeometry == "" {
				targetGeometry = constants.DefaultGeometry
			}
			
			windowNamePattern := args[0]
			windowNamePattern = strings.TrimSpace(windowNamePattern)
			if windowNamePattern == "" {
				stderr.Println("Error: <pattern> cannot be empty")
				return
			}

			windows, err := aerospaceClient.GetAllWindows()
			if err != nil {
				logger.LogError("SHOW: unable to get windows", "error", err)
				stderr.Println("Error: unable to get windows")
				return
			}
			logger.LogDebug("SHOW: retrieved windows", "windows", windows)

			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			if err != nil {
				logger.LogError("SHOW: unable to get focused workspace", "error", err)
				stderr.Println("Error: unable to get focused workspace")
				return
			}
			logger.LogDebug("SHOW: retrieved focused workspace", "workspace", focusedWorkspace)

			querier := aerospace.NewAerospaceQuerier(aerospaceClient)

			// instantiate the regex
			windowPattern, err := regexp.Compile(windowNamePattern)
			if err != nil {
				logger.LogError(
					"SHOW: unable to compile window pattern",
					"pattern",
					windowNamePattern,
					"error",
					err,
				)
				stderr.Println("Error: invalid window-name-pattern")
				return
			}
			logger.LogDebug("SHOW: compiled window pattern", "pattern", windowPattern)

			var windowsOutsideView []aerospacecli.Window
			var windowsInFocusedWorkspace []aerospacecli.Window
			var hasAtLeastOneWindowFocused bool
			for _, window := range windows {
				if !windowPattern.MatchString(window.AppName) {
					continue
				}

				var isWindowInFocusedWorkspace bool
				if window.Workspace == "" {
					isWindowInFocusedWorkspace, err = querier.IsWindowInWorkspace(
						window.WindowID,
						focusedWorkspace.Workspace,
					)
					if err != nil {
						stderr.Printf("Error: unable to check if window '%+v' is in workspace '%s'\n", window, focusedWorkspace.Workspace)
						return
					}

				} else {
					isWindowInFocusedWorkspace = window.Workspace == focusedWorkspace.Workspace

				}
				if isWindowInFocusedWorkspace {
					windowsInFocusedWorkspace = append(windowsInFocusedWorkspace, window)

					isWindowFocused, err := querier.IsWindowFocused(window.WindowID)
					if err != nil {
						stderr.Printf("Error: unable to check if window '%+v' is focused\n", window)
						return
					}

					// Make sure that once hasAtLeastOneWindowFocused is true, it will remain true
					hasAtLeastOneWindowFocused = hasAtLeastOneWindowFocused || isWindowFocused
				} else {
					// Only consider windows that are in the source workspace as "outside view"
					// For backward compatibility, if no workspace specified, check all workspaces
					if workspace == "" {
						// Original behavior: any window not in focused workspace
						windowsOutsideView = append(windowsOutsideView, window)
					} else {
						// New behavior: only windows in the specified source workspace
						var isInSourceWorkspace bool
						if window.Workspace == "" {
							isInSourceWorkspace, err = querier.IsWindowInWorkspace(window.WindowID, sourceWorkspace)
							if err != nil {
								stderr.Printf("Error: unable to check if window '%+v' is in workspace '%s'\n", window, sourceWorkspace)
								return
							}
						} else {
							isInSourceWorkspace = window.Workspace == sourceWorkspace
						}
						if isInSourceWorkspace {
							windowsOutsideView = append(windowsOutsideView, window)
						}
					}

				}

				logger.LogDebug(
					"SHOW: loop",
					"windowsOutsideView", windowsOutsideView,
					"windowsInFocusedWorkspace", windowsInFocusedWorkspace,
					"hasAtLeastOneWindowFocused", hasAtLeastOneWindowFocused,
				)
			}

			logger.LogDebug(
				"SHOW: filtered windows",
				"windowsOutsideView", windowsOutsideView,
				"windowsInFocusedWorkspace", windowsInFocusedWorkspace,
				"hasAtLeastOneWindowFocused", hasAtLeastOneWindowFocused,
			)

			// If no matching windows found in the expected source workspace, search all workspaces as fallback
			if len(windowsOutsideView) == 0 && len(windowsInFocusedWorkspace) == 0 {
				logger.LogDebug("SHOW: no windows found in expected workspace, searching all workspaces for stuck scratchpads")
				
				for _, window := range windows {
					if !windowPattern.MatchString(window.AppName) {
						continue
					}
					
					// Skip windows already found in the focused workspace
					if window.Workspace == focusedWorkspace.Workspace {
						continue
					}
					
					// This is a matching window stuck in some other workspace
					logger.LogDebug("SHOW: found stuck scratchpad window", "window", window, "stuckWorkspace", window.Workspace)
					fmt.Printf("Found scratchpad window '%s' stuck in workspace '%s'\n", window.AppName, window.Workspace)
					windowsOutsideView = append(windowsOutsideView, window)
				}
				
				logger.LogDebug("SHOW: fallback search found", "windowsOutsideView", windowsOutsideView)
			}

			// Move other scratchpads back to their respective workspaces before showing this one
			// Pass the source workspace so cleanup respects the --workspace flag
			err = moveOtherScratchpadsToWorkspaces(aerospaceClient, windows, windowPattern, focusedWorkspace.Workspace, sourceWorkspace)
			if err != nil {
				logger.LogError("SHOW: unable to move other scratchpads to workspaces", "error", err)
				// Don't return error here, just log it and continue
			}

			for _, window := range windowsOutsideView {
				err := sendToFocusedWorkspace(
					aerospaceClient,
					extendedClient,
					window,
					focusedWorkspace,
					!hasAtLeastOneWindowFocused,
					targetGeometry,
				)
				if err != nil {
					stderr.Printf(
						"Error: unable to move window '%+v' to scratchpad\n%s",
						window,
						err,
					)
					return
				}
			}

			if len(windowsInFocusedWorkspace) == 0 && len(windowsOutsideView) == 0 {
				stderr.Println("Error: no windows matched the pattern '%s'", windowNamePattern)
				return
			}

			// NOTE: To avoid the ping pong of windows, so priority is
			// for bringing windows to the focused workspace
			if len(windowsOutsideView) > 0 {
				// Make sure to bring the remaining matched windows to the front
				for _, window := range windowsInFocusedWorkspace {
					err = aerospaceClient.SetFocusByWindowID(window.WindowID)
					if err != nil {
						stderr.Printf(
							"Error: unable to set focus to window '%+v'\n%s",
							window,
							err,
						)
						return
					}
					logger.LogDebug("SHOW: set focus to window", "window", window)
					fmt.Printf("Window '%+v' is focused\n", window)
				}

				return
			}

			for _, window := range windowsInFocusedWorkspace {
				logger.LogDebug(
					"SHOW: processing window in focused workspace",
					"window", window,
					"hasAtLeastOneWindowFocused", hasAtLeastOneWindowFocused,
				)
				if hasAtLeastOneWindowFocused {
					if err = sendToScratchpad(aerospaceClient, window, sourceWorkspace); err != nil {
						logger.LogDebug(
							"Error: unable to move window '%+v' to workspace\n%s",
							"window", window,
							"error", err,
						)
						continue
					}
				} else {
					err = aerospaceClient.SetFocusByWindowID(window.WindowID)
					if err != nil {
						stderr.Printf(
							"Error: unable to set focus to window '%+v'\n%s",
							window,
							err,
						)
						return
					}
					fmt.Printf("Window '%+v' is focused\n", window)
				}
			}
		},
	}

	// Add flags
	showCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Source workspace (defaults to .scratchpad)")
	showCmd.Flags().StringVarP(&geometry, "geometry", "g", "", "Window geometry when pulled to current workspace (e.g., 60%x90%@bottom)")

	return showCmd
}

func sendToScratchpad(
	aerospaceClient aerospacecli.AeroSpaceClient,
	window aerospacecli.Window,
	targetWorkspace string,
) error {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("SHOW: sendToScratchpad ", "window", window, "targetWorkspace", targetWorkspace)

	err := aerospaceClient.MoveWindowToWorkspace(
		window.WindowID,
		targetWorkspace,
	)
	logger.LogDebug(
		"SHOW: after aerospaceClient.MoveWindowToWorkspace",
		"window", window,
		"to-workspace", targetWorkspace,
		"error", err,
	)
	if err != nil {
		return err
	}

	// Set to floating layout for all workspaces (no automatic fullscreen)
	err = aerospaceClient.SetLayout(
		window.WindowID,
		"floating",
	)
	if err != nil {
		fmt.Printf(
			"Warn: unable to set layout for window '%+v' to floating\n%s",
			window,
			err,
		)
	}
	logger.LogDebug(
		"SHOW: set floating layout",
		"window", window,
		"workspace", targetWorkspace,
		"error", err,
	)

	fmt.Printf("Window '%+v' hidden to workspace %s\n", window, targetWorkspace)
	return nil
}

func sendToFocusedWorkspace(
	aerospaceClient aerospacecli.AeroSpaceClient,
	extendedClient *aerospace.ExtendedAeroSpaceClient,
	window aerospacecli.Window,
	focusedWorkspace *aerospacecli.Workspace,
	shouldSetFocus bool,
	geometry string,
) error {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("SHOW: sendToFocusedWorkspace called", "window", window, "targetWorkspace", focusedWorkspace.Workspace, "shouldSetFocus", shouldSetFocus)
	
	if focusedWorkspace == nil {
		return fmt.Errorf("focused workspace is nil")
	}

	if err := aerospaceClient.MoveWindowToWorkspace(
		window.WindowID,
		focusedWorkspace.Workspace,
	); err != nil {
		return fmt.Errorf("unable to move window '%+v' to workspace '%s': %w", window, focusedWorkspace.Workspace, err)
	}

	// FIRST: Always focus the window immediately after moving it
	if err := aerospaceClient.SetFocusByWindowID(window.WindowID); err != nil {
		return fmt.Errorf("unable to set focus to window '%+v': %w", window, err)
	}

	// Focus the window using AeroSpace's built-in focus mechanism (faster than external commands)
	logger.LogDebug("SHOW: focusing window", "appName", window.AppName, "windowID", window.WindowID)

	// SECOND: Apply geometry if specified (this will focus again internally)
	if geometry != "" {
		if err := extendedClient.ApplyGeometry(window.WindowID, geometry); err != nil {
			return fmt.Errorf("unable to apply geometry to window '%+v': %w", window, err)
		}
		
		// THIRD: Focus one more time after geometry changes
		if err := aerospaceClient.SetFocusByWindowID(window.WindowID); err != nil {
			return fmt.Errorf("unable to set focus to window '%+v' after geometry: %w", window, err)
		}
	}

	fmt.Printf("Window '%+v' is summoned\n", window)
	return nil
}

func moveOtherScratchpadsToWorkspaces(
	aerospaceClient aerospacecli.AeroSpaceClient,
	windows []aerospacecli.Window,
	currentPatternRegex *regexp.Regexp,
	currentWorkspace string,
	sourceWorkspace string,
) error {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("SHOW: moveOtherScratchpadsToWorkspaces", "currentWorkspace", currentWorkspace, "sourceWorkspace", sourceWorkspace)

	// Find windows that are in the current workspace but don't match the current pattern
	// Only move windows that are known scratchpad applications to avoid moving regular tiled windows
	for _, window := range windows {
		// Skip if this window matches the current pattern (it's the one being shown)
		if currentPatternRegex.MatchString(window.AppName) {
			continue
		}

		// Debug: Log all windows we're considering
		logger.LogDebug("SHOW: considering window for scratchpad cleanup", "appName", window.AppName, "windowID", window.WindowID, "workspace", window.Workspace)

		// Only process applications that are configured as scratchpad apps
		defaultTargetWorkspace, isScratchpadApp := constants.DefaultScratchpadAppWorkspaces[window.AppName]
		if !isScratchpadApp {
			// This is not a known scratchpad app, skip it to avoid moving regular tiled windows
			logger.LogDebug("SHOW: skipping non-scratchpad app", "appName", window.AppName)
			continue
		}

		// If this app matches the current pattern AND there's a custom source workspace specified,
		// don't move it - let the main show logic handle it with the custom workspace
		if currentPatternRegex.MatchString(window.AppName) && sourceWorkspace != constants.DefaultScratchpadWorkspaceName {
			logger.LogDebug("SHOW: skipping current pattern app with custom workspace", "appName", window.AppName, "sourceWorkspace", sourceWorkspace)
			continue
		}

		targetWorkspace := defaultTargetWorkspace
		logger.LogDebug("SHOW: found scratchpad app to potentially move", "appName", window.AppName, "targetWorkspace", targetWorkspace)

		// Check if window is in the current workspace
		var isInCurrentWorkspace bool
		if window.Workspace == "" {
			// Use querier to check workspace
			querier := aerospace.NewAerospaceQuerier(aerospaceClient)
			var err error
			isInCurrentWorkspace, err = querier.IsWindowInWorkspace(window.WindowID, currentWorkspace)
			if err != nil {
				logger.LogError("SHOW: unable to check if window is in current workspace", "window", window, "workspace", currentWorkspace, "error", err)
				continue
			}
		} else {
			isInCurrentWorkspace = window.Workspace == currentWorkspace
		}

		if !isInCurrentWorkspace {
			continue
		}

		// This is a scratchpad app in the current workspace, move it back to its assigned workspace
		if err := aerospaceClient.MoveWindowToWorkspace(window.WindowID, targetWorkspace); err != nil {
			logger.LogError("SHOW: unable to move scratchpad window back to workspace", "window", window, "targetWorkspace", targetWorkspace, "error", err)
			continue
		}

		// Set to floating layout for scratchpad apps
		if err := aerospaceClient.SetLayout(window.WindowID, "floating"); err != nil {
			logger.LogError("SHOW: unable to set floating layout for scratchpad", "window", window, "error", err)
			// Continue anyway, layout is not critical
		}

		logger.LogDebug("SHOW: moved scratchpad app back to workspace", "window", window, "appName", window.AppName, "targetWorkspace", targetWorkspace)
		fmt.Printf("Moved scratchpad '%s' back to workspace %s\n", window.AppName, targetWorkspace)
	}

	return nil
}
