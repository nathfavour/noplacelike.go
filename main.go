package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mycli",
		Short: "MyCLI is a simple CLI application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello from MyCLI!")
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of MyCLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("MyCLI v0.1.0")
		},
	}

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
