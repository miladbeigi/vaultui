package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/miladbeigi/vaultui/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of vaultui",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
