/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"
	"github.com/spf13/cobra"
)

// moveCmd represents the move command
func MoveCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	moveCmd := &cobra.Command{
		Use:   "move <pattern>",
		Short: "Move a window to scratchpad",
		Long: `Move a window to scratchpad.

This command moves a window to the scratchpad.
It uses a regex to match the window name or title.
`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Error: missing pattern argument")
				cmd.Usage()
				return
			}

			windowNamePattern := args[0]
			// instantiate the regex
			regex, err := regexp.Compile(windowNamePattern)
			if err != nil {
				fmt.Println("Error: invalid window-name-pattern")
			}

			// Get all windows
			windows, err := aerospaceClient.GetAllWindows()
			if err != nil {
				fmt.Println("Error: unable to get windows")
				return
			}

			var movedCount int
			for _, window := range windows {
				// Check if the window name matches the regex
				if regex.MatchString(window.AppName) || regex.MatchString(window.WindowTitle) {
					// Move the window to the scratchpad
					fmt.Printf("Moving window %+v to scratchpad\n", window)
					err := aerospaceClient.MoveWindowToWorkspace(window.WindowID, "scratchpad")
					if err != nil {
						if strings.Contains(err.Error(), "already belongs to workspace") {
							fmt.Printf("Window '%+v' already belongs to scratchpad\n", window)
							movedCount++
							continue
						}

						fmt.Printf("Error: unable to move window '%+v' to scratchpad\n", window)
						fmt.Println(err)
						return
					}
					movedCount++
				}
			}

			if movedCount == 0 {
				fmt.Println("No windows matched the pattern")
			}
		},
	}

	return moveCmd
}
