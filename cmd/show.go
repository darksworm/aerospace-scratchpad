/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"regexp"

	_ "net/http/pprof"

	"github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
func ShowCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show <pattern>",
		Short: "Show a window from scratchpad",
		Long: `Show a window from scratchpad on the current workspace.
By default, it will set the window to floating and focus it.

Similar to SwayWM it will toggle show/hide the window if called multiple times.
`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Error: missing pattern argument")
				os.Exit(1)
				return
			}

			windowNamePattern := args[0]
			// instantiate the regex
			regex, err := regexp.Compile(windowNamePattern)
			if err != nil {
				fmt.Println("Error: invalid window-name-pattern")
				return
			}

			// Get all windows from scratchpad workspace
			windows, err := aerospaceClient.GetAllWindowsByWorkspace("scratchpad")
			if err != nil {
				fmt.Println("Error: unable to get windows")
				return
			}

			var windowsToShow []aerospacecli.Window
			for _, window := range windows {
				if regex.MatchString(window.AppName) {
					windowsToShow = append(windowsToShow, window)
				}
			}

			if len(windowsToShow) == 0 {
				fmt.Println("No windows found matching the pattern")
				os.Exit(1)
				return
			}

			// Set the windows to floating
			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			for _, window := range windowsToShow {
				err = aerospaceClient.MoveWindowToWorkspace(
					window.WindowID,
					focusedWorkspace.Workspace,
				)

				if err != nil {
					fmt.Printf("Window '%+v' already belongs to workspace '%s'\n", window, focusedWorkspace.Workspace)
					continue
				}

				err = aerospaceClient.SetFocusByWindowID(window.WindowID)
				if err != nil {
					fmt.Printf("Error: unable to set focus to window '%+v'\n", window)
					continue
				}
				fmt.Printf("Window '%+v' shown from scratchpad\n", window)
			}
		},
	}

	return showCmd
}
