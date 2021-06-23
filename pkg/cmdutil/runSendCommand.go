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

	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
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
