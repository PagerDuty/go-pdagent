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
package cmdutil

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/PagerDuty/go-pdagent/pkg/persistentqueue"
	"github.com/PagerDuty/go-pdagent/pkg/server"
	"github.com/spf13/viper"
)

func RunSendCommand(config *Config, sendEvent eventsapi.EventV2, customDetails map[string]string) error {
	c, _ := config.Client()

	// Manually mapping as a workaround for the map type mismatch.
	sendEvent.Payload.CustomDetails = map[string]interface{}{}
	for k, v := range customDetails {
		sendEvent.Payload.CustomDetails[k] = v
	}

	resp, err := c.Send(sendEvent)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(string(respBody))
	return nil
}

func RunRetryCommand(config *Config, routingKey string) error {
	c, _ := config.Client()

	resp, err := c.QueueRetry(routingKey)
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

func RunServerCommand() error {
	address := viper.GetString("address")
	database := viper.GetString("database")
	pidfile := viper.GetString("pidfile")
	secret := viper.GetString("secret")

	queue := persistentqueue.NewPersistentQueue(persistentqueue.WithFile(database))

	server := server.NewServer(address, secret, pidfile, queue)
	err := server.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

func RunStatusCommand(config *Config, routingKey string) error {
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

func RunStopCommand() error {
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
