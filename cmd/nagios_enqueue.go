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
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/spf13/cobra"
)

var errNotificationType = errors.New("notification-type must be one of: \"PROBLEM\", \"ACKNOWLEDGEMENT\", \"RECOVERY\"")
var errSourceType = errors.New("source-type must be one of: \"host\", \"service\"")
var errSeverity = errors.New("severity must be one of: \"critical\", \"warning\", \"error\", \"info\"")

var requiredFields = map[string][]string{
	"host":    {"HOSTNAME", "HOSTSTATE"},
	"service": {"HOSTNAME", "SERVICEDESC", "SERVICESTATE"},
}

var nagiosToPagerDutyEventType = map[string]string{
	"PROBLEM":         "trigger",
	"ACKNOWLEDGEMENT": "acknowledge",
	"RECOVERY":        "resolve",
}

func NewNagiosEnqueueCmd(config *Config) *cobra.Command {
	var customDetails map[string]string

	var sourceType string

	var sendEvent = eventsapi.EventV2{
		Payload: eventsapi.PayloadV2{},
	}

	requiredFlags := []string{"routing-key", "notification-type", "source-type", "severity"}

	cmd := &cobra.Command{
		Use:   "enqueue",
		Short: "Enqueue an event from Nagios to PagerDuty.",
		Long: fmt.Sprintf(`Enqueue an event from Nagios to PagerDuty.

	The following flags are required to be set for this command: %v.

	When the source type is "host", the following fields must be set using the -f flag:
	%v

	When the source type is "service", the following fields must be set using the -f flag:
	%v
		`, strings.Join(requiredFields["host"], ", "), strings.Join(requiredFields["service"], ", "), strings.Join(requiredFlags, ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := validateNagiosSendCommand(sendEvent, sourceType, customDetails)
			if err != nil {
				return err
			}

			transformedSendEvent, transformedCustomDetails := nagiosTransformations(sendEvent, sourceType, customDetails)
			return runSendCommand(config, transformedSendEvent, transformedCustomDetails)
		},
	}

	cmd.Flags().StringVarP(&sendEvent.RoutingKey, "routing-key", "k", "", "Service Events API Key (required)")
	cmd.Flags().StringVarP(&sendEvent.EventAction, "notification-type", "t", "", "The Nagios notification type (required)")
	cmd.Flags().StringVarP(&sourceType, "source-type", "u", "", "The Nagios source type (host or service, required)")
	cmd.Flags().StringVarP(&sendEvent.Payload.Severity, "severity", "e", "", "The perceived severity of the event (required)")
	cmd.Flags().StringVarP(&sendEvent.DedupKey, "dedup-key", "y", "", "Deduplication key for correlating triggers and resolves")
	cmd.Flags().StringToStringVarP(&customDetails, "field", "f", map[string]string{}, "Add given KEY=VALUE pair to the event details")

	for _, flag := range requiredFlags {
		cmd.MarkFlagRequired(flag)
	}

	return cmd
}

func nagiosTransformations(
	sendEvent eventsapi.EventV2, sourceType string, customDetails map[string]string,
) (eventsapi.EventV2, map[string]string) {
	sendEvent.Payload.Summary = buildEventDescription(sourceType, customDetails)
	sendEvent.EventAction = nagiosToPagerDutyEventType[sendEvent.EventAction]
	sendEvent.Payload.Source = customDetails["HOSTNAME"]
	if sendEvent.DedupKey == "" {
		sendEvent.DedupKey = buildDedupKey(sourceType, customDetails)
	}

	customDetails["pd_nagios_object"] = sourceType

	return sendEvent, customDetails
}

func buildEventDescription(sourceType string, customDetails map[string]string) string {
	descriptionFields := []string{}
	for _, field := range requiredFields[sourceType] {
		descriptionFields = append(descriptionFields, fmt.Sprintf("%v=%v", field, customDetails[field]))
	}
	return strings.Join(descriptionFields, "; ")
}

func buildDedupKey(sourceType string, customDetails map[string]string) string {
	if sourceType == "host" {
		return fmt.Sprintf("event_source=host;host_name=%v", customDetails["HOSTNAME"])
	}
	return fmt.Sprintf("event_source=service;host_name=%v;service_desc=%v", customDetails["HOSTNAME"], customDetails["SERVICEDESC"])
}

func validateNagiosSendCommand(sendEvent eventsapi.EventV2, sourceType string, customDetails map[string]string) error {
	allowedNotificationTypes := []string{"PROBLEM", "ACKNOWLEDGEMENT", "RECOVERY"}
	if err := validateEnumField(sendEvent.EventAction, allowedNotificationTypes, errNotificationType); err != nil {
		return err
	}

	allowedSourceTypes := []string{"host", "service"}
	if err := validateEnumField(sourceType, allowedSourceTypes, errSourceType); err != nil {
		return err
	}

	allowedSeverities := []string{"critical", "warning", "error", "info"}
	if err := validateEnumField(sendEvent.Payload.Severity, allowedSeverities, errSeverity); err != nil {
		return err
	}

	if err := validateCustomDetails(sourceType, customDetails); err != nil {
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

func validateCustomDetails(sourceType string, customDetails map[string]string) error {
	requiredKeys := requiredFields[sourceType]
	for _, key := range requiredKeys {
		if _, ok := customDetails[key]; !ok {
			errorString := fmt.Sprintf("The %v field must be set for source-type \"%v\" using the -f flag", key, sourceType)
			return errors.New(errorString)
		}
	}
	return nil
}
