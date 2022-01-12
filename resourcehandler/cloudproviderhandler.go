package resourcehandler

import (
	"os"

	"github.com/armosec/k8s-interface/cloudsupport"
	"github.com/armosec/k8s-interface/k8sinterface"
)

var (
	KS_KUBE_CLUSTER_ENV_VAR   = "KS_KUBE_CLUSTER"
	KS_CLOUD_PROVIDER_ENV_VAR = "KS_CLOUD_PROVIDER"
	KS_CLOUD_REGION_ENV_VAR   = "KS_CLOUD_REGION"
	KS_GKE_PROJECT_ENV_VAR    = "KS_GKE_PROJECT"
)

type ICloudProvider interface {
	getKubeCluster() string
	getRegion(cluster string, provider string) (string, error)
	getProject(cluster string, provider string) (string, error)
	getKubeClusterName() string
}

func initCloudProvider() ICloudProvider {
	var provider string
	if isEnvVars() {
		provider = getCloudProviderFromEnvVar()
		switch provider {
		case "gke":
			return NewGKEProviderEnvVar()
		case "eks":
			return NewEKSProviderEnvVar()
		}
	} else {
		provider = getCloudProviderFromContext()
		switch provider {
		case "gke":
			return NewGKEProviderContext()
		case "eks":
			return NewEKSProviderContext()
		}
	}
	return NewEmptyCloudProvider()
}

func getCloudProvider() string {
	if isEnvVars() {
		return getCloudProviderFromEnvVar()
	}
	return getCloudProviderFromContext()
}

func getCloudProviderFromContext() string {
	return cloudsupport.GetCloudProvider(getClusterFromContext())
}

func getClusterFromContext() string {
	cluster := k8sinterface.GetCurrentContext().Cluster
	if cluster != "" {
		return cluster
	}
	return k8sinterface.GetClusterName()
}

func getCloudProviderFromEnvVar() string {
	val, present := os.LookupEnv(KS_CLOUD_PROVIDER_ENV_VAR)
	if present {
		return val
	}
	return ""
}

func isEnvVars() bool {
	_, present := os.LookupEnv(KS_KUBE_CLUSTER_ENV_VAR)
	if !present {
		return false
	}
	_, present = os.LookupEnv(KS_CLOUD_PROVIDER_ENV_VAR)
	if !present {
		return false
	}
	_, present = os.LookupEnv(KS_CLOUD_REGION_ENV_VAR)
	return present
}

type EmptyCloudProvider struct {
}

func NewEmptyCloudProvider() *EmptyCloudProvider {
	return &EmptyCloudProvider{}
}

func (emptyCloudProvider *EmptyCloudProvider) getKubeCluster() string {
	return getClusterFromContext()
}

func (emptyCloudProvider *EmptyCloudProvider) getKubeClusterName() string {
	return emptyCloudProvider.getKubeCluster()
}

func (emptyCloudProvider *EmptyCloudProvider) getRegion(cluster string, provider string) (string, error) {
	return "", nil
}

func (emptyCloudProvider *EmptyCloudProvider) getProject(cluster string, provider string) (string, error) {
	return "", nil
}
