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
package zabbix

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/PagerDuty/go-pdagent/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

type zabbixCommandTestInput struct {
	integrationKey string
	messageType    string
	rawDetails     string
}

func allArgs(inputs zabbixCommandTestInput) []string {
	return []string{inputs.integrationKey, inputs.messageType, inputs.rawDetails}
}

func onlyTwoArgs(inputs zabbixCommandTestInput) []string {
	return []string{inputs.integrationKey, inputs.messageType}
}

func TestZabbixEnqueue_errors(t *testing.T) {
	tests := []struct {
		name          string
		inputs        zabbixCommandTestInput
		buildArgs     func(zabbixCommandTestInput) []string
		expectedError error
	}{
		{
			name: "missingDetailsForDedupKey",
			inputs: zabbixCommandTestInput{
				integrationKey: "zabbix_integration_key",
				messageType:    "trigger",
				rawDetails: `name:{TRIGGER.NAME}
				status:{TRIGGER.STATUS}
				ip:{HOST.IP}
				value:{TRIGGER.VALUE}
				event_id:{EVENT.ID}
				severity:{TRIGGER.SEVERITY}`,
			},
			buildArgs:     allArgs,
			expectedError: errCouldNotBuildDedupKey,
		},
		{
			name: "missingDetailsFordescription",
			inputs: zabbixCommandTestInput{
				integrationKey: "zabbix_integration_key",
				messageType:    "trigger",
				rawDetails: `id:{TRIGGER.ID}
				hostname:{HOST.NAME}
				ip:{HOST.IP}
				value:{TRIGGER.VALUE}
				event_id:{EVENT.ID}
				severity:{TRIGGER.SEVERITY}`,
			},
			buildArgs:     allArgs,
			expectedError: errCouldNotBuildSummary,
		},
		{
			name: "incorrectNumberOfArgs",
			inputs: zabbixCommandTestInput{
				integrationKey: "zabbix_integration_key",
				messageType:    "trigger",
			},
			buildArgs:     onlyTwoArgs,
			expectedError: errors.New("accepts 3 arg(s), received 2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.InitConfigForIntegrationsTesting()

			realConfig := cmdutil.NewConfig()

			cmd := NewZabbixEnqueueCmd(realConfig)
			cmd.SetArgs(tt.buildArgs(tt.inputs))

			_, err := cmd.ExecuteC()

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestZabbixEnqueue_validInputs(t *testing.T) {
	tests := []struct {
		name                 string
		cmdInputs            zabbixCommandTestInput
		expectedResponseBody map[string]interface{}
	}{
		{
			name: "basicValidInputs",
			cmdInputs: zabbixCommandTestInput{
				integrationKey: "zabbix_integration_key",
				messageType:    "trigger",
				rawDetails: `name:{TRIGGER.NAME}
				id:{TRIGGER.ID}
				status:{TRIGGER.STATUS}
				hostname:{HOST.NAME}`,
			},
			expectedResponseBody: map[string]interface{}{
				"service_key":  "zabbix_integration_key",
				"event_type":   "trigger",
				"incident_key": "{TRIGGER.ID}-{HOST.NAME}",
				"description":  "{TRIGGER.NAME} : {TRIGGER.STATUS} for {HOST.NAME}",
				"details": map[string]interface{}{
					"name":     "{TRIGGER.NAME}",
					"id":       "{TRIGGER.ID}",
					"status":   "{TRIGGER.STATUS}",
					"hostname": "{HOST.NAME}",
				},
				"agent": map[string]interface{}{
					"agent_id":  common.UserAgent(),
					"queued_by": "pd-zabbix",
					"queued_at": "2021-01-01T00:00:00Z",
				},
			},
		},
		{
			name: "providedDedupKey",
			cmdInputs: zabbixCommandTestInput{
				integrationKey: "zabbix_integration_key",
				messageType:    "resolve",
				rawDetails: `name:{TRIGGER.NAME}
				incident_key:provided_incident_key
				status:{TRIGGER.STATUS}
				hostname:{HOST.NAME}`,
			},
			expectedResponseBody: map[string]interface{}{
				"service_key":  "zabbix_integration_key",
				"event_type":   "resolve",
				"incident_key": "provided_incident_key",
				"description":  "{TRIGGER.NAME} : {TRIGGER.STATUS} for {HOST.NAME}",
				"details": map[string]interface{}{
					"name":         "{TRIGGER.NAME}",
					"incident_key": "provided_incident_key",
					"status":       "{TRIGGER.STATUS}",
					"hostname":     "{HOST.NAME}",
				},
				"agent": map[string]interface{}{
					"agent_id":  common.UserAgent(),
					"queued_by": "pd-zabbix",
					"queued_at": "2021-01-01T00:00:00Z",
				},
			},
		},
		{
			name: "resolvingIncidentWithNote",
			cmdInputs: zabbixCommandTestInput{
				integrationKey: "zabbix_integration_key",
				messageType:    "trigger",
				rawDetails: `name:{TRIGGER.NAME}
				id:{TRIGGER.ID}
				status:{TRIGGER.STATUS}
				hostname:{HOST.NAME}
				NOTE:Escalation cancelled`,
			},
			expectedResponseBody: map[string]interface{}{
				"service_key":  "zabbix_integration_key",
				"event_type":   "resolve",
				"incident_key": "{TRIGGER.ID}-{HOST.NAME}",
				"description":  "{TRIGGER.NAME} : {TRIGGER.STATUS} for {HOST.NAME}",
				"details": map[string]interface{}{
					"name":     "{TRIGGER.NAME}",
					"id":       "{TRIGGER.ID}",
					"status":   "{TRIGGER.STATUS}",
					"hostname": "{HOST.NAME}",
					"NOTE":     "Escalation cancelled (converted from trigger to resolve by pdagent integration)",
				},
				"agent": map[string]interface{}{
					"agent_id":  common.UserAgent(),
					"queued_by": "pd-zabbix",
					"queued_at": "2021-01-01T00:00:00Z",
				},
			},
		},
		{
			name: "setClientAndClientUrl",
			cmdInputs: zabbixCommandTestInput{
				integrationKey: "zabbix_integration_key",
				messageType:    "trigger",
				rawDetails: `name:{TRIGGER.NAME}
				id:{TRIGGER.ID}
				status:{TRIGGER.STATUS}
				hostname:{HOST.NAME}
				url:some.url`,
			},
			expectedResponseBody: map[string]interface{}{
				"service_key":  "zabbix_integration_key",
				"event_type":   "trigger",
				"incident_key": "{TRIGGER.ID}-{HOST.NAME}",
				"client":       "Zabbix",
				"client_url":   "some.url",
				"description":  "{TRIGGER.NAME} : {TRIGGER.STATUS} for {HOST.NAME}",
				"details": map[string]interface{}{
					"name":     "{TRIGGER.NAME}",
					"id":       "{TRIGGER.ID}",
					"status":   "{TRIGGER.STATUS}",
					"hostname": "{HOST.NAME}",
					"url":      "some.url",
				},
				"agent": map[string]interface{}{
					"agent_id":  common.UserAgent(),
					"queued_by": "pd-zabbix",
					"queued_at": "2021-01-01T00:00:00Z",
				},
			},
		},
		{
			name: "errantNewLinesInDetails",
			cmdInputs: zabbixCommandTestInput{
				integrationKey: "zabbix_integration_key",
				messageType:    "trigger",
				rawDetails: `someErrantKeyHere
				name:{TRIGGER.NAME}
				id:{TRIGGER.ID}
				status:{TRIGGER.STATUS}
				hostname:{HOST.
					NAME}`,
			},
			expectedResponseBody: map[string]interface{}{
				"service_key":  "zabbix_integration_key",
				"event_type":   "trigger",
				"incident_key": "{TRIGGER.ID}-{HOST.NAME}",
				"description":  "{TRIGGER.NAME} : {TRIGGER.STATUS} for {HOST.NAME}",
				"details": map[string]interface{}{
					"name":              "{TRIGGER.NAME}",
					"id":                "{TRIGGER.ID}",
					"status":            "{TRIGGER.STATUS}",
					"hostname":          "{HOST.NAME}",
					"someErrantKeyHere": "someErrantKeyHere",
				},
				"agent": map[string]interface{}{
					"agent_id":  common.UserAgent(),
					"queued_by": "pd-zabbix",
					"queued_at": "2021-01-01T00:00:00Z",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock = test.TestClock{}
			test.InitConfigForIntegrationsTesting()

			defer gock.Off()

			defaultHTTPClient := &http.Client{
				Timeout: 5 * time.Minute,
			}

			realConfig := cmdutil.NewConfig()
			realConfig.HttpClient = func() (*http.Client, error) {
				return defaultHTTPClient, nil
			}

			cmd := NewZabbixEnqueueCmd(realConfig)
			cmd.SetArgs(allArgs(tt.cmdInputs))

			gock.New(cmdutil.GetDefaults().Address).
				Post("/send").JSON(tt.expectedResponseBody).
				Reply(200).JSON(map[string]interface{}{"key": tt.cmdInputs.integrationKey})

			gock.InterceptClient(defaultHTTPClient)

			out, err := test.CaptureStdout(func() error {
				_, err := cmd.ExecuteC()
				return err
			})

			if err != nil {
				t.Errorf("error running command `enqueue`: %v", err)
			}

			assert.Contains(t, out, fmt.Sprintf(`{"key":"%v"}`, tt.cmdInputs.integrationKey))
		})
	}
}
