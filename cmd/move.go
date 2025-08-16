/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"log"
	_ "net/http/pprof"
	"regexp"
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/ilmars/aerospace-sticky/internal/aerospace"
	"github.com/ilmars/aerospace-sticky/internal/constants"
	"github.com/ilmars/aerospace-sticky/internal/logger"
	"github.com/ilmars/aerospace-sticky/internal/stderr"
	"github.com/spf13/cobra"
)

// moveCmd represents the move command
func MoveCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	var workspace string
	var fullscreen bool
	
	moveCmd := &cobra.Command{
		Use:   "move <pattern>",
		Short: "Move a window to scratchpad or specified workspace",
		Long: `Move a window to the scratchpad or a specified workspace.

This command moves a window to the scratchpad (or custom workspace) using a regex to match the app name.
If no pattern is provided, it moves the currently focused window.

Examples:
  # Move to default scratchpad
  aerospace-scratchpad move Terminal
  
  # Move to custom workspace and make fullscreen
  aerospace-scratchpad move --workspace "dev" --fullscreen Terminal
`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			logger.LogDebug("MOVE: start command", "args", args, "workspace", workspace, "fullscreen", fullscreen)
			
			// Create extended client for additional functionality
			extendedClient := aerospace.NewExtendedAeroSpaceClient(aerospaceClient)
			
			// Determine target workspace
			targetWorkspace := workspace
			if targetWorkspace == "" {
				targetWorkspace = constants.DefaultScratchpadWorkspaceName
			}
			
			var windowNamePattern string
			if len(args) == 0 {
				windowNamePattern = ""
			} else {
				windowNamePattern = strings.TrimSpace(args[0])
			}

			if windowNamePattern == "" {
				focusedWindow, err := aerospaceClient.GetFocusedWindow()
				logger.LogDebug(
					"MOVE: retrieving focused window",
					"focusedWindow", focusedWindow,
					"error", err,
				)
				if err != nil {
					stderr.Println("Error: unable to get focused window: %v", err)
					return
				}
				if focusedWindow == nil {
					stderr.Println("Error: no focused window found")
					return
				}
				windowNamePattern = fmt.Sprintf("^%s$", focusedWindow.AppName)
				logger.LogDebug(
					"MOVE: using focused window app name as pattern",
					"windowNamePattern", windowNamePattern,
				)
			}

			// instantiate the regex
			regex, err := regexp.Compile(windowNamePattern)
			if err != nil {
				logger.LogError(
					"MOVE: error compiling regex",
					"windowNamePattern", windowNamePattern,
					"error", err,
				)
				log.Fatalf("Error compiling regex '%s': %v", windowNamePattern, err)
			}

			// Get all windows
			windows, err := aerospaceClient.GetAllWindows()
			logger.LogDebug(
				"MOVE: retrieved all windows",
				"windows", windows,
				"error", err,
			)
			if err != nil {
				stderr.Println("Error: unable to get windows")
				return
			}

			var movedCount int
			for _, window := range windows {
				if !regex.MatchString(window.AppName) {
					continue
				}

				// Move the window to the target workspace
				fmt.Printf("Moving window %+v to workspace %s\n", window, targetWorkspace)
				err := aerospaceClient.MoveWindowToWorkspace(
					window.WindowID,
					targetWorkspace,
				)
				logger.LogDebug(
					"MOVE: moving window to workspace",
					"window", window,
					"workspace", targetWorkspace,
					"error", err,
				)
				if err != nil {
					if strings.Contains(err.Error(), "already belongs to workspace") {
						continue
					}

					stderr.Println("Error: unable to move window '%+v' to workspace %s", window, targetWorkspace)
					// exit loop
					return
				}

				// Apply fullscreen by default for specific workspaces, floating for default scratchpad
				shouldUseFullscreen := fullscreen || (targetWorkspace != constants.DefaultScratchpadWorkspaceName)
				
				if shouldUseFullscreen {
					err = extendedClient.SetFullscreen(window.WindowID, true)
					logger.LogDebug(
						"MOVE: setting window to fullscreen",
						"window", window,
						"workspace", targetWorkspace,
						"reason", func() string {
							if fullscreen {
								return "explicit fullscreen flag"
							}
							return "default for specific workspace"
						}(),
						"error", err,
					)
					if err != nil {
						fmt.Printf(
							"warn: unable to set fullscreen for window '%+v'\n%s",
							window,
							err,
						)
					}
				} else {
					// Set to floating layout for default scratchpad workspace
					err = aerospaceClient.SetLayout(
						window.WindowID,
						"floating",
					)
					logger.LogDebug(
						"MOVE: setting layout to floating",
						"window", window,
						"layout", "floating",
						"error", err,
					)
					if err != nil {
						fmt.Printf(
							"warn: unable to set layout for window '%+v' to floating\n%s",
							window,
							err,
						)
					}
				}

				movedCount++
			}

			if movedCount == 0 {
				logger.LogDebug(
					"MOVE: no windows matched the pattern",
					"pattern", windowNamePattern,
				)

				stderr.Println(
					"Error: no windows matched the pattern '%s'",
					windowNamePattern,
				)
				return
			}
		},
	}

	// Add flags
	moveCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Target workspace (defaults to .scratchpad)")
	moveCmd.Flags().BoolVarP(&fullscreen, "fullscreen", "f", false, "Make window fullscreen in target workspace")

	return moveCmd
}
