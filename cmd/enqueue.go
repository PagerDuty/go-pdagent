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

func NewEnqueueCmd(config *cmdutil.Config) *cobra.Command {
	var customDetails map[string]string

	var sendEvent = eventsapi.EventV2{
		Payload: eventsapi.PayloadV2{},
	}

	cmd := &cobra.Command{
		Use:   "enqueue",
		Short: "Queue up a trigger, acknowledge, or resolve v2 event to PagerDuty",
		RunE: func(cmd *cobra.Command, args []string) error {
			sendEvent.Payload.CustomDetails = cmdutil.StringMapToInterfaceMap(customDetails)
			return cmdutil.RunSendCommand(config, &sendEvent)
		},
	}

	cmd.Flags().StringVarP(&sendEvent.RoutingKey, "routing-key", "k", "", "Service Events API Key")
	cmd.Flags().StringVarP(&sendEvent.EventAction, "event-action", "t", "", "The type of event")
	cmd.Flags().StringVarP(&sendEvent.DedupKey, "dedup-key", "y", "", "Deduplication key for correlating triggers and resolves")
	cmd.Flags().StringVarP(&sendEvent.Payload.Summary, "summary", "d", "", "A brief text summary of the event")
	cmd.Flags().StringVarP(&sendEvent.Payload.Source, "source", "u", "", "The unique location of the affected system")
	cmd.Flags().StringVarP(&sendEvent.Payload.Severity, "severity", "e", "error", "The perceived severity of the status the event is describing with respect to the affected system")
	cmd.Flags().StringVar(&sendEvent.Payload.Component, "component", "", "Component of the source machine that is responsible for the event")
	cmd.Flags().StringVarP(&sendEvent.Payload.Group, "group", "g", "", "Logical grouping of components of a service")
	cmd.Flags().StringVar(&sendEvent.Payload.Class, "class", "", "The class/type of the event")
	cmd.Flags().StringToStringVarP(&customDetails, "field", "f", map[string]string{}, "Add given KEY=VALUE pair to the event details")

	return cmd
}
