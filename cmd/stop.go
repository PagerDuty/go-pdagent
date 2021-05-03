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

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

func NewServerStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Gracefully stop a running pdagent server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStopCommand()
		},
	}

	return cmd
}

func runStopCommand() error {
	pidfile := viper.GetString("pidfile")

	if err := common.TerminateProcess(pidfile); err != nil {
		fmt.Printf("Error terminating server: %v\n", err)

		if err == common.ErrPidfileDoesntExist {
			fmt.Println("This normally means a server isn't currently running, or you're running this command using a different configuration.")
		}

		os.Exit(1)
	}

	fmt.Println("Server terminated.")
	os.Exit(0)
	return nil
}
