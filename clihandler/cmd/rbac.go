package cmd

import (
	"github.com/armosec/k8s-interface/k8sinterface"
	"github.com/armosec/kubescape/cautils"
	"github.com/armosec/kubescape/cautils/getter"
	"github.com/armosec/kubescape/cautils/logger"
	"github.com/armosec/kubescape/cautils/logger/helpers"
	"github.com/armosec/kubescape/clihandler"
	"github.com/armosec/kubescape/clihandler/cliinterfaces"
	reporterv1 "github.com/armosec/kubescape/resultshandling/reporter/v1"
	"github.com/armosec/rbac-utils/rbacscanner"
	"github.com/spf13/cobra"
)

// rabcCmd represents the RBAC command
var rabcCmd = &cobra.Command{
	Use:   "rbac \nExample:\n$ kubescape submit rbac",
	Short: "Submit cluster's Role-Based Access Control(RBAC)",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {

		k8s := k8sinterface.NewKubernetesApi()

		// get config
		clusterConfig := getTenantConfig(submitInfo.Account, "", k8s)
		if err := clusterConfig.SetTenant(); err != nil {
			logger.L().Error("failed setting account ID", helpers.Error(err))
		}

		// list RBAC
		rbacObjects := cautils.NewRBACObjects(rbacscanner.NewRbacScannerFromK8sAPI(k8s, clusterConfig.GetAccountID(), clusterConfig.GetClusterName()))

		// submit resources
		r := reporterv1.NewReportEventReceiver(clusterConfig.GetConfigObj())

		submitInterfaces := cliinterfaces.SubmitInterfaces{
			ClusterConfig: clusterConfig,
			SubmitObjects: rbacObjects,
			Reporter:      r,
		}

		if err := clihandler.Submit(submitInterfaces); err != nil {
			logger.L().Fatal(err.Error())
		}
		return nil
	},
}

func init() {
	submitCmd.AddCommand(rabcCmd)
}

// getKubernetesApi
func getKubernetesApi() *k8sinterface.KubernetesApi {
	if !k8sinterface.IsConnectedToCluster() {
		return nil
	}
	return k8sinterface.NewKubernetesApi()
}
func getTenantConfig(Account, clusterName string, k8s *k8sinterface.KubernetesApi) cautils.ITenantConfig {
	if !k8sinterface.IsConnectedToCluster() || k8s == nil {
		return cautils.NewLocalConfig(getter.GetArmoAPIConnector(), Account, clusterName)
	}
	return cautils.NewClusterConfig(k8s, getter.GetArmoAPIConnector(), Account, clusterName)
}
