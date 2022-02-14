package cmd

import (
	"fmt"
	"os"

	"github.com/armosec/kubescape/cautils"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get current version",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		v := cautils.NewIVersionCheckHandler()
		v.CheckLatestVersion(cautils.NewVersionCheckRequest(cautils.BuildNumber, "", "", "version"))
		fmt.Fprintln(os.Stdout, "Your current version is: "+cautils.BuildNumber)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
