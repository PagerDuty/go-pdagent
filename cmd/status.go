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

	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewQueueStatusCmd(config *cmdutil.Config) *cobra.Command {
	var routingKey string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print queue status information.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusCommand(config, routingKey)
		},
	}

	cmd.Flags().StringVarP(&routingKey, "routing-key", "k", "", "The Events API Key to check")

	return cmd
}

func runStatusCommand(config *cmdutil.Config, routingKey string) error {
	c, _ := config.Client()

	resp, err := c.QueueStatus(routingKey)
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
	return nil
}
