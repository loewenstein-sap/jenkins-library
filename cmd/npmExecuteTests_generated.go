// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/bmatcuk/doublestar"
	"github.com/spf13/cobra"
)

type npmExecuteTestsOptions struct {
	InstallCommand   string                   `json:"installCommand,omitempty"`
	RunCommand       string                   `json:"runCommand,omitempty"`
	VaultURLs        []map[string]interface{} `json:"vaultURLs,omitempty"`
	VaultUsername    string                   `json:"vaultUsername,omitempty"`
	VaultPassword    string                   `json:"vaultPassword,omitempty"`
	BaseURL          string                   `json:"baseUrl,omitempty"`
	UsernameEnvVar   string                   `json:"usernameEnvVar,omitempty"`
	PasswordEnvVar   string                   `json:"passwordEnvVar,omitempty"`
	UrlOptionPrefix  string                   `json:"urlOptionPrefix,omitempty"`
	Envs             []string                 `json:"envs,omitempty"`
	Paths            []string                 `json:"paths,omitempty"`
	WorkingDirectory string                   `json:"workingDirectory,omitempty"`
}

type npmExecuteTestsReports struct {
}

func (p *npmExecuteTestsReports) persist(stepConfig npmExecuteTestsOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/e2e-results.xml", ParamRef: "", StepResultType: "end-to-end-test"},
	}
	envVars := []gcs.EnvVar{
		{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: gcpJsonKeyFilePath, Modified: false},
	}
	gcsClient, err := gcs.NewClient(gcs.WithEnvVars(envVars))
	if err != nil {
		log.Entry().Errorf("creation of GCS client failed: %v", err)
		return
	}
	defer gcsClient.Close()
	structVal := reflect.ValueOf(&stepConfig).Elem()
	inputParameters := map[string]string{}
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Type().Field(i)
		if field.Type.String() == "string" {
			paramName := strings.Split(field.Tag.Get("json"), ",")
			paramValue, _ := structVal.Field(i).Interface().(string)
			inputParameters[paramName[0]] = paramValue
		}
	}
	if err := gcs.PersistReportsToGCS(gcsClient, content, inputParameters, gcsFolderPath, gcsBucketId, gcsSubFolder, doublestar.Glob, os.Stat); err != nil {
		log.Entry().Errorf("failed to persist reports: %v", err)
	}
}

// NpmExecuteTestsCommand Executes end-to-end tests using npm
func NpmExecuteTestsCommand() *cobra.Command {
	const STEP_NAME = "npmExecuteTests"

	metadata := npmExecuteTestsMetadata()
	var stepConfig npmExecuteTestsOptions
	var startTime time.Time
	var reports npmExecuteTestsReports
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createNpmExecuteTestsCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes end-to-end tests using npm",
		Long: `This step executes end-to-end tests in a Docker environment using npm.

The step spins up a Docker container based on the specified ` + "`" + `dockerImage` + "`" + ` and executes the ` + "`" + `installScript` + "`" + ` and ` + "`" + `runScript` + "`" + ` from ` + "`" + `package.json` + "`" + `.

The application URLs and credentials can be specified in ` + "`" + `appUrls` + "`" + ` and ` + "`" + `credentialsId` + "`" + ` respectively. If ` + "`" + `wdi5` + "`" + ` is set to ` + "`" + `true` + "`" + `, the step uses ` + "`" + `wdi5_username` + "`" + ` and ` + "`" + `wdi5_password` + "`" + ` for authentication.

The tests can be restricted to run only on the productive branch by setting ` + "`" + `onlyRunInProductiveBranch` + "`" + ` to ` + "`" + `true` + "`" + `.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, err := os.Getwd()
			if err != nil {
				return err
			}
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err = PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 || len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			if err = log.RegisterANSHookIfConfigured(GeneralConfig.CorrelationID); err != nil {
				log.Entry().WithError(err).Warn("failed to set up SAP Alert Notification Service log hook")
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			vaultClient := config.GlobalVaultClient()
			if vaultClient != nil {
				defer vaultClient.MustRevokeToken()
			}

			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				reports.persist(stepConfig, GeneralConfig.GCPJsonKeyFilePath, GeneralConfig.GCSBucketId, GeneralConfig.GCSFolderPath, GeneralConfig.GCSSubFolder)
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.Dsn,
						GeneralConfig.HookConfig.SplunkConfig.Token,
						GeneralConfig.HookConfig.SplunkConfig.Index,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblToken,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblIndex,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if GeneralConfig.HookConfig.GCPPubSubConfig.Enabled {
					err := gcp.NewGcpPubsubClient(
						vaultClient,
						GeneralConfig.HookConfig.GCPPubSubConfig.ProjectNumber,
						GeneralConfig.HookConfig.GCPPubSubConfig.IdentityPool,
						GeneralConfig.HookConfig.GCPPubSubConfig.IdentityProvider,
						GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.OIDCConfig.RoleID,
					).Publish(GeneralConfig.HookConfig.GCPPubSubConfig.Topic, telemetryClient.GetDataBytes())
					if err != nil {
						log.Entry().WithError(err).Warn("event publish failed")
					}
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME, GeneralConfig.HookConfig.PendoConfig.Token)
			npmExecuteTests(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addNpmExecuteTestsFlags(createNpmExecuteTestsCmd, &stepConfig)
	return createNpmExecuteTestsCmd
}

func addNpmExecuteTestsFlags(cmd *cobra.Command, stepConfig *npmExecuteTestsOptions) {
	cmd.Flags().StringVar(&stepConfig.InstallCommand, "installCommand", `npm ci`, "Command to be executed for installation`.")
	cmd.Flags().StringVar(&stepConfig.RunCommand, "runCommand", `npm run wdi5`, "Command to be executed for running tests`.")

	cmd.Flags().StringVar(&stepConfig.VaultUsername, "vaultUsername", os.Getenv("PIPER_vaultUsername"), "The base URL username.")
	cmd.Flags().StringVar(&stepConfig.VaultPassword, "vaultPassword", os.Getenv("PIPER_vaultPassword"), "The base URL password.")
	cmd.Flags().StringVar(&stepConfig.BaseURL, "baseUrl", `http://localhost:8080/index.html`, "Base URL of the application to be tested.")
	cmd.Flags().StringVar(&stepConfig.UsernameEnvVar, "usernameEnvVar", `wdi5_username`, "Env var for username.")
	cmd.Flags().StringVar(&stepConfig.PasswordEnvVar, "passwordEnvVar", `wdi5_password`, "Env var for password.")
	cmd.Flags().StringVar(&stepConfig.UrlOptionPrefix, "urlOptionPrefix", os.Getenv("PIPER_urlOptionPrefix"), "If you want to specify an extra option that the tested url it appended to.\nFor example if the test URL is `http://localhost and urlOptionPrefix is `--base-url=`,\nwe'll add `--base-url=http://localhost` to your runScript.\n")
	cmd.Flags().StringSliceVar(&stepConfig.Envs, "envs", []string{}, "List of environment variables to be set")
	cmd.Flags().StringSliceVar(&stepConfig.Paths, "paths", []string{}, "List of paths to be added to $PATH")
	cmd.Flags().StringVar(&stepConfig.WorkingDirectory, "workingDirectory", `.`, "Directory where your tests are located relative to the root of your project")

	cmd.MarkFlagRequired("runCommand")
}

// retrieve step metadata
func npmExecuteTestsMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "npmExecuteTests",
			Aliases:     []config.Alias{},
			Description: "Executes end-to-end tests using npm",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "installCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `npm ci`,
					},
					{
						Name:        "runCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `npm run wdi5`,
					},
					{
						Name: "vaultURLs",
						ResourceRef: []config.ResourceReference{
							{
								Name:    "appMetadataVaultSecretName",
								Type:    "vaultSecret",
								Default: "appMetadata",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "[]map[string]interface{}",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name: "vaultUsername",
						ResourceRef: []config.ResourceReference{
							{
								Name:    "appMetadataVaultSecretName",
								Type:    "vaultSecret",
								Default: "appMetadata",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_vaultUsername"),
					},
					{
						Name: "vaultPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:    "appMetadataVaultSecretName",
								Type:    "vaultSecret",
								Default: "appMetadata",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_vaultPassword"),
					},
					{
						Name:        "baseUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `http://localhost:8080/index.html`,
					},
					{
						Name:        "usernameEnvVar",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `wdi5_username`,
					},
					{
						Name:        "passwordEnvVar",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `wdi5_password`,
					},
					{
						Name:        "urlOptionPrefix",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_urlOptionPrefix"),
					},
					{
						Name:        "envs",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "paths",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "workingDirectory",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `.`,
					},
				},
			},
			Containers: []config.Container{
				{Name: "node", Image: "node:lts-bookworm", EnvVars: []config.EnvVar{{Name: "BASE_URL", Value: "${{params.baseUrl}}"}, {Name: "CREDENTIALS_ID", Value: "${{params.credentialsId}}"}, {Name: "no_proxy", Value: "localhost,selenium,$no_proxy"}, {Name: "NO_PROXY", Value: "localhost,selenium,$NO_PROXY"}}, WorkingDir: "/home/node"},
			},
			Sidecars: []config.Container{
				{Name: "selenium", Image: "selenium/standalone-chrome", EnvVars: []config.EnvVar{{Name: "NO_PROXY", Value: "localhost,selenium,$NO_PROXY"}, {Name: "no_proxy", Value: "localhost,selenium,$no_proxy"}}},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "reports",
						Type: "reports",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/e2e-results.xml", "type": "end-to-end-test"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
