package printer

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/armosec/k8s-interface/workloadinterface"
	"github.com/armosec/kubescape/cautils"
	"github.com/armosec/opa-utils/objectsenvelopes"
	"github.com/armosec/opa-utils/reporthandling"
	"github.com/enescakir/emoji"
	"github.com/olekukonko/tablewriter"
)

type PrettyPrinter struct {
	writer             *os.File
	summary            Summary
	verboseMode        bool
	sortedControlNames []string
	frameworkSummary   ResultSummary
}

func NewPrettyPrinter(verboseMode bool) *PrettyPrinter {
	return &PrettyPrinter{
		verboseMode: verboseMode,
		summary:     NewSummary(),
	}
}

func (printer *PrettyPrinter) ActionPrint(opaSessionObj *cautils.OPASessionObj) {
	// score := calculatePostureScore(opaSessionObj.PostureReport)
	failedResources := []string{}
	warningResources := []string{}
	allResources := []string{}
	frameworkNames := []string{}
	frameworkScores := []float32{}

	var overallRiskScore float32 = 0
	for _, frameworkReport := range opaSessionObj.PostureReport.FrameworkReports {
		frameworkNames = append(frameworkNames, frameworkReport.Name)
		frameworkScores = append(frameworkScores, frameworkReport.Score)
		failedResources = reporthandling.GetUniqueResourcesIDs(append(failedResources, frameworkReport.ListResourcesIDs().GetFailedResources()...))
		warningResources = reporthandling.GetUniqueResourcesIDs(append(warningResources, frameworkReport.ListResourcesIDs().GetWarningResources()...))
		allResources = reporthandling.GetUniqueResourcesIDs(append(allResources, frameworkReport.ListResourcesIDs().GetAllResources()...))
		printer.summarySetup(frameworkReport, opaSessionObj.AllResources)
		overallRiskScore += frameworkReport.Score
	}

	overallRiskScore /= float32(len(opaSessionObj.PostureReport.FrameworkReports))

	printer.frameworkSummary = ResultSummary{
		RiskScore:      overallRiskScore,
		TotalResources: len(allResources),
		TotalFailed:    len(failedResources),
		TotalWarning:   len(warningResources),
	}

	printer.printResults()
	printer.printSummaryTable(frameworkNames, frameworkScores)

}

func (printer *PrettyPrinter) SetWriter(outputFile string) {
	printer.writer = getWriter(outputFile)
}

func (printer *PrettyPrinter) Score(score float32) {
}

func (printer *PrettyPrinter) summarySetup(fr reporthandling.FrameworkReport, allResources map[string]workloadinterface.IMetadata) {

	for _, cr := range fr.ControlReports {
		if len(cr.RuleReports) == 0 {
			continue
		}
		workloadsSummary := listResultSummary(cr.RuleReports, allResources)

		var passedWorkloads map[string][]WorkloadSummary
		if printer.verboseMode {
			passedWorkloads = groupByNamespaceOrKind(workloadsSummary, workloadSummaryPassed)
		}

		//controlSummary
		printer.summary[cr.Name] = ResultSummary{
			ID:                cr.ControlID,
			RiskScore:         cr.Score,
			TotalResources:    cr.GetNumberOfResources(),
			TotalFailed:       cr.GetNumberOfFailedResources(),
			TotalWarning:      cr.GetNumberOfWarningResources(),
			FailedWorkloads:   groupByNamespaceOrKind(workloadsSummary, workloadSummaryFailed),
			ExcludedWorkloads: groupByNamespaceOrKind(workloadsSummary, workloadSummaryExclude),
			PassedWorkloads:   passedWorkloads,
			Description:       cr.Description,
			Remediation:       cr.Remediation,
			ListInputKinds:    cr.ListControlsInputKinds(),
		}

	}
	printer.sortedControlNames = printer.getSortedControlsNames()
}
func (printer *PrettyPrinter) printResults() {
	for i := 0; i < len(printer.sortedControlNames); i++ {
		controlSummary := printer.summary[printer.sortedControlNames[i]]
		printer.printTitle(printer.sortedControlNames[i], &controlSummary)
		printer.printResources(&controlSummary)
		if printer.summary[printer.sortedControlNames[i]].TotalResources > 0 {
			printer.printSummary(printer.sortedControlNames[i], &controlSummary)
		}

	}
}

func (printer *PrettyPrinter) printSummary(controlName string, controlSummary *ResultSummary) {
	cautils.SimpleDisplay(printer.writer, "Summary - ")
	cautils.SuccessDisplay(printer.writer, "Passed:%v   ", controlSummary.TotalResources-controlSummary.TotalFailed-controlSummary.TotalWarning)
	cautils.WarningDisplay(printer.writer, "Excluded:%v   ", controlSummary.TotalWarning)
	cautils.FailureDisplay(printer.writer, "Failed:%v   ", controlSummary.TotalFailed)
	cautils.InfoDisplay(printer.writer, "Total:%v\n", controlSummary.TotalResources)
	if controlSummary.TotalFailed > 0 {
		cautils.DescriptionDisplay(printer.writer, "Remediation: %v\n", controlSummary.Remediation)
	}
	cautils.DescriptionDisplay(printer.writer, "\n")

}
func (printer *PrettyPrinter) printTitle(controlName string, controlSummary *ResultSummary) {
	cautils.InfoDisplay(printer.writer, "[control: %s - %s] ", controlName, getControlURL(controlSummary.ID))
	if controlSummary.TotalResources == 0 {
		cautils.InfoDisplay(printer.writer, "skipped %v\n", emoji.ConfusedFace)
	} else if controlSummary.TotalFailed != 0 {
		cautils.FailureDisplay(printer.writer, "failed %v\n", emoji.SadButRelievedFace)
	} else if controlSummary.TotalWarning != 0 {
		cautils.WarningDisplay(printer.writer, "excluded %v\n", emoji.NeutralFace)
	} else {
		cautils.SuccessDisplay(printer.writer, "passed %v\n", emoji.ThumbsUp)
	}

	cautils.DescriptionDisplay(printer.writer, "Description: %s\n", controlSummary.Description)

}
func (printer *PrettyPrinter) printResources(controlSummary *ResultSummary) {

	if len(controlSummary.FailedWorkloads) > 0 {
		cautils.FailureDisplay(printer.writer, "Failed:\n")
		printer.printGroupedResources(controlSummary.FailedWorkloads)
	}
	if len(controlSummary.ExcludedWorkloads) > 0 {
		cautils.WarningDisplay(printer.writer, "Excluded:\n")
		printer.printGroupedResources(controlSummary.ExcludedWorkloads)
	}
	if len(controlSummary.PassedWorkloads) > 0 {
		cautils.SuccessDisplay(printer.writer, "Passed:\n")
		printer.printGroupedResources(controlSummary.PassedWorkloads)
	}

}

func (printer *PrettyPrinter) printGroupedResources(workloads map[string][]WorkloadSummary) {
	indent := INDENT
	for title, rsc := range workloads {
		printer.printGroupedResource(indent, title, rsc)
	}
}

func (printer *PrettyPrinter) printGroupedResource(indent string, title string, rsc []WorkloadSummary) {
	preIndent := indent
	if title != "" {
		cautils.SimpleDisplay(printer.writer, "%s%s\n", indent, title)
		indent += indent
	}

	for r := range rsc {
		relatedObjectsStr := generateRelatedObjectsStr(rsc[r])
		cautils.SimpleDisplay(printer.writer, fmt.Sprintf("%s%s - %s %s\n", indent, rsc[r].resource.GetKind(), rsc[r].resource.GetName(), relatedObjectsStr))
	}
	indent = preIndent
}

func generateRelatedObjectsStr(workload WorkloadSummary) string {
	relatedStr := ""
	if workload.resource.GetObjectType() == workloadinterface.TypeWorkloadObject {
		relatedObjects := objectsenvelopes.NewRegoResponseVectorObject(workload.resource.GetObject()).GetRelatedObjects()
		for i, related := range relatedObjects {
			if ns := related.GetNamespace(); i == 0 && ns != "" {
				relatedStr += fmt.Sprintf("Namespace - %s, ", ns)
			}
			relatedStr += fmt.Sprintf("%s - %s, ", related.GetKind(), related.GetName())
		}
	}
	if relatedStr != "" {
		relatedStr = fmt.Sprintf(" [%s]", relatedStr[:len(relatedStr)-2])
	}
	return relatedStr
}

func generateRow(control string, cs ResultSummary) []string {
	row := []string{control}
	row = append(row, cs.ToSlice()...)
	if cs.TotalResources != 0 {
		row = append(row, fmt.Sprintf("%d", int(cs.RiskScore))+"%")
	} else {
		row = append(row, "skipped")
	}
	return row
}

func generateHeader() []string {
	return []string{"Control Name", "Failed Resources", "Excluded Resources", "All Resources", "% risk-score"}
}

func generateFooter(printer *PrettyPrinter) []string {
	// Control name | # failed resources | all resources | % success
	row := []string{}
	row = append(row, "Resource Summary") //fmt.Sprintf(""%d", numControlers"))
	row = append(row, fmt.Sprintf("%d", printer.frameworkSummary.TotalFailed))
	row = append(row, fmt.Sprintf("%d", printer.frameworkSummary.TotalWarning))
	row = append(row, fmt.Sprintf("%d", printer.frameworkSummary.TotalResources))
	row = append(row, fmt.Sprintf("%.2f%s", printer.frameworkSummary.RiskScore, "%"))

	return row
}
func (printer *PrettyPrinter) printSummaryTable(frameworksNames []string, frameworkScores []float32) {
	// For control scan framework will be nil
	printer.printFramework(frameworksNames, frameworkScores)

	summaryTable := tablewriter.NewWriter(printer.writer)
	summaryTable.SetAutoWrapText(false)
	summaryTable.SetHeader(generateHeader())
	summaryTable.SetHeaderLine(true)
	alignments := []int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER}
	summaryTable.SetColumnAlignment(alignments)

	for i := 0; i < len(printer.sortedControlNames); i++ {
		controlSummary := printer.summary[printer.sortedControlNames[i]]
		summaryTable.Append(generateRow(printer.sortedControlNames[i], controlSummary))
	}

	summaryTable.SetFooter(generateFooter(printer))

	// summaryTable.SetFooter(generateFooter())
	summaryTable.Render()
}

func (printer *PrettyPrinter) printFramework(frameworksNames []string, frameworkScores []float32) {
	if len(frameworksNames) == 1 {
		cautils.InfoTextDisplay(printer.writer, fmt.Sprintf("FRAMEWORK %s\n", frameworksNames[0]))
	} else if len(frameworksNames) > 1 {
		p := "FRAMEWORKS: "
		for i := 0; i < len(frameworksNames)-1; i++ {
			p += fmt.Sprintf("%s (risk: %.2f), ", frameworksNames[i], frameworkScores[i])
		}
		p += fmt.Sprintf("%s (risk: %.2f)\n", frameworksNames[len(frameworksNames)-1], frameworkScores[len(frameworkScores)-1])
		cautils.InfoTextDisplay(printer.writer, p)
	}
}

func (printer *PrettyPrinter) getSortedControlsNames() []string {
	controlNames := make([]string, 0, len(printer.summary))
	for k := range printer.summary {
		controlNames = append(controlNames, k)
	}
	sort.Strings(controlNames)
	return controlNames
}

func getWriter(outputFile string) *os.File {
	os.Remove(outputFile)
	if outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("failed to open file for writing, reason: ", err.Error())
			return os.Stdout
		}
		return f
	}
	return os.Stdout

}

func getControlURL(controlID string) string {
	return fmt.Sprintf("https://hub.armo.cloud/docs/%s", strings.ToLower(controlID))
}
