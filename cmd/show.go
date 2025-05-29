/*
Copyright © 2025 Cristian Oliveira licence@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
func ShowCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show <pattern>",
		Short: "Show a window from scratchpad",
		Long: `Show a window from the scratchpad in the current workspace.
By default, it will set the window to floating and focus on it.

Similar to I3/Sway WM, it will toggle show/hide the window if called multiple times.
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()
			defer func() {
				log.Printf("Finished in %s", time.Since(start))
			}()
			windowNamePattern := args[0]
			windowNamePattern = strings.TrimSpace(windowNamePattern)
			if windowNamePattern == "" {
				stderr.Println("Error: <pattern> cannot be empty")
				return
			}

			windows, err := aerospaceClient.GetAllWindows()
			if err != nil {
				stderr.Println("Error: unable to get windows")
				return
			}
			log.Printf("GetAllWindows took %s", time.Since(start))

			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			if err != nil {
				stderr.Println("Error: unable to get focused workspace")
				return
			}
			log.Printf("GetFocusedWorkspace took %s", time.Since(start))

			querier := aerospace.NewAerospaceQuerier(aerospaceClient)

			// instantiate the regex
			windowPattern, err := regexp.Compile(windowNamePattern)
			if err != nil {
				stderr.Println("Error: invalid window-name-pattern")
				return
			}

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
				log.Printf("IsWindowInWorkspace took %s", time.Since(start))

				if isWindowInFocusedWorkspace {
					isWindowFocused, err := querier.IsWindowFocused(window.WindowID)
					if err != nil {
						stderr.Printf("Error: unable to check if window '%+v' is focused\n", window)
						return
					}
					if isWindowFocused {
						if err = aerospaceClient.MoveWindowToWorkspace(
							window.WindowID,
							constants.DefaultScratchpadWorkspaceName,
						); err != nil {
							stderr.Printf("Error: unable to move window '%+v' to scratchpad\n", window)
							return
						}

						aerospaceClient.SetLayout(
							window.WindowID,
							"floating",
						)

						fmt.Printf("Window '%+v' hidden to scratchpad\n", window)
						return
					}

					aerospaceClient.SetFocusByWindowID(window.WindowID)
					log.Printf("SetFocusByWindowID took %s", time.Since(start))
					fmt.Printf("Setting focus to window '%s'\n", window.AppName)
					return
				}

				log.Printf("Bef MoveWindowToWorkspace took %s", time.Since(start))
				if err = aerospaceClient.MoveWindowToWorkspace(
					window.WindowID,
					focusedWorkspace.Workspace,
				); err != nil {
					stderr.Printf("Error: unable to move window '%+v' to workspace '%s'\n", window, focusedWorkspace.Workspace)
					return
				}
				log.Printf("MoveWindowToWorkspace took %s", time.Since(start))

				if err = aerospaceClient.SetFocusByWindowID(window.WindowID); err != nil {
					stderr.Printf("Error: unable to set focus to window '%+v'\n", window)
					return
				}
				log.Printf("SetFocusByWindowID took %s", time.Since(start))

				fmt.Printf("Window '%+v' is summoned\n", window)
			}
		},
	}

	return showCmd
}
