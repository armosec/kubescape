package cmd

import (
	"github.com/armosec/k8s-interface/k8sinterface"
	"github.com/armosec/kubescape/cautils"
	"github.com/spf13/cobra"
)

var scanInfo cautils.ScanInfo

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [command]",
	Short: "Scan the current running cluster or yaml files",
	Long:  `The action you want to perform`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if args[0] != "framework" && args[0] != "control" {
				scanInfo.ScanAll = true
				return frameworkCmd.RunE(cmd, append([]string{"all"}, args...))
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			scanInfo.ScanAll = true
			return frameworkCmd.RunE(cmd, []string{"all"})
		}
		return nil
	},
}

func frameworkInitConfig() {
	k8sinterface.SetClusterContextName(scanInfo.KubeContext)
}

func init() {

	cobra.OnInitialize(frameworkInitConfig)

	rootCmd.AddCommand(scanCmd)

	scanCmd.PersistentFlags().StringVarP(&scanInfo.Account, "account", "", "", "Armo portal account ID. Default will load account ID from configMap or config file")
	scanCmd.PersistentFlags().StringVarP(&scanInfo.KubeContext, "kube-context", "", "", "Kube context. Default will use the current-context")
	scanCmd.PersistentFlags().StringVar(&scanInfo.ControlsInputs, "controls-config", "", "Path to an controls-config obj. If not set will download controls-config from ARMO management portal")
	scanCmd.PersistentFlags().StringVar(&scanInfo.UseExceptions, "exceptions", "", "Path to an exceptions obj. If not set will download exceptions from ARMO management portal")
	scanCmd.PersistentFlags().StringVar(&scanInfo.UseArtifactsFrom, "use-artifacts-from", "", "Load artifacts from local directory. If not used will download them")
	scanCmd.PersistentFlags().StringVarP(&scanInfo.ExcludedNamespaces, "exclude-namespaces", "e", "", "Namespaces to exclude from scanning. Recommended: kube-system,kube-public")
	scanCmd.PersistentFlags().Uint16VarP(&scanInfo.FailThreshold, "fail-threshold", "t", 100, "Failure threshold is the percent above which the command fails and returns exit code 1")
	scanCmd.PersistentFlags().StringVarP(&scanInfo.Format, "format", "f", "pretty-printer", `Output format. Supported formats: "pretty-printer","json","junit","prometheus","pdf"`)
	scanCmd.PersistentFlags().StringVar(&scanInfo.IncludeNamespaces, "include-namespaces", "", "scan specific namespaces. e.g: --include-namespaces ns-a,ns-b")
	scanCmd.PersistentFlags().BoolVarP(&scanInfo.Local, "keep-local", "", false, "If you do not want your Kubescape results reported to Armo backend. Use this flag if you ran with the '--submit' flag in the past and you do not want to submit your current scan results")
	scanCmd.PersistentFlags().StringVarP(&scanInfo.Output, "output", "o", "", "Output file. Print output to file and not stdout")
	scanCmd.PersistentFlags().BoolVar(&scanInfo.VerboseMode, "verbose", false, "Display all of the input resources and not only failed resources")
	scanCmd.PersistentFlags().BoolVar(&scanInfo.UseDefault, "use-default", false, "Load local policy object from default path. If not used will download latest")
	scanCmd.PersistentFlags().StringSliceVar(&scanInfo.UseFrom, "use-from", nil, "Load local policy object from specified path. If not used will download latest")
	scanCmd.PersistentFlags().BoolVarP(&scanInfo.Silent, "silent", "s", false, "Silent progress messages")
	scanCmd.PersistentFlags().BoolVarP(&scanInfo.Submit, "submit", "", false, "Send the scan results to Armo management portal where you can see the results in a user-friendly UI, choose your preferred compliance framework, check risk results history and trends, manage exceptions, get remediation recommendations and much more. By default the results are not submitted")

	hostF := scanCmd.PersistentFlags().VarPF(&scanInfo.HostSensor, "enable-host-scan", "", "Deploy ARMO K8s host-sensor daemonset in the scanned cluster. Deleting it right after we collecting the data. Required to collect valueable data from cluster nodes for certain controls")
	hostF.NoOptDefVal = "true"
	hostF.DefValue = "false, for no TTY in stdin"

}
