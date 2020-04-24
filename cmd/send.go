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
	"io/ioutil"
	"os"

	"github.com/PagerDuty/go-pdagent/pkg/client"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var customDetails map[string]string

var sendEvent = eventsapi.EventV2{
	Payload: eventsapi.PayloadV2{
		// TODO Support as CLI option.
		Severity: "error",
	},
}

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Queue up a trigger, acknowledge, or resolve event to PagerDuty",
	Run: func(cmd *cobra.Command, args []string) {
		c := client.NewClient(viper.GetString("address"), viper.GetString("secret"))

		// Manually mapping as a workaround for the map type mismatch.
		sendEvent.Payload.CustomDetails = map[string]interface{}{}
		for k, v := range customDetails {
			sendEvent.Payload.CustomDetails[k] = v
		}

		resp, err := c.Send(sendEvent)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(string(respBody))
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)

	sendCmd.PersistentFlags().StringVarP(&sendEvent.RoutingKey, "routing-key", "k", "", "Service Events API Key")
	sendCmd.PersistentFlags().StringVarP(&sendEvent.EventAction, "event-type", "t", "", "Event type")
	sendCmd.PersistentFlags().StringVarP(&sendEvent.Payload.Summary, "description", "d", "", "Short description of the problem")
	sendCmd.PersistentFlags().StringVarP(&sendEvent.DedupKey, "incident-key", "i", "", "Incident Key")
	sendCmd.PersistentFlags().StringVarP(&sendEvent.Payload.Component, "client", "c", "", "Client")
	sendCmd.PersistentFlags().StringVarP(&sendEvent.Payload.Source, "client-url", "u", "", "Client URL")
	sendCmd.PersistentFlags().StringToStringVarP(&customDetails, "field", "f", map[string]string{}, "Add given KEY=VALUE pair to the event details")
}
