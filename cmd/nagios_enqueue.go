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
	"io/ioutil"
	"strings"

	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/spf13/cobra"
)

func NewNagiosEnqueue(config *Config) *cobra.Command {
	var customDetails map[string]string

	var sourceType string

	var sendEvent = eventsapi.EventV2{
		Payload: eventsapi.PayloadV2{},
	}

	cmd := &cobra.Command{
		Use:   "enqueue",
		Short: "Enqueue an event from Nagios to PagerDuty.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := validateNagiosSendCommand(sendEvent, sourceType, customDetails)
			if err != nil {
				return err
			}
			transformedSendEvent := transformNagiosSendEvent(sendEvent, sourceType, customDetails)

			return runNagiosSendCommand(config, sourceType, transformedSendEvent, customDetails)
		},
	}

	cmd.Flags().StringVarP(&sendEvent.RoutingKey, "routing-key", "k", "", "Service Events API Key (required)")
	cmd.Flags().StringVarP(&sendEvent.EventAction, "notification-type", "t", "", "The Nagios notification type (required)")
	cmd.Flags().StringVarP(&sourceType, "source-type", "u", "", "The Nagios source type (host or service, required)")
	cmd.Flags().StringVarP(&sendEvent.Payload.Severity, "severity", "e", "", "The perceived severity of the event (required)")
	cmd.Flags().StringVarP(&sendEvent.DedupKey, "dedup-key", "y", "", "Deduplication key for correlating triggers and resolves")
	cmd.Flags().StringToStringVarP(&customDetails, "field", "f", map[string]string{}, "Add given KEY=VALUE pair to the event details")

	cmd.MarkFlagRequired("routing-key")
	cmd.MarkFlagRequired("notification-type")
	cmd.MarkFlagRequired("source-type")
	cmd.MarkFlagRequired("severity")

	return cmd
}

func validateNagiosSendCommand(sendEvent eventsapi.EventV2, sourceType string, customDetails map[string]string) error {
	err := validateNotificationType(sendEvent.EventAction)
	if err != nil {
		return err
	}

	err = validateSourceType(sourceType)
	if err != nil {
		return err
	}

	err = validateSeverity(sendEvent.Payload.Severity)
	if err != nil {
		return err
	}

	err = validateCustomDetails(sourceType, customDetails)
	if err != nil {
		return err
	}

	return nil
}

func runNagiosSendCommand(config *Config, sourceType string, sendEvent eventsapi.EventV2, customDetails map[string]string) error {
	c, _ := config.Client()

	resp, err := c.Send(sendEvent)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(string(respBody))
	return nil
}

func transformNagiosSendEvent(sendEvent eventsapi.EventV2, sourceType string, customDetails map[string]string) eventsapi.EventV2 {
	sendEvent.Payload.Source = customDetails["HOSTNAME"]
	sendEvent.Payload.Summary = buildEventDescription(sourceType, customDetails)

	// Manually mapping as a workaround for the map type mismatch.
	sendEvent.Payload.CustomDetails = map[string]interface{}{}
	for k, v := range customDetails {
		sendEvent.Payload.CustomDetails[k] = v
	}
	sendEvent.Payload.CustomDetails["pd_nagios_object"] = sourceType

	nagiosToPagerDutyEventType := map[string]string{
		"PROBLEM":         "trigger",
		"ACKNOWLEDGEMENT": "acknowledge",
		"RECOVERY":        "resolve",
	}

	sendEvent.EventAction = nagiosToPagerDutyEventType[sendEvent.EventAction]

	if sendEvent.DedupKey == "" {
		sendEvent.DedupKey = buildDedupKey(sourceType, customDetails)
	}

	return sendEvent
}

func buildEventDescription(sourceType string, customDetails map[string]string) string {
	descriptionFields := []string{}
	for _, field := range requiredFields()[sourceType] {
		descriptionFields = append(descriptionFields, fmt.Sprintf("%v=%v", field, customDetails[field]))
	}
	return strings.Join(descriptionFields, "; ")
}

func requiredFields() map[string][]string {
	return map[string][]string{
		"host":    {"HOSTNAME", "HOSTSTATE"},
		"service": {"HOSTNAME", "SERVICEDESC", "SERVICESTATE"},
	}
}

func buildDedupKey(sourceType string, customDetails map[string]string) string {
	if sourceType == "service" {
		return fmt.Sprintf(
			"event_source=service;host_name=%v;service_desc=%v", customDetails["HOSTNAME"], customDetails["SERVICEDESC"],
		)
	}
	return fmt.Sprintf("event_source=host;host_name=%v", customDetails["HOSTNAME"])
}

func validateNotificationType(notificationType string) error {
	allowedValues := []string{"PROBLEM", "ACKNOWLEDGEMENT", "RECOVERY"}
	for _, value := range allowedValues {
		if notificationType == value {
			return nil
		}
	}

	err := errors.New("notification-type must be one of: \"PROBLEM\", \"ACKNOWLEDGEMENT\", \"RECOVERY\"")
	return err
}

func validateSourceType(sourceType string) error {
	allowedValues := []string{"host", "service"}
	for _, value := range allowedValues {
		if sourceType == value {
			return nil
		}
	}

	err := errors.New("source-type must be one of: \"host\", \"service\"")
	return err
}

func validateSeverity(severity string) error {
	allowedValues := []string{"critical", "warning", "error", "info"}
	for _, value := range allowedValues {
		if severity == value {
			return nil
		}
	}

	err := errors.New("severity must be one of: \"critical\", \"warning\", \"error\", \"info\"")
	return err
}

func validateCustomDetails(sourceType string, customDetails map[string]string) error {
	requiredKeys := requiredFields()[sourceType]
	for _, key := range requiredKeys {
		if _, ok := customDetails[key]; !ok {
			errorString := fmt.Sprintf("The %v field must be set for source-type \"%v\"", key, sourceType)
			err := errors.New(errorString)
			return err
		}
	}

	return nil
}
