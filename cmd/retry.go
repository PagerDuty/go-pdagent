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
	"github.com/PagerDuty/pagerduty-agent/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
)

// queueCmd represents the storage command
var retryCmd = &cobra.Command{
	Use:   "retry",
	Short: "Retry failed events.",
	Run: func(cmd *cobra.Command, args []string) {
		c := client.NewClient(viper.GetString("address"), viper.GetString("secret"))
		rk, err := cmd.Flags().GetString("routing-key")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		resp, err := c.QueueRetry(rk)
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
	queueCmd.AddCommand(retryCmd)
}
