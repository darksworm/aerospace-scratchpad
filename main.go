/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package main

import (
	"fmt"
	"log"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/ilmars/aerospace-sticky/cmd"
	"github.com/ilmars/aerospace-sticky/internal/logger"
)

func main() {
	defaultLogger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Error: creating logger\n%v", err)
		return
	}
	defer func() {
		if err := defaultLogger.Close(); err != nil {
			log.Fatalf("Error: closing logger\n%v", err)
		}
	}()
	logger.SetDefaultLogger(defaultLogger)
	defaultLogger.LogInfo("Executing Aerospace Sticky CLI")

	aerospaceMarkClient, err := aerospacecli.NewAeroSpaceClient()
	if err != nil {
		fmt.Println("Error creating Aerospace client:", err)
	}

	cmd.Execute(aerospaceMarkClient)
}
