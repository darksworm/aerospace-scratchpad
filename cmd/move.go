/*
Copyright © 2025 Cristian Oliveira licence@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	_ "net/http/pprof"
	"regexp"
	"strings"

	"github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/cli"
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
		Args:  cobra.MatchAll(
			cobra.ExactArgs(1),
			cli.ValidateAllNonEmpty,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowNamePattern := args[0]

			// instantiate the regex
			regex, err := regexp.Compile(windowNamePattern)
			if err != nil {
				fmt.Println("invalid window-name-pattern")
			}

			// Get all windows
			windows, err := aerospaceClient.GetAllWindows()
			if err != nil {
				return fmt.Errorf("unable to get windows")
			}

			var movedCount int
			for _, window := range windows {
				if !regex.MatchString(window.AppName) {
					continue
				}

				// Move the window to the scratchpad
				fmt.Printf("Moving window %+v to scratchpad\n", window)
				err := aerospaceClient.MoveWindowToWorkspace(window.WindowID, "scratchpad")
				if err != nil {
					if strings.Contains(err.Error(), "already belongs to workspace") {
						return fmt.Errorf("Window '%+v' already belongs to scratchpad\n", window)
					}

					return fmt.Errorf("Error: unable to move window '%+v' to scratchpad\n", window)
				}

				aerospaceClient.SetLayout(
					window.WindowID,
					"floating",
				)

				movedCount++
			}

			if movedCount == 0 {
				fmt.Println("No windows matched the pattern")
			}

			return nil
		},
	}

	return moveCmd
}
