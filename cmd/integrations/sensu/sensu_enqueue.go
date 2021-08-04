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

var errCouldNotReadStdin = errors.New("could not read stdin for sensu enqueue command")
var errCheckResultNotValidJson = errors.New("could not unmarshal check result, perhaps stdin did not contain valid JSON")

var sensuToPagerDutyEventType = map[string]string{
	"resolve": "resolve",
	"create":  "trigger",
}

const sensuInegrationSource = "SET_BY_PAGERDUTY"
const sensuIntegraionSeverity = "error" // This will be set later during the PagerDuty event transform

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

			err = validateSensuSendCommand(cmdInput)
			if err != nil {
				return err
			}

			sendEvent := buildSendEvent(cmdInput)

			return cmdutil.RunSendCommand(config, &sendEvent)
		},
	}

	cmd.Flags().StringVarP(&cmdInput.incidentKey, "incident-key", "y", "", "Incident key for correlating triggers and resolves")
	cmd.Flags().StringVarP(&cmdInput.integrationKey, "integration-key", "k", "", "PagerDuty Sensu integration Key (required)")

	cmd.MarkFlagRequired("integration-key")

	return cmd
}

func validateSensuSendCommand(cmdInput sensuCommandInput) error {
	return nil
}

func buildSendEvent(cmdInput sensuCommandInput) eventsapi.EventV2 {
	dedupKey := buildDedupKey(cmdInput)
	sendEvent := eventsapi.EventV2{
		RoutingKey:  cmdInput.integrationKey,
		EventAction: getEventAction(cmdInput.checkResult["action"]),
		DedupKey:    dedupKey,
		Payload: eventsapi.PayloadV2{
			Summary:       buildSummary(dedupKey, cmdInput),
			Source:        sensuInegrationSource,
			Severity:      sensuIntegraionSeverity,
			CustomDetails: cmdInput.checkResult,
		},
	}

	return sendEvent
}

func getEventAction(action interface{}) string {
	if pagerDutyEventAction, ok := sensuToPagerDutyEventType[action.(string)]; ok {
		return pagerDutyEventAction
	}
	return sensuToPagerDutyEventType["create"]
}

func buildDedupKey(cmdInput sensuCommandInput) string {
	if cmdInput.incidentKey != "" {
		return cmdInput.incidentKey
	}

	client, clientExists := cmdInput.checkResult["client"]
	check, checkExists := cmdInput.checkResult["check"]
	if clientExists && checkExists {
		clientName := client.(map[string]interface{})["name"].(string)
		checkName := check.(map[string]interface{})["name"].(string)
		return fmt.Sprintf("%v/%v", clientName, checkName)
	}

	return cmdInput.checkResult["id"].(string)
}

func buildSummary(dedupKey string, cmdInput sensuCommandInput) string {
	checkOutput := cmdInput.checkResult["check"].(map[string]interface{})["output"].(string)
	return fmt.Sprintf("%v : %v", dedupKey, checkOutput)
}
