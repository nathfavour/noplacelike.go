package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "noplacelike",
		Short: "NoPlaceLike is a simple CLI application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello from NoPlaceLike!")
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of NoPlaceLike",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("NoPlaceLike v0.1.0")
		},
	}

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
