package cmd

import (
	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:        "cluster",
	Short:      "Set configuration for cluster",
	Long:       ``,
	Deprecated: "use the 'set' command instead",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	configCmd.AddCommand(clusterCmd)
}
