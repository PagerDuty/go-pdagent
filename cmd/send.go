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

	"github.com/spf13/cobra"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Queue up a trigger, acknowledge, or resolve event to PagerDuty",
	Long: `TODO: A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("send called", cmd.Flag("routing-key").Value)
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)

	sendCmd.PersistentFlags().StringP("routing-key", "k", "", "Service Events API Key")
	sendCmd.PersistentFlags().StringP("event-type", "t", "", "Event type")
	sendCmd.PersistentFlags().StringP("description", "d", "", "Short description of the problem")
	sendCmd.PersistentFlags().StringP("incident-key", "i", "", "Incident Key")
	sendCmd.PersistentFlags().StringP("client", "c", "", "Client")
	sendCmd.PersistentFlags().StringP("client-url", "c", "", "Client URL")
	sendCmd.PersistentFlags().StringP("field", "f", "", "Add given KEY=VALUE pair to the event details")
}
