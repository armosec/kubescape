package getter

import (
	"strings"

	"github.com/armosec/opa-utils/gitregostore"
	"github.com/armosec/opa-utils/reporthandling"
)

// =======================================================================================================================
// ======================================== DownloadReleasedPolicy =======================================================
// =======================================================================================================================

// Use gitregostore to get policies from github release
type DownloadReleasedPolicy struct {
	gs *gitregostore.GitRegoStore
}

func NewDownloadReleasedPolicy() *DownloadReleasedPolicy {
	return &DownloadReleasedPolicy{
		gs: gitregostore.NewDefaultGitRegoStore(-1),
	}
}

func (drp *DownloadReleasedPolicy) GetControl(policyName string) (*reporthandling.Control, error) {
	var control *reporthandling.Control
	var err error

	control, err = drp.gs.GetOPAControl(policyName)
	if err != nil {
		return nil, err
	}
	return control, nil
}

func (drp *DownloadReleasedPolicy) GetFramework(name string) (*reporthandling.Framework, error) {
	framework, err := drp.gs.GetOPAFrameworkByName(name)
	if err != nil {
		return nil, err
	}
	return framework, err
}

func (drp *DownloadReleasedPolicy) SetRegoObjects() error {
	return drp.gs.SetRegoObjects()
}

func isNativeFramework(framework string) bool {
	return contains(NativeFrameworks, framework)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.EqualFold(v, str) {
			return true
		}
	}
	return false
}
