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
package sensu

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/PagerDuty/go-pdagent/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func buildCmdArgs(inputs sensuCommandInput) []string {
	args := []string{}
	flags := []struct {
		flag string
		val  string
	}{
		{"-k", inputs.integrationKey}, {"-y", inputs.incidentKey},
	}
	for _, f := range flags {
		if f.val != "" {
			args = append(args, f.flag, f.val)
		}
	}
	return args
}

func TestSensuEnqueue_errors(t *testing.T) {
	tests := []struct {
		name          string
		inputs        sensuCommandInput
		expectedError error
	}{
		{
			name: "missingRequiredFlags",
			inputs: sensuCommandInput{
				checkResult: map[string]interface{}{},
			},
			expectedError: errors.New("required flag(s) \"integration-key\" not set"),
		},
		{
			name: "actionNotPresent",
			inputs: sensuCommandInput{
				integrationKey: "integration_key",
				checkResult: map[string]interface{}{
					"not_action": "not_action",
				},
			},
			expectedError: errActionNotPresent,
		},
		{
			name: "actionNotString",
			inputs: sensuCommandInput{
				integrationKey: "integration_key",
				checkResult:    map[string]interface{}{"action": true},
			},
			expectedError: errActionNotPresent,
		},
		{
			name: "clientCheckAndIdNotProvided",
			inputs: sensuCommandInput{
				integrationKey: "integration_key",
				checkResult: map[string]interface{}{
					"action": "action",
				},
			},
			expectedError: errCouldNotBuildDedupKey,
		},
		{
			name: "clientAndCheckNotMaps",
			inputs: sensuCommandInput{
				integrationKey: "integration_key",
				checkResult: map[string]interface{}{
					"action": "action",
					"client": "client",
					"check":  "check",
				},
			},
			expectedError: errCouldNotBuildDedupKey,
		},
		{
			name: "clientAndCheckDoNotHaveName",
			inputs: sensuCommandInput{
				integrationKey: "integration_key",
				checkResult: map[string]interface{}{
					"action": "action",
					"client": map[string]interface{}{},
					"check":  map[string]interface{}{},
				},
			},
			expectedError: errCouldNotBuildDedupKey,
		},
		{
			name: "clientAndCheckNameNotString",
			inputs: sensuCommandInput{
				integrationKey: "integration_key",
				checkResult: map[string]interface{}{
					"action": "action",
					"client": map[string]interface{}{"name": true},
					"check":  map[string]interface{}{"name": true},
				},
			},
			expectedError: errCouldNotBuildDedupKey,
		},
		{
			name: "outputNotPresent",
			inputs: sensuCommandInput{
				integrationKey: "integration_key",
				checkResult: map[string]interface{}{
					"action": "action",
					"client": map[string]interface{}{"name": "name"},
					"check":  map[string]interface{}{"name": "name"},
				},
			},
			expectedError: errCouldNotBuildSummary,
		},
		{
			name: "outputNotString",
			inputs: sensuCommandInput{
				integrationKey: "integration_key",
				checkResult: map[string]interface{}{
					"action": "action",
					"client": map[string]interface{}{"name": "name"},
					"check": map[string]interface{}{
						"name":   "name",
						"output": true,
					},
				},
			},
			expectedError: errCouldNotBuildSummary,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.InitConfigForIntegrationsTesting()

			realConfig := cmdutil.NewConfig()

			cmd := NewSensuEnqueueCmd(realConfig)
			cmd.SetArgs(buildCmdArgs(tt.inputs))

			json, _ := json.Marshal(tt.inputs.checkResult)

			tmpfile, err := ioutil.TempFile("", "stdin")
			if err != nil {
				log.Fatal(err)
			}

			defer os.Remove(tmpfile.Name()) // clean up

			if _, err := tmpfile.Write(json); err != nil {
				log.Fatal(err)
			}

			if _, err := tmpfile.Seek(0, 0); err != nil {
				log.Fatal(err)
			}

			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }() // Restore original Stdin

			os.Stdin = tmpfile

			_, err = cmd.ExecuteC()

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)

			if err := tmpfile.Close(); err != nil {
				log.Fatal(err)
			}
		})
	}
}

func TestSensuEnqueue_validInputs(t *testing.T) {
	tests := []struct {
		name                 string
		cmdInputs            sensuCommandInput
		expectedResponseBody map[string]interface{}
	}{
		{
			name: "buildDedupKeyFromClientAndCheck",
			cmdInputs: sensuCommandInput{
				integrationKey: "sensu_integration_key",
				checkResult: map[string]interface{}{
					"action": "action",
					"check": map[string]interface{}{
						"name":   "checkname",
						"output": "output",
					},
					"client": map[string]interface{}{"name": "clientname"},
				},
			},
			expectedResponseBody: map[string]interface{}{
				"service_key":  "sensu_integration_key",
				"event_type":   "trigger",
				"incident_key": "clientname/checkname",
				"description":  "clientname/checkname : output",
				"details": map[string]interface{}{
					"action": "action",
					"check": map[string]interface{}{
						"name":   "checkname",
						"output": "output",
					},
					"client": map[string]interface{}{"name": "clientname"},
				},
			},
		},
		{
			name: "buildDedupKeyFromId",
			cmdInputs: sensuCommandInput{
				integrationKey: "sensu_integration_key",
				checkResult: map[string]interface{}{
					"action": "action",
					"check":  map[string]interface{}{"output": "output"},
					"id":     "some_id",
				},
			},
			expectedResponseBody: map[string]interface{}{
				"service_key":  "sensu_integration_key",
				"event_type":   "trigger",
				"incident_key": "some_id",
				"description":  "some_id : output",
				"details": map[string]interface{}{
					"action": "action",
					"check":  map[string]interface{}{"output": "output"},
					"id":     "some_id",
				},
			},
		},
		{
			name: "userProvidedDedupKey",
			cmdInputs: sensuCommandInput{
				integrationKey: "sensu_integration_key",
				checkResult: map[string]interface{}{
					"action": "action",
					"check":  map[string]interface{}{"output": "output"},
				},
				incidentKey: "userProvidedDedupKey",
			},
			expectedResponseBody: map[string]interface{}{
				"service_key":  "sensu_integration_key",
				"event_type":   "trigger",
				"incident_key": "userProvidedDedupKey",
				"description":  "userProvidedDedupKey : output",
				"details": map[string]interface{}{
					"action": "action",
					"check":  map[string]interface{}{"output": "output"},
				},
			},
		},
		{
			name: "createAction",
			cmdInputs: sensuCommandInput{
				integrationKey: "sensu_integration_key",
				checkResult: map[string]interface{}{
					"action": "create",
					"check":  map[string]interface{}{"output": "output"},
				},
				incidentKey: "userProvidedDedupKey",
			},
			expectedResponseBody: map[string]interface{}{
				"service_key":  "sensu_integration_key",
				"event_type":   "trigger",
				"incident_key": "userProvidedDedupKey",
				"description":  "userProvidedDedupKey : output",
				"details": map[string]interface{}{
					"action": "create",
					"check":  map[string]interface{}{"output": "output"},
				},
			},
		},
		{
			name: "resolveAction",
			cmdInputs: sensuCommandInput{
				integrationKey: "sensu_integration_key",
				checkResult: map[string]interface{}{
					"action": "resolve",
					"check":  map[string]interface{}{"output": "output"},
				},
				incidentKey: "userProvidedDedupKey",
			},
			expectedResponseBody: map[string]interface{}{
				"service_key":  "sensu_integration_key",
				"event_type":   "resolve",
				"incident_key": "userProvidedDedupKey",
				"description":  "userProvidedDedupKey : output",
				"details": map[string]interface{}{
					"action": "resolve",
					"check":  map[string]interface{}{"output": "output"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.InitConfigForIntegrationsTesting()

			defer gock.Off()

			defaultHTTPClient := &http.Client{
				Timeout: 5 * time.Minute,
			}

			realConfig := cmdutil.NewConfig()
			realConfig.HttpClient = func() (*http.Client, error) {
				return defaultHTTPClient, nil
			}

			cmd := NewSensuEnqueueCmd(realConfig)
			cmd.SetArgs(buildCmdArgs(tt.cmdInputs))

			gock.New(cmdutil.GetDefaults().Address).
				Post("/send").JSON(tt.expectedResponseBody).
				Reply(200).JSON(map[string]interface{}{"key": tt.cmdInputs.integrationKey})

			gock.InterceptClient(defaultHTTPClient)

			json, _ := json.Marshal(tt.cmdInputs.checkResult)

			tmpfile, err := ioutil.TempFile("", "stdin")
			if err != nil {
				log.Fatal(err)
			}

			defer os.Remove(tmpfile.Name()) // clean up

			if _, err := tmpfile.Write(json); err != nil {
				log.Fatal(err)
			}

			if _, err := tmpfile.Seek(0, 0); err != nil {
				log.Fatal(err)
			}

			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }() // Restore original Stdin

			os.Stdin = tmpfile

			out, err := test.CaptureStdout(func() error {
				_, err := cmd.ExecuteC()
				return err
			})

			if err != nil {
				t.Errorf("error running command `enqueue`: %v", err)
			}

			assert.Contains(t, out, fmt.Sprintf(`{"key":"%v"}`, tt.cmdInputs.integrationKey))

			if err := tmpfile.Close(); err != nil {
				log.Fatal(err)
			}
		})
	}
}
