package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of NoPlaceLike",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("NoPlaceLike v%s\n", version)
		},
	}
}
