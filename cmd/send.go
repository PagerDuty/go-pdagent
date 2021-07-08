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
	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/spf13/cobra"
)

func NewSendCmd(config *cmdutil.Config) *cobra.Command {
	var customDetails map[string]string

	var sendEvent = eventsapi.EventV1{
		Details: eventsapi.DetailsV1{},
	}

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Queue up a trigger, acknowledge, or resolve a V1 event to PagerDuty",
		Long: `Queue up a trigger, acknowledge, or resolve V1 event to PagerDuty
		using a backwards-compatible set of flags.

		Required flags: "routing-key", "event-type"`,

		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunSendCommand(config, &sendEvent, customDetails)
		},
	}

	cmd.Flags().StringVarP(&sendEvent.ServiceKey, "service-key", "k", "", "Service Events API Key")
	cmd.Flags().StringVarP(&sendEvent.EventType, "event-type", "t", "", `Event type, either "trigger", "acknowledge", or "resolve"`)
	cmd.Flags().StringVarP(&sendEvent.Description, "description", "d", "", "Short description of the problem")
	cmd.Flags().StringVarP(&sendEvent.IncidentKey, "incident-key", "i", "", "Incident Key")
	cmd.Flags().StringVarP(&sendEvent.Client, "client", "c", "", "Client")
	cmd.Flags().StringVarP(&sendEvent.ClientURL, "client-url", "u", "", "Client URL")
	cmd.Flags().StringToStringVarP(&customDetails, "field", "f", map[string]string{}, "Add given KEY=VALUE pair to the event details")

	cmd.MarkFlagRequired("routing-key")
	cmd.MarkFlagRequired("event-type")

	return cmd
}
