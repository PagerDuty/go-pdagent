/*
Copyright © 2020 PagerDuty, Inc. <info@pagerduty.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package nagios

import (
	"fmt"
	"strings"

	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/spf13/cobra"
)

type nagiosEnqueueInput struct {
	routingKey       string
	notificationType string
	sourceType       string
	dedupKey         string
	customFields     map[string]string
}

const defaultNagiosIntegrationSeverity = "error"

var allowedNotificationTypes = []string{"PROBLEM", "ACKNOWLEDGEMENT", "RECOVERY"}
var allowedSourceTypes = []string{"host", "service"}

var errNotificationType = fmt.Errorf("notification-type must be one of: %v", strings.Join(allowedNotificationTypes, ", "))
var errSourceType = fmt.Errorf("source-type must be one of: %v", strings.Join(allowedSourceTypes, ", "))

var requiredFields = map[string][]string{
	"host":    {"HOSTNAME", "HOSTSTATE"},
	"service": {"HOSTNAME", "SERVICEDESC", "SERVICESTATE"},
}

var nagiosToPagerDutyEventType = map[string]string{
	"PROBLEM":         "trigger",
	"ACKNOWLEDGEMENT": "acknowledge",
	"RECOVERY":        "resolve",
}

func NewNagiosEnqueueCmd(config *cmdutil.Config) *cobra.Command {
	var cmdInput nagiosEnqueueInput

	requiredFlags := []string{"routing-key", "notification-type", "source-type"}

	cmd := &cobra.Command{
		Use:   "enqueue",
		Short: "Enqueue an event from Nagios to PagerDuty.",
		Long: fmt.Sprintf(`Enqueue an event from Nagios to PagerDuty.

	The following flags are required to be set for this command: %v.

	When the source type is "host", the following fields must be set using the -f flag:
	%v

	When the source type is "service", the following fields must be set using the -f flag:
	%v
		`, strings.Join(requiredFlags, ", "), strings.Join(requiredFields["host"], ", "), strings.Join(requiredFields["service"], ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := validateNagiosSendCommand(cmdInput)
			if err != nil {
				return err
			}

			sendEvent := buildSendEvent(cmdInput)

			return cmdutil.RunSendCommand(config, sendEvent, eventsapi.EventVersion2)
		},
	}

	cmd.Flags().StringVarP(&cmdInput.routingKey, "routing-key", "k", "", "Service Events API Key (required)")
	cmd.Flags().StringVarP(&cmdInput.notificationType, "notification-type", "t", "", "The Nagios notification type (required)")
	cmd.Flags().StringVarP(&cmdInput.sourceType, "source-type", "n", "", "The Nagios source type (host or service, required)")
	cmd.Flags().StringVarP(&cmdInput.dedupKey, "dedup-key", "y", "", "Deduplication key for correlating triggers and resolves")
	cmd.Flags().StringToStringVarP(&cmdInput.customFields, "field", "f", map[string]string{}, "Add given KEY=VALUE pair to the event details")

	for _, flag := range requiredFlags {
		cmd.MarkFlagRequired(flag)
	}

	return cmd
}

func buildSendEvent(cmdInputs nagiosEnqueueInput) eventsapi.EventV2 {
	sendEvent := eventsapi.EventV2{
		RoutingKey:  cmdInputs.routingKey,
		EventAction: nagiosToPagerDutyEventType[cmdInputs.notificationType],
		DedupKey:    cmdInputs.dedupKey,
		Payload: eventsapi.PayloadV2{
			Summary:  buildEventDescription(cmdInputs),
			Source:   cmdInputs.customFields["HOSTNAME"],
			Severity: defaultNagiosIntegrationSeverity,
		},
	}
	if sendEvent.DedupKey == "" {
		sendEvent.DedupKey = buildDedupKey(cmdInputs)
	}

	customDetails := cmdInputs.customFields
	customDetails["pd_nagios_object"] = cmdInputs.sourceType

	// Manually mapping as a workaround for the map type mismatch.
	sendEvent.Payload.CustomDetails = map[string]interface{}{}
	for k, v := range customDetails {
		sendEvent.Payload.CustomDetails[k] = v
	}

	return sendEvent
}

func buildEventDescription(cmdInputs nagiosEnqueueInput) string {
	descriptionFields := []string{}
	for _, field := range requiredFields[cmdInputs.sourceType] {
		descriptionFields = append(descriptionFields, fmt.Sprintf("%v=%v", field, cmdInputs.customFields[field]))
	}
	return strings.Join(descriptionFields, "; ")
}

func buildDedupKey(cmdInputs nagiosEnqueueInput) string {
	if cmdInputs.sourceType == "host" {
		return fmt.Sprintf("event_source=host;host_name=%v", cmdInputs.customFields["HOSTNAME"])
	}
	return fmt.Sprintf(
		"event_source=service;host_name=%v;service_desc=%v",
		cmdInputs.customFields["HOSTNAME"], cmdInputs.customFields["SERVICEDESC"],
	)
}

func validateNagiosSendCommand(cmdInputs nagiosEnqueueInput) error {
	if err := validateEnumField(cmdInputs.notificationType, allowedNotificationTypes, errNotificationType); err != nil {
		return err
	}

	if err := validateEnumField(cmdInputs.sourceType, allowedSourceTypes, errSourceType); err != nil {
		return err
	}

	if err := validateCustomDetails(cmdInputs); err != nil {
		return err
	}

	return nil
}

func validateEnumField(inputVal string, allowedValues []string, err error) error {
	for _, value := range allowedValues {
		if value == inputVal {
			return nil
		}
	}
	return err
}

func validateCustomDetails(cmdInputs nagiosEnqueueInput) error {
	requiredKeys := requiredFields[cmdInputs.sourceType]
	for _, key := range requiredKeys {
		if _, ok := cmdInputs.customFields[key]; !ok {
			return fmt.Errorf("the %v field must be set for source-type \"%v\" using the -f flag", key, cmdInputs.sourceType)
		}
	}
	return nil
}
