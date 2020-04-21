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
	"github.com/PagerDuty/pagerduty-agent/pkg/common"
	"github.com/spf13/viper"
	"os"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Gracefully stop a running pdagent server.",
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

func init() {
	serverCmd.AddCommand(stopCmd)
}
