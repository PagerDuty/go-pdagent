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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var notificationType string

func hasKey(theMap map[string]string, theKey string) bool {
	_, ok := theMap[theKey]
	return ok
}

func runSendNagiosCommand(cmd *cobra.Command, args []string) {
	if sendEvent.EventAction == "PROBLEM" {
		sendEvent.EventAction = "trigger"
	} else if sendEvent.EventAction == "ACKNOWLEDGEMENT" {
		sendEvent.EventAction = "acknowledge"
	} else if sendEvent.EventAction == "RECOVERY" {
		sendEvent.EventAction = "resolve"
	} else {
		println("Error: Event type must be PROBLEM, ACKNOWLEDGEMENT or RECOVERY")
		os.Exit(1)
	}

	if notificationType == "service" {
		if !(hasKey(customDetails, "HOSTNAME") && hasKey(customDetails, "SERVICEDESC") && hasKey(customDetails, "SERVICESTATE")) {
			println("Error: Notification type 'service' requires HOSTNAME, SERVICEDESC and SERVICESTATE fields")
			os.Exit(1)
		}
		if sendEvent.DedupKey == "" {
			sendEvent.DedupKey = fmt.Sprintf("event_source=service;host_name=%s;service_desc=%s",
				customDetails["HOSTNAME"], customDetails["SERVICEDESC"])
		}
		if sendEvent.Payload.Source == "" {
			sendEvent.Payload.Source = fmt.Sprintf("%s on %s",
				customDetails["HOSTNAME"], customDetails["SERVICEDESC"])
		}
		if sendEvent.Payload.Summary == "" {
			sendEvent.Payload.Summary = fmt.Sprintf("SERVICEDESC=%s; SERVICESTATE=%s; HOSTNAME=%s",
				customDetails["SERVICEDESC"], customDetails["SERVICESTATE"], customDetails["HOSTNAME"])
		}
	} else if notificationType == "host" {
		if !(hasKey(customDetails, "HOSTNAME") && hasKey(customDetails, "HOSTSTATE")) {
			println("Error: Notification type 'host' requires HOSTNAME and HOSTSTATE fields")
			os.Exit(1)
		}
		if sendEvent.DedupKey == "" {
			sendEvent.DedupKey = fmt.Sprintf("event_source=host;host_name=%s",
				customDetails["HOSTNAME"])
		}
		if sendEvent.Payload.Source == "" {
			sendEvent.Payload.Source = customDetails["HOSTNAME"]
		}
		if sendEvent.Payload.Summary == "" {
			sendEvent.Payload.Summary = fmt.Sprintf("HOSTNAME=%s; HOSTSTATE=%s",
				customDetails["HOSTNAME"], customDetails["HOSTSTATE"])
		}
	} else {
		println("Error: Notification type must be service or host")
		os.Exit(1)
	}
	customDetails["pd_nagios_object"] = notificationType

	runSendCommand(cmd, args)
}

var sendNagiosCmd = &cobra.Command{
	Use:   "send-nagios",
	Short: "Queue up a trigger, acknowledge, or resolve event to PagerDuty from Nagios",
	Long: `Queue up a trigger, acknowledge, or resolve V2 event to PagerDuty 
using a Nagios-compatible set of flags.`,
	Run: runSendNagiosCommand,
}

func init() {
	rootCmd.AddCommand(sendNagiosCmd)

	sendNagiosCmd.Flags().StringVarP(&sendEvent.RoutingKey, "routing-key", "k", "", "Service Events API Key")
	sendNagiosCmd.Flags().StringVarP(&sendEvent.EventAction, "event-type", "t", "", "Event type")
	sendNagiosCmd.Flags().StringVarP(&sendEvent.DedupKey, "incident-key", "i", "", "Incident Key")
	sendNagiosCmd.Flags().StringVarP(&sendEvent.Payload.Summary, "summary", "d", "", "A brief text summary of the event")
	sendNagiosCmd.Flags().StringVarP(&sendEvent.Payload.Source, "source", "u", "", "The unique location of the affected system")
	sendNagiosCmd.Flags().StringVarP(&sendEvent.Payload.Severity, "severity", "e", "critical", "The perceived severity of the status the event is describing with respect to the affected system")
	sendNagiosCmd.Flags().StringVar(&sendEvent.Payload.Component, "component", "", "Component of the source machine that is responsible for the event")
	sendNagiosCmd.Flags().StringVarP(&sendEvent.Payload.Group, "group", "g", "", "Logical grouping of components of a service")
	sendNagiosCmd.Flags().StringVar(&sendEvent.Payload.Class, "class", "", "The class/type of the event")
	sendNagiosCmd.Flags().StringToStringVarP(&customDetails, "field", "f", map[string]string{}, "Add given KEY=VALUE pair to the event details")
	sendNagiosCmd.Flags().StringVarP(&notificationType, "notification-type", "n", "", "Notification type")
	sendNagiosCmd.MarkFlagRequired("routing-key")
	sendNagiosCmd.MarkFlagRequired("event-type")
	sendNagiosCmd.MarkFlagRequired("notification-type")
}
