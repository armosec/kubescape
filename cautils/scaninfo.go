package cautils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/armosec/kubescape/cautils/getter"
	"github.com/armosec/kubescape/cautils/logger"
	"github.com/armosec/kubescape/cautils/logger/helpers"
	"github.com/armosec/opa-utils/reporthandling"
)

const (
	ScanCluster                string = "cluster"
	ScanLocalFiles             string = "yaml"
	localControlInputsFilename string = "controls-inputs.json"
	localExceptionsFilename    string = "exceptions.json"
)

type BoolPtrFlag struct {
	valPtr *bool
}

func (bpf *BoolPtrFlag) Type() string {
	return "bool"
}

func (bpf *BoolPtrFlag) String() string {
	if bpf.valPtr != nil {
		return fmt.Sprintf("%v", *bpf.valPtr)
	}
	return ""
}
func (bpf *BoolPtrFlag) Get() *bool {
	return bpf.valPtr
}

func (bpf *BoolPtrFlag) SetBool(val bool) {
	bpf.valPtr = &val
}

func (bpf *BoolPtrFlag) Set(val string) error {
	switch val {
	case "true":
		bpf.SetBool(true)
	case "false":
		bpf.SetBool(false)
	}
	return nil
}

type RootInfo struct {
	Logger   string // logger level
	CacheDir string // cached dir
}
type ScanInfo struct {
	Getters
	PolicyIdentifier   []reporthandling.PolicyIdentifier
	UseExceptions      string      // Load file with exceptions configuration
	ControlsInputs     string      // Load file with inputs for controls
	UseFrom            []string    // Load framework from local file (instead of download). Use when running offline
	UseDefault         bool        // Load framework from cached file (instead of download). Use when running offline
	UseArtifactsFrom   string      // Load artifacts from local path. Use when running offline
	VerboseMode        bool        // Display all of the input resources and not only failed resources
	Format             string      // Format results (table, json, junit ...)
	Output             string      // Store results in an output file, Output file name
	OutputVersion      string      // Output object can be differnet between versions, this is for testing and backward compatibility
	ExcludedNamespaces string      // used for host sensor namespace
	IncludeNamespaces  string      // DEPRECATED?
	InputPatterns      []string    // Yaml files input patterns
	Silent             bool        // Silent mode - Do not print progress logs
	FailThreshold      float32     // Failure score threshold
	Submit             bool        // Submit results to Armo BE
	HostSensorEnabled  BoolPtrFlag // Deploy ARMO K8s host sensor to collect data from certain controls
	HostSensorYamlPath string      // Path to hostsensor file
	Local              bool        // Do not submit results
	Account            string      // account ID
	KubeContext        string      // context name
	FrameworkScan      bool        // false if scanning control
	ScanAll            bool        // true if scan all frameworks
}

type Getters struct {
	ExceptionsGetter     getter.IExceptionsGetter
	ControlsInputsGetter getter.IControlsInputsGetter
	PolicyGetter         getter.IPolicyGetter
}

func (scanInfo *ScanInfo) Init() {
	scanInfo.setUseFrom()
	scanInfo.setOutputFile()
	scanInfo.setUseArtifactsFrom()
}

func (scanInfo *ScanInfo) setUseArtifactsFrom() {
	if scanInfo.UseArtifactsFrom == "" {
		return
	}
	// UseArtifactsFrom must be a path without a filename
	dir, file := filepath.Split(scanInfo.UseArtifactsFrom)
	if dir == "" {
		scanInfo.UseArtifactsFrom = file
	} else if strings.Contains(file, ".json") {
		scanInfo.UseArtifactsFrom = dir
	}
	// set frameworks files
	files, err := ioutil.ReadDir(scanInfo.UseArtifactsFrom)
	if err != nil {
		logger.L().Fatal("failed to read files from directory", helpers.String("dir", scanInfo.UseArtifactsFrom), helpers.Error(err))
	}
	framework := &reporthandling.Framework{}
	for _, f := range files {
		filePath := filepath.Join(scanInfo.UseArtifactsFrom, f.Name())
		file, err := os.ReadFile(filePath)
		if err == nil {
			if err := json.Unmarshal(file, framework); err == nil {
				scanInfo.UseFrom = append(scanInfo.UseFrom, filepath.Join(scanInfo.UseArtifactsFrom, f.Name()))
			}
		}
	}
	// set config-inputs file
	scanInfo.ControlsInputs = filepath.Join(scanInfo.UseArtifactsFrom, localControlInputsFilename)
	// set exceptions
	scanInfo.UseExceptions = filepath.Join(scanInfo.UseArtifactsFrom, localExceptionsFilename)
}

func (scanInfo *ScanInfo) setUseExceptions() {
	if scanInfo.UseExceptions != "" {
		// load exceptions from file
		scanInfo.ExceptionsGetter = getter.NewLoadPolicy([]string{scanInfo.UseExceptions})
	} else {
		scanInfo.ExceptionsGetter = getter.GetArmoAPIConnector()
	}
}

func (scanInfo *ScanInfo) setUseFrom() {
	if scanInfo.UseDefault {
		for _, policy := range scanInfo.PolicyIdentifier {
			scanInfo.UseFrom = append(scanInfo.UseFrom, getter.GetDefaultPath(policy.Name+".json"))
		}
	}
}

func (scanInfo *ScanInfo) setOutputFile() {
	if scanInfo.Output == "" {
		return
	}
	if scanInfo.Format == "json" {
		if filepath.Ext(scanInfo.Output) != ".json" {
			scanInfo.Output += ".json"
		}
	}
	if scanInfo.Format == "junit" {
		if filepath.Ext(scanInfo.Output) != ".xml" {
			scanInfo.Output += ".xml"
		}
	}
	if scanInfo.Format == "pdf" {
		if filepath.Ext(scanInfo.Output) != ".pdf" {
			scanInfo.Output += ".pdf"
		}
	}
}

func (scanInfo *ScanInfo) GetScanningEnvironment() string {
	if len(scanInfo.InputPatterns) != 0 {
		return ScanLocalFiles
	}
	return ScanCluster
}

func (scanInfo *ScanInfo) SetPolicyIdentifiers(policies []string, kind reporthandling.NotificationPolicyKind) {
	for _, policy := range policies {
		if !scanInfo.contains(policy) {
			newPolicy := reporthandling.PolicyIdentifier{}
			newPolicy.Kind = kind // reporthandling.KindFramework
			newPolicy.Name = policy
			scanInfo.PolicyIdentifier = append(scanInfo.PolicyIdentifier, newPolicy)
		}
	}
}

func (scanInfo *ScanInfo) contains(policyName string) bool {
	for _, policy := range scanInfo.PolicyIdentifier {
		if policy.Name == policyName {
			return true
		}
	}
	return false
}
