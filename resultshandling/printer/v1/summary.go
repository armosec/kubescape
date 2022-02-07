package v1

import (
	"fmt"

	"github.com/armosec/k8s-interface/workloadinterface"
	"github.com/armosec/opa-utils/reporthandling"
)

type Summary map[string]ResultSummary

func NewSummary() Summary {
	return make(map[string]ResultSummary)
}

type ResultSummary struct {
	ID                string
	RiskScore         float32
	TotalResources    int
	TotalFailed       int
	TotalWarning      int
	Description       string
	Remediation       string
	Framework         []string
	ListInputKinds    []string
	FailedWorkloads   map[string][]WorkloadSummary // <namespace>:[<WorkloadSummary>]
	ExcludedWorkloads map[string][]WorkloadSummary // <namespace>:[<WorkloadSummary>]
	PassedWorkloads   map[string][]WorkloadSummary // <namespace>:[<WorkloadSummary>]
}

type WorkloadSummary struct {
	resource workloadinterface.IMetadata
	status   string
}

func (controlSummary *ResultSummary) ToSlice() []string {
	s := []string{}
	s = append(s, fmt.Sprintf("%d", controlSummary.TotalFailed))
	s = append(s, fmt.Sprintf("%d", controlSummary.TotalWarning))
	s = append(s, fmt.Sprintf("%d", controlSummary.TotalResources))
	return s
}

func workloadSummaryFailed(workloadSummary *WorkloadSummary) bool {
	return workloadSummary.status == reporthandling.StatusFailed
}

func workloadSummaryExclude(workloadSummary *WorkloadSummary) bool {
	return workloadSummary.status == reporthandling.StatusWarning
}

func workloadSummaryPassed(workloadSummary *WorkloadSummary) bool {
	return workloadSummary.status == reporthandling.StatusPassed
}
