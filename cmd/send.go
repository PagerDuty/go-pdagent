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
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/spf13/cobra"
)

func NewSendCmd(config *Config) *cobra.Command {
	var customDetails map[string]string

	var sendEvent = eventsapi.EventV2{
		Payload: eventsapi.PayloadV2{},
	}

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Queue up a trigger, acknowledge, or resolve event to PagerDuty",
		Long: `Queue up a trigger, acknowledge, or resolve V2 event to PagerDuty 
		using a backwards-compatible set of flags.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			return runSendCommand(config, sendEvent, customDetails)
		},
	}

	cmd.Flags().StringVarP(&sendEvent.RoutingKey, "routing-key", "k", "", "Service Events API Key")
	cmd.Flags().StringVarP(&sendEvent.EventAction, "event-type", "t", "", "Event type")
	cmd.Flags().StringVarP(&sendEvent.Payload.Summary, "description", "d", "", "Short description of the problem")
	cmd.Flags().StringVarP(&sendEvent.DedupKey, "incident-key", "i", "", "Incident Key")
	cmd.Flags().StringVarP(&sendEvent.Payload.Component, "client", "c", "", "Client")
	cmd.Flags().StringVarP(&sendEvent.Payload.Source, "client-url", "u", "", "Client URL")
	cmd.Flags().StringToStringVarP(&customDetails, "field", "f", map[string]string{}, "Add given KEY=VALUE pair to the event details")

	return cmd
}
