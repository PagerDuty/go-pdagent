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
	"io"
	"os"

	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewHealthCmd(config *cmdutil.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check the health of the server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHealthCommand(config)
		},
	}

	return cmd
}

func runHealthCommand(config *cmdutil.Config) error {
	c, _ := config.Client()

	resp, err := c.HealthCheck()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(respBody))
	return nil
}
