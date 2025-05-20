/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"

	"github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
)

func main() {
	aerospaceMarkClient, err := aerospacecli.NewAeroSpaceConnection()
	if err != nil {
		fmt.Println("Error creating Aerospace client:", err)
	}

	cmd.Execute(aerospaceMarkClient)
}
