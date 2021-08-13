package sensu

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/spf13/cobra"
)

type sensuCommandInput struct {
	integrationKey string
	incidentKey    string
	checkResult    map[string]interface{}
}

var errCouldNotReadStdin = errors.New(`could not read stdin for sensu enqueue command`)
var errCheckResultNotValidJson = errors.New("could not unmarshal check result, perhaps stdin did not contain valid JSON")
var errActionNotPresent = errors.New(`could not get event action, set the "action" key`)
var errCouldNotBuildDedupKey = errors.New(`could not build incident key, set the "id" field or "client.name" and "check.name" fields`)
var errCouldNotBuildSummary = errors.New(`could not build summary, set the "check.output" field`)

var sensuToPagerDutyEventType = map[string]string{
	"resolve": "resolve",
	"create":  "trigger",
}

func NewSensuEnqueueCmd(config *cmdutil.Config) *cobra.Command {
	var cmdInput sensuCommandInput
	cmd := &cobra.Command{
		Use:   "enqueue",
		Short: "Enqueue a Sensu event to PagerDuty.",
		RunE: func(cmd *cobra.Command, args []string) error {
			stdin, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return errCouldNotReadStdin
			}

			err = json.Unmarshal(stdin, &cmdInput.checkResult)
			if err != nil {
				return errCheckResultNotValidJson
			}

			sendEvent, err := buildSendEvent(cmdInput)
			if err != nil {
				return err
			}

			return cmdutil.RunSendCommand(config, &sendEvent)
		},
	}

	cmd.Flags().StringVarP(&cmdInput.incidentKey, "incident-key", "y", "", "Incident key for correlating triggers and resolves")
	cmd.Flags().StringVarP(&cmdInput.integrationKey, "integration-key", "k", "", "PagerDuty Sensu integration Key (required)")

	cmd.MarkFlagRequired("integration-key")

	return cmd
}

func buildSendEvent(cmdInput sensuCommandInput) (eventsapi.EventV1, error) {
	eventAction, err := getEventAction(cmdInput)
	if err != nil {
		return eventsapi.EventV1{}, err
	}

	dedupKey, err := buildDedupKey(cmdInput)
	if err != nil {
		return eventsapi.EventV1{}, err
	}

	summary, err := buildSummary(dedupKey, cmdInput)
	if err != nil {
		return eventsapi.EventV1{}, err
	}

	sendEvent := eventsapi.EventV1{
		ServiceKey:  cmdInput.integrationKey,
		EventType:   eventAction,
		IncidentKey: dedupKey,
		Description: summary,
		Details:     cmdInput.checkResult,
	}

	return sendEvent, nil
}

func getEventAction(cmdInput sensuCommandInput) (string, error) {
	if action, isActionPresent := cmdutil.GetNestedStringField(cmdInput.checkResult, "action"); isActionPresent {
		if pagerDutyEventAction, isActionPresent := sensuToPagerDutyEventType[action]; isActionPresent {
			return pagerDutyEventAction, nil
		}
		return sensuToPagerDutyEventType["create"], nil
	}
	return "", errActionNotPresent
}

func buildDedupKey(cmdInput sensuCommandInput) (string, error) {
	if cmdInput.incidentKey != "" {
		return cmdInput.incidentKey, nil
	}

	clientName, isClientNameString := cmdutil.GetNestedStringField(cmdInput.checkResult, "client.name")
	checkName, isCheckNameString := cmdutil.GetNestedStringField(cmdInput.checkResult, "check.name")

	if isClientNameString && isCheckNameString {
		return fmt.Sprintf("%v/%v", clientName, checkName), nil
	}

	if id, isIdPresent := cmdutil.GetNestedStringField(cmdInput.checkResult, "id"); isIdPresent {
		return id, nil
	}

	return "", errCouldNotBuildDedupKey
}

func buildSummary(dedupKey string, cmdInput sensuCommandInput) (string, error) {
	if output, outputPresent := cmdutil.GetNestedStringField(cmdInput.checkResult, "check.output"); outputPresent {
		return fmt.Sprintf("%v : %v", dedupKey, output), nil
	}

	return "", errCouldNotBuildSummary
}