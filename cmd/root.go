/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
func RootCmd(
	aerospaceMarkClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "aerospace-scratchpad",
		Short: "Scratchpad for AeroSpace WM",
		Long: `Scratchpad for AeroSpace WM

Allows you manage your windows in a scratchpad-like manner.
This is heavily inspired by i3wm's scratchpad feature, but follows aerospace command line conventions.

See:
https://i3wm.org/docs/userguide.html#_scratchpad
`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}

	// Commands
	rootCmd.AddCommand(MoveCmd(aerospaceMarkClient))
	rootCmd.AddCommand(ShowCmd(aerospaceMarkClient))

	// Flags
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(
	aerospaceMarkClient aerospacecli.AeroSpaceClient,
) {
	rootCmd := RootCmd(aerospaceMarkClient)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

