package cmd

import (
	"fmt"

	"github.com/miladbeigi/vaultui/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of vaultui",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Banner())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
