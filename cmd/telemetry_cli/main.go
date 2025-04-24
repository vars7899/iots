package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// var (
// 	apiEndpoint string
// )

func main() {
	Execute()
}

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "telpush",
		Short: "A CLI tool to push telemetry data to an IoT API",
		Long:  `A command-line interface for sending simulated telemetry data to test an IoT backend.`,
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
