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
	"os"

	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/PagerDuty/go-pdagent/pkg/persistentqueue"
	"github.com/PagerDuty/go-pdagent/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var errInvalidRegion = errors.New(`region must be either "us" or "eu"`)

func NewServerCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the server daemon.",
		Long:  `Starts the daemon server and begins processing any event backlog.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServerCommand()
		},
	}

	defaults := cmdutil.GetDefaults()

	cmd.PersistentFlags().String("database", defaults.Database, "database file for event queuing (default is /var/db/pdagent/agent.db)")
	cmd.PersistentFlags().String("region", defaults.Region, `PagerDuty region the daemon sends events to, either "us" or "eu"`)

	if err := viper.BindPFlag("database", cmd.PersistentFlags().Lookup("database")); err != nil {
		fmt.Println(err)
	}
	if err := viper.BindPFlag("region", cmd.PersistentFlags().Lookup("region")); err != nil {
		fmt.Println(err)
	}

	cmd.AddCommand(NewServerStopCmd())

	return cmd
}

func runServerCommand() error {
	address := viper.GetString("address")
	database := viper.GetString("database")
	pidfile := viper.GetString("pidfile")
	secret := viper.GetString("secret")
	region := viper.GetString("region")

	allowedRegions := []string{"us", "eu"}
	if err := cmdutil.ValidateEnumField(region, allowedRegions, errInvalidRegion); err != nil {
		return err
	}

	queue := persistentqueue.NewPersistentQueue(persistentqueue.WithFile(database))

	server := server.NewServer(address, secret, pidfile, queue)
	err := server.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}
