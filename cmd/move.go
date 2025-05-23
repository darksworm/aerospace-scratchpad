/*
Copyright © 2025 Cristian Oliveira licence@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"regexp"
	"strings"
	_ "net/http/pprof"

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
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				fmt.Println("Error: missing pattern argument")
				cmd.Usage()
				return nil
			}

			windowNamePattern := args[0]
			if windowNamePattern == "" {
				fmt.Println("Error: empty pattern argument")
				cmd.Usage()
				return nil
			}
			// instantiate the regex
			regex, err := regexp.Compile(windowNamePattern)
			if err != nil {
				fmt.Println("Error: invalid window-name-pattern")
			}

			// Get all windows
			windows, err := aerospaceClient.GetAllWindows()
			if err != nil {
				return fmt.Errorf("Error: unable to get windows")
			}

			var movedCount int
			for _, window := range windows {
				// Check if the window name matches the regex
				if regex.MatchString(window.AppName) {
					// Move the window to the scratchpad
					fmt.Printf("Moving window %+v to scratchpad\n", window)
					err := aerospaceClient.MoveWindowToWorkspace(window.WindowID, "scratchpad")
					if err != nil {
						if strings.Contains(err.Error(), "already belongs to workspace") {
							return fmt.Errorf("Window '%+v' already belongs to scratchpad\n", window)
						}

						return fmt.Errorf("Error: unable to move window '%+v' to scratchpad\n", window)
					}

					conn := aerospaceClient.(*aerospacecli.AeroSpaceWM).Conn
					conn.SendCommand(
						"layout",
						[]string{
							"floating",
							"--window-id",
							fmt.Sprintf("%d", window.WindowID),
						},
					)

					movedCount++
				}
			}

			if movedCount == 0 {
				fmt.Println("No windows matched the pattern")
			}

			return nil
		},
	}

	return moveCmd
}
