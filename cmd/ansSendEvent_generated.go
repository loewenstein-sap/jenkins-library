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

type ansSendEventOptions struct {
	AnsServiceKey    string                 `json:"ansServiceKey,omitempty"`
	EventType        string                 `json:"eventType,omitempty"`
	Severity         string                 `json:"severity,omitempty" validate:"possible-values=INFO NOTICE WARNING ERROR FATAL"`
	Category         string                 `json:"category,omitempty" validate:"possible-values=NOTIFICATION ALERT EXCEPTION"`
	Subject          string                 `json:"subject,omitempty"`
	Body             string                 `json:"body,omitempty"`
	Priority         int                    `json:"priority,omitempty"`
	Tags             map[string]interface{} `json:"tags,omitempty"`
	ResourceName     string                 `json:"resourceName,omitempty"`
	ResourceType     string                 `json:"resourceType,omitempty"`
	ResourceInstance string                 `json:"resourceInstance,omitempty"`
	ResourceTags     map[string]interface{} `json:"resourceTags,omitempty"`
}

// AnsSendEventCommand Send Event to the SAP Alert Notification Service
func AnsSendEventCommand() *cobra.Command {
	const STEP_NAME = "ansSendEvent"

	metadata := ansSendEventMetadata()
	var stepConfig ansSendEventOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createAnsSendEventCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Send Event to the SAP Alert Notification Service",
		Long:  `With this step one can send an Event to the SAP Alert Notification Service.`,
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
			log.RegisterSecret(stepConfig.AnsServiceKey)

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
			ansSendEvent(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addAnsSendEventFlags(createAnsSendEventCmd, &stepConfig)
	return createAnsSendEventCmd
}

func addAnsSendEventFlags(cmd *cobra.Command, stepConfig *ansSendEventOptions) {
	cmd.Flags().StringVar(&stepConfig.AnsServiceKey, "ansServiceKey", os.Getenv("PIPER_ansServiceKey"), "Service key JSON string to access the SAP Alert Notification Service")
	cmd.Flags().StringVar(&stepConfig.EventType, "eventType", `Piper`, "Type of the event")
	cmd.Flags().StringVar(&stepConfig.Severity, "severity", `INFO`, "Event severity")
	cmd.Flags().StringVar(&stepConfig.Category, "category", `NOTIFICATION`, "Event category")
	cmd.Flags().StringVar(&stepConfig.Subject, "subject", `ansSendEvent`, "Short description of the event")
	cmd.Flags().StringVar(&stepConfig.Body, "body", `Call from Piper step ansSendEvent`, "Detailed description of the event")
	cmd.Flags().IntVar(&stepConfig.Priority, "priority", 0, "Event priority in the range of 1 to 1000")

	cmd.Flags().StringVar(&stepConfig.ResourceName, "resourceName", `Pipeline`, "Unique resource name")
	cmd.Flags().StringVar(&stepConfig.ResourceType, "resourceType", `Pipeline`, "Resource type identifier")
	cmd.Flags().StringVar(&stepConfig.ResourceInstance, "resourceInstance", os.Getenv("PIPER_resourceInstance"), "Optional resource instance identifier")

	cmd.MarkFlagRequired("ansServiceKey")
}

// retrieve step metadata
func ansSendEventMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "ansSendEvent",
			Aliases:     []config.Alias{},
			Description: "Send Event to the SAP Alert Notification Service",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "ansServiceKeyCredentialsId", Description: "Jenkins secret text credential ID containing the service key to access the SAP Alert Notification Service", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "ansServiceKey",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "ansServiceKeyCredentialsId",
								Param: "ansServiceKey",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_ansServiceKey"),
					},
					{
						Name:        "eventType",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `Piper`,
					},
					{
						Name:        "severity",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `INFO`,
					},
					{
						Name:        "category",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `NOTIFICATION`,
					},
					{
						Name:        "subject",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `ansSendEvent`,
					},
					{
						Name:        "body",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `Call from Piper step ansSendEvent`,
					},
					{
						Name:        "priority",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     0,
					},
					{
						Name:        "tags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "map[string]interface{}",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "resourceName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `Pipeline`,
					},
					{
						Name:        "resourceType",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `Pipeline`,
					},
					{
						Name:        "resourceInstance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_resourceInstance"),
					},
					{
						Name:        "resourceTags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "map[string]interface{}",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
