// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
)

type githubPublishReleaseOptions struct {
	AddClosedIssues       bool     `json:"addClosedIssues,omitempty"`
	AddDeltaToLastRelease bool     `json:"addDeltaToLastRelease,omitempty"`
	APIURL                string   `json:"apiUrl,omitempty"`
	AssetPath             string   `json:"assetPath,omitempty"`
	AssetPathList         []string `json:"assetPathList,omitempty"`
	Commitish             string   `json:"commitish,omitempty"`
	ExcludeLabels         []string `json:"excludeLabels,omitempty"`
	Labels                []string `json:"labels,omitempty"`
	Owner                 string   `json:"owner,omitempty"`
	PreRelease            bool     `json:"preRelease,omitempty"`
	ReleaseBodyHeader     string   `json:"releaseBodyHeader,omitempty"`
	Repository            string   `json:"repository,omitempty"`
	ServerURL             string   `json:"serverUrl,omitempty"`
	TagPrefix             string   `json:"tagPrefix,omitempty"`
	Token                 string   `json:"token,omitempty"`
	UploadURL             string   `json:"uploadUrl,omitempty"`
	Version               string   `json:"version,omitempty"`
}

// GithubPublishReleaseCommand Publish a release in GitHub
func GithubPublishReleaseCommand() *cobra.Command {
	const STEP_NAME = "githubPublishRelease"

	metadata := githubPublishReleaseMetadata()
	var stepConfig githubPublishReleaseOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createGithubPublishReleaseCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Publish a release in GitHub",
		Long: `This step creates a tag in your GitHub repository together with a release.
The release can be filled with text plus additional information like:

* Closed pull request since last release
* Closed issues since last release
* Link to delta information showing all commits since last release

The result looks like

![Example release](../images/githubRelease.png)`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.Token)

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
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := cmd.Root().Context()
			tracer := telemetry.GetTracer(ctx)
			_, span := tracer.Start(ctx, "piper.step.run")
			span.SetAttributes(attribute.String("piper.step.name", STEP_NAME))

			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				defer span.End()
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
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME, GeneralConfig.HookConfig.PendoConfig.Token)
			githubPublishRelease(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGithubPublishReleaseFlags(createGithubPublishReleaseCmd, &stepConfig)
	return createGithubPublishReleaseCmd
}

func addGithubPublishReleaseFlags(cmd *cobra.Command, stepConfig *githubPublishReleaseOptions) {
	cmd.Flags().BoolVar(&stepConfig.AddClosedIssues, "addClosedIssues", false, "If set to `true`, closed issues and merged pull-requests since the last release will added below the `releaseBodyHeader`")
	cmd.Flags().BoolVar(&stepConfig.AddDeltaToLastRelease, "addDeltaToLastRelease", false, "If set to `true`, a link will be added to the release information that brings up all commits since the last release.")
	cmd.Flags().StringVar(&stepConfig.APIURL, "apiUrl", `https://api.github.com`, "Set the GitHub API url.")
	cmd.Flags().StringVar(&stepConfig.AssetPath, "assetPath", os.Getenv("PIPER_assetPath"), "Path to a release asset which should be uploaded to the list of release assets.")
	cmd.Flags().StringSliceVar(&stepConfig.AssetPathList, "assetPathList", []string{}, "List of paths to a release asset which should be uploaded to the list of release assets.")
	cmd.Flags().StringVar(&stepConfig.Commitish, "commitish", `master`, "Target git commitish for the release")
	cmd.Flags().StringSliceVar(&stepConfig.ExcludeLabels, "excludeLabels", []string{}, "Allows to exclude issues with dedicated list of labels.")
	cmd.Flags().StringSliceVar(&stepConfig.Labels, "labels", []string{}, "Labels to include in issue search.")
	cmd.Flags().StringVar(&stepConfig.Owner, "owner", os.Getenv("PIPER_owner"), "Name of the GitHub organization.")
	cmd.Flags().BoolVar(&stepConfig.PreRelease, "preRelease", false, "If set to `true` the release will be marked as Pre-release.")
	cmd.Flags().StringVar(&stepConfig.ReleaseBodyHeader, "releaseBodyHeader", os.Getenv("PIPER_releaseBodyHeader"), "Content which will appear for the release.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Name of the GitHub repository.")
	cmd.Flags().StringVar(&stepConfig.ServerURL, "serverUrl", `https://github.com`, "GitHub server url for end-user access.")
	cmd.Flags().StringVar(&stepConfig.TagPrefix, "tagPrefix", ``, "Defines a prefix to be added to the tag.")
	cmd.Flags().StringVar(&stepConfig.Token, "token", os.Getenv("PIPER_token"), "GitHub personal access token as per https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line")
	cmd.Flags().StringVar(&stepConfig.UploadURL, "uploadUrl", `https://uploads.github.com`, "Set the GitHub API url.")
	cmd.Flags().StringVar(&stepConfig.Version, "version", os.Getenv("PIPER_version"), "Define the version number which will be written as tag as well as release name.")

	cmd.MarkFlagRequired("apiUrl")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("serverUrl")
	cmd.MarkFlagRequired("token")
	cmd.MarkFlagRequired("uploadUrl")
	cmd.MarkFlagRequired("version")
}

// retrieve step metadata
func githubPublishReleaseMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "githubPublishRelease",
			Aliases:     []config.Alias{},
			Description: "Publish a release in GitHub",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "githubTokenCredentialsId", Description: "Jenkins 'Secret text' credentials ID containing token to authenticate to GitHub.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "addClosedIssues",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "addDeltaToLastRelease",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "apiUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubApiUrl"}},
						Default:     `https://api.github.com`,
					},
					{
						Name:        "assetPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_assetPath"),
					},
					{
						Name:        "assetPathList",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name: "commitish",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "git/headCommitId",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   `master`,
					},
					{
						Name:        "excludeLabels",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "labels",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name: "owner",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "github/owner",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "githubOrg"}},
						Default:   os.Getenv("PIPER_owner"),
					},
					{
						Name:        "preRelease",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "releaseBodyHeader",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_releaseBodyHeader"),
					},
					{
						Name: "repository",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "github/repository",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "githubRepo"}},
						Default:   os.Getenv("PIPER_repository"),
					},
					{
						Name:        "serverUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubServerUrl"}},
						Default:     `https://github.com`,
					},
					{
						Name:        "tagPrefix",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     ``,
					},
					{
						Name: "token",
						ResourceRef: []config.ResourceReference{
							{
								Name: "githubTokenCredentialsId",
								Type: "secret",
							},

							{
								Name:    "githubVaultSecretName",
								Type:    "vaultSecret",
								Default: "github",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "githubToken"}, {Name: "access_token"}},
						Default:   os.Getenv("PIPER_token"),
					},
					{
						Name:        "uploadUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubUploadUrl"}},
						Default:     `https://uploads.github.com`,
					},
					{
						Name: "version",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_version"),
					},
				},
			},
		},
	}
	return theMetaData
}
