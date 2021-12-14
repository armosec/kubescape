package opaprocessor

import (
	"testing"

	"github.com/armosec/kubescape/cautils"
	"github.com/armosec/opa-utils/objectsenvelopes"
	"github.com/armosec/opa-utils/reporthandling"
	"github.com/armosec/opa-utils/resources"

	"github.com/armosec/k8s-interface/k8sinterface"
	"github.com/armosec/k8s-interface/workloadinterface"
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
)

func NewOPAProcessorMock() *OPAProcessor {
	return &OPAProcessor{}
}
func TestProcess(t *testing.T) {

	// set k8s
	k8sResources := make(cautils.K8SResources)
	allResources := make(map[string]workloadinterface.IMetadata)
	imetaObj := objectsenvelopes.ListMapToMeta(k8sinterface.ConvertUnstructuredSliceToMap(k8sinterface.V1KubeSystemNamespaceMock().Items))
	for i := range imetaObj {
		allResources[imetaObj[i].GetID()] = imetaObj[i]
	}
	k8sResources["/v1/pods"] = workloadinterface.ListMetaIDs(imetaObj)

	// set opaSessionObj
	opaSessionObj := cautils.NewOPASessionObjMock()
	opaSessionObj.Frameworks = []reporthandling.Framework{*reporthandling.MockFrameworkA()}
	opaSessionObj.K8SResources = &k8sResources
	opaSessionObj.AllResources = allResources

	opap := NewOPAProcessor(opaSessionObj, resources.NewRegoDependenciesDataMock())
	opap.Process()
	opap.updateResults()
	for _, f := range opap.PostureReport.FrameworkReports {
		for _, c := range f.ControlReports {
			for _, r := range c.RuleReports {
				for _, rr := range r.RuleResponses {
					// t.Errorf("AlertMessage: %v", rr.AlertMessage)
					if rr.Exception != nil {
						t.Errorf("Exception: %v", rr.Exception)
					}
				}
			}
		}
	}

}
