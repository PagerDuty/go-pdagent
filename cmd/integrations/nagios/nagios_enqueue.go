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
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/spf13/cobra"
)

type nagiosEnqueueInput struct {
	serviceKey       string
	notificationType string
	sourceType       string
	incidentKey      string
	customFields     map[string]string
}

var clock common.Clock = common.NewClock()

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

	requiredFlags := []string{"service-key", "notification-type", "source-type"}

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

			return cmdutil.RunSendCommand(config, &sendEvent)
		},
	}

	cmd.Flags().StringVarP(&cmdInput.serviceKey, "service-key", "k", "", "Service Events API Key (required)")
	cmd.Flags().StringVarP(&cmdInput.notificationType, "notification-type", "t", "", "The Nagios notification type (required)")
	cmd.Flags().StringVarP(&cmdInput.sourceType, "source-type", "n", "", "The Nagios source type (host or service, required)")
	cmd.Flags().StringVarP(&cmdInput.incidentKey, "incident-key", "y", "", "Incident key for correlating triggers and resolves")
	cmd.Flags().StringToStringVarP(&cmdInput.customFields, "field", "f", map[string]string{}, "Add given KEY=VALUE pair to the event details")

	for _, flag := range requiredFlags {
		cmd.MarkFlagRequired(flag)
	}

	return cmd
}

func buildSendEvent(cmdInputs nagiosEnqueueInput) eventsapi.EventV1 {
	sendEvent := eventsapi.EventV1{
		ServiceKey:  cmdInputs.serviceKey,
		EventType:   nagiosToPagerDutyEventType[cmdInputs.notificationType],
		IncidentKey: cmdInputs.incidentKey,
		Description: buildEventDescription(cmdInputs),
		Details:     cmdutil.StringMapToInterfaceMap(cmdInputs.customFields),
		Agent: eventsapi.AgentContext{
			QueuedBy: "pd-nagios",
			QueuedAt: clock.Now().UTC().Format(time.RFC3339),
			AgentId:  common.UserAgent(),
		},
	}
	if sendEvent.IncidentKey == "" {
		sendEvent.IncidentKey = buildIncidentKey(cmdInputs)
	}

	sendEvent.Details["pd_nagios_object"] = cmdInputs.sourceType

	return sendEvent
}

func buildEventDescription(cmdInputs nagiosEnqueueInput) string {
	descriptionFields := []string{}
	for _, field := range requiredFields[cmdInputs.sourceType] {
		descriptionFields = append(descriptionFields, fmt.Sprintf("%v=%v", field, cmdInputs.customFields[field]))
	}
	return strings.Join(descriptionFields, "; ")
}

func buildIncidentKey(cmdInputs nagiosEnqueueInput) string {
	if cmdInputs.sourceType == "host" {
		return fmt.Sprintf("event_source=host;host_name=%v", cmdInputs.customFields["HOSTNAME"])
	}
	return fmt.Sprintf(
		"event_source=service;host_name=%v;service_desc=%v",
		cmdInputs.customFields["HOSTNAME"], cmdInputs.customFields["SERVICEDESC"],
	)
}

func validateNagiosSendCommand(cmdInputs nagiosEnqueueInput) error {
	if err := cmdutil.ValidateEnumField(cmdInputs.notificationType, allowedNotificationTypes, errNotificationType); err != nil {
		return err
	}

	if err := cmdutil.ValidateEnumField(cmdInputs.sourceType, allowedSourceTypes, errSourceType); err != nil {
		return err
	}

	if err := validateCustomDetails(cmdInputs); err != nil {
		return err
	}

	return nil
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
