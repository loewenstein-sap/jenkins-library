package blackduck

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/pkg/errors"
)

// CreateSarifResultFile creates a SARIF result from the Vulnerabilities that were brought up by the scan
func CreateSarifResultFile(vulns *Vulnerabilities, components *Components) *format.SARIF {
	log.Entry().Debug("Creating SARIF file for data transfer")
	var sarif format.SARIF
	sarif.Schema = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos02/schemas/sarif-schema-2.1.0.json"
	sarif.Version = "2.1.0"
	var wsRun format.Runs

	//handle the tool object
	tool := format.Tool{}
	tool.Driver = format.Driver{}
	tool.Driver.Name = "Blackduck Hub Detect"
	tool.Driver.Version = "unknown"
	tool.Driver.InformationUri = "https://community.synopsys.com/s/document-item?bundleId=integrations-detect&topicId=introduction.html&_LANG=enus"

	// Handle results/vulnerabilities
	collectedRules := []string{}
	cweIdsForTaxonomies := []string{}
	if vulns != nil && vulns.Items != nil {
		for _, v := range vulns.Items {
			result := format.Results{}
			result.RuleID = v.VulnerabilityWithRemediation.VulnerabilityName
			log.Entry().Debugf("Transforming alert %v into SARIF format", result.RuleID)
			result.Level = transformToLevel(v.VulnerabilityWithRemediation.Severity)
			result.Message = &format.Message{}
			result.Message.Text = v.VulnerabilityWithRemediation.Description
			result.AnalysisTarget = &format.ArtifactLocation{}
			result.AnalysisTarget.URI = v.Component.ToPackageUrl().ToString()
			result.AnalysisTarget.Index = 0
			location := format.Location{PhysicalLocation: format.PhysicalLocation{ArtifactLocation: format.ArtifactLocation{URI: v.Name}}}
			result.Locations = append(result.Locations, location)
			partialFingerprints := format.PartialFingerprints{}
			partialFingerprints.PackageURLPlusCVEHash = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%v+%v", v.Component.ToPackageUrl().ToString(), v.CweID)))
			result.PartialFingerprints = partialFingerprints
			cweIdsForTaxonomies = append(cweIdsForTaxonomies, v.VulnerabilityWithRemediation.CweID)

			// append the result
			wsRun.Results = append(wsRun.Results, result)

			// only create rule on new CVE
			if !piperutils.ContainsString(collectedRules, result.RuleID) {
				collectedRules = append(collectedRules, result.RuleID)

				sarifRule := format.SarifRule{}
				sarifRule.ID = result.RuleID
				sarifRule.ShortDescription = &format.Message{}
				sarifRule.ShortDescription.Text = fmt.Sprintf("%v in Package %v", v.VulnerabilityName, v.Component.Name)
				sarifRule.FullDescription = &format.Message{}
				sarifRule.FullDescription.Text = v.VulnerabilityWithRemediation.Description
				sarifRule.DefaultConfiguration = &format.DefaultConfiguration{}
				sarifRule.DefaultConfiguration.Level = transformToLevel(v.VulnerabilityWithRemediation.Severity)
				sarifRule.HelpURI = ""
				markdown, _ := v.ToMarkdown()
				sarifRule.Help = &format.Help{}
				sarifRule.Help.Text = v.ToTxt()
				sarifRule.Help.Markdown = string(markdown)

				ruleProp := format.SarifRuleProperties{}
				ruleProp.Tags = append(ruleProp.Tags, "SECURITY_VULNERABILITY")
				ruleProp.Tags = append(ruleProp.Tags, v.Component.ToPackageUrl().ToString())
				ruleProp.Tags = append(ruleProp.Tags, v.VulnerabilityWithRemediation.CweID)
				ruleProp.Precision = "very-high"
				ruleProp.Impact = fmt.Sprint(v.VulnerabilityWithRemediation.ImpactSubscore)
				ruleProp.Probability = fmt.Sprint(v.VulnerabilityWithRemediation.ExploitabilitySubscore)
				ruleProp.SecuritySeverity = fmt.Sprint(v.OverallScore)
				sarifRule.Properties = &ruleProp

				// append the rule
				tool.Driver.Rules = append(tool.Driver.Rules, sarifRule)
			}
		}
	}
	//Finalize: tool
	wsRun.Tool = tool

	// Threadflowlocations is no loger useful: voiding it will make for smaller reports
	wsRun.ThreadFlowLocations = []format.Locations{}

	// Add a conversion object to highlight this isn't native SARIF
	conversion := &format.Conversion{}
	conversion.Tool.Driver.Name = "Piper FPR to SARIF converter"
	conversion.Tool.Driver.InformationUri = "https://github.com/SAP/jenkins-library"
	conversion.Invocation.ExecutionSuccessful = true
	convInvocProp := &format.InvocationProperties{}
	convInvocProp.Platform = runtime.GOOS
	conversion.Invocation.Properties = convInvocProp
	wsRun.Conversion = conversion

	//handle taxonomies
	//Only one exists apparently: CWE. It is fixed
	taxonomy := format.Taxonomies{}
	taxonomy.GUID = "25F72D7E-8A92-459D-AD67-64853F788765"
	taxonomy.Name = "CWE"
	taxonomy.Organization = "MITRE"
	taxonomy.ShortDescription.Text = "The MITRE Common Weakness Enumeration"
	for key := range cweIdsForTaxonomies {
		taxa := format.Taxa{}
		taxa.Id = fmt.Sprint(key)
		taxonomy.Taxa = append(taxonomy.Taxa, taxa)
	}
	wsRun.Taxonomies = append(wsRun.Taxonomies, taxonomy)
	sarif.Runs = append(sarif.Runs, wsRun)

	return &sarif
}

func transformToLevel(severity string) string {
	switch severity {
	case "LOW":
		return "warning"
	case "MEDIUM":
		return "warning"
	case "HIGH":
		return "error"
	case "CRITICAL":
		return "error"
	}
	return "none"
}

// WriteVulnerabilityReports writes vulnerability information from ScanReport into dedicated outputs e.g. HTML
func WriteVulnerabilityReports(scanReport reporting.ScanReport, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := "piper_detect_vulnerability_report.html"
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "BlackDuck Vulnerability Report", Target: htmlReportPath})

	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("detectExecuteScan_oss_%v.json", fmt.Sprintf("%v", utils.CurrentTime("")))), jsonReport, 0666); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}

	return reportPaths, nil
}

// WriteSarifFile write a JSON sarif format file for upload into e.g. GCP
func WriteSarifFile(sarif *format.SARIF, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	sarifReport, errorMarshall := json.Marshal(sarif)
	if errorMarshall != nil {
		return reportPaths, errors.Wrapf(errorMarshall, "failed to marshall SARIF json file")
	}
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}
	sarifReportPath := filepath.Join(ReportsDirectory, "piper_detect_vulnerability.sarif")
	if err := utils.FileWrite(sarifReportPath, sarifReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write SARIF file")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Blackduck Detect Vulnerability SARIF file", Target: sarifReportPath})

	return reportPaths, nil
}
