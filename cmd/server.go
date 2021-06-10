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

	"github.com/PagerDuty/go-pdagent/cmd/cmdutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewServerCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the server daemon.",
		Long:  `Starts the daemon server and begins processing any event backlog.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunServerCommand()
		},
	}

	defaults := cmdutil.GetDefaults()

	cmd.PersistentFlags().String("database", defaults.Database, "database file for event queuing (default is /var/db/pdagent/agent.db)")
	if err := viper.BindPFlag("database", cmd.PersistentFlags().Lookup("database")); err != nil {
		fmt.Println(err)
	}

	cmd.AddCommand(NewServerStopCmd())

	return cmd
}
