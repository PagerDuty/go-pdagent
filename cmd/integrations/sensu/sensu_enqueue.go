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
var errActionNotPresent = errors.New(`check result must contain and "action" field`)
var errActionValueNotValid = errors.New(`check result "action" must be "resolve" or "create"`)
var errActionMustBeAString = errors.New(`key "action" must be of type string`)
var errCouldNotBuildDedupKey = errors.New(`could not build incident key, set the "id" field or "client.name" and "check.name" fields`)
var errCouldNotBuildSummary = errors.New(`could not build summary, set the "check.output" field`)

var sensuToPagerDutyEventType = map[string]string{
	"resolve": "resolve",
	"create":  "trigger",
}

// PagerDuty will set these fields during event transformation
const sensuInegrationSource = "SET_BY_PAGERDUTY"
const sensuIntegraionSeverity = "error"

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

			return cmdutil.RunSendCommand(config, &sendEvent)
		},
	}

	cmd.Flags().StringVarP(&cmdInput.incidentKey, "incident-key", "y", "", "Incident key for correlating triggers and resolves")
	cmd.Flags().StringVarP(&cmdInput.integrationKey, "integration-key", "k", "", "PagerDuty Sensu integration Key (required)")

	cmd.MarkFlagRequired("integration-key")

	return cmd
}

func buildSendEvent(cmdInput sensuCommandInput) (eventsapi.EventV2, error) {
	dedupKey, err := buildDedupKey(cmdInput)
	if err != nil {
		return eventsapi.EventV2{}, err
	}

	eventAction, err := getEventAction(cmdInput)
	if err != nil {
		return eventsapi.EventV2{}, err
	}

	summary, err := buildSummary(dedupKey, cmdInput)
	if err != nil {
		return eventsapi.EventV2{}, err
	}

	sendEvent := eventsapi.EventV2{
		RoutingKey:  cmdInput.integrationKey,
		EventAction: eventAction,
		DedupKey:    dedupKey,
		Payload: eventsapi.PayloadV2{
			Summary:       summary,
			Source:        sensuInegrationSource,
			Severity:      sensuIntegraionSeverity,
			CustomDetails: cmdInput.checkResult,
		},
	}

	return sendEvent, nil
}

func getEventAction(cmdInput sensuCommandInput) (string, error) {
	action, actionPresent := cmdInput.checkResult["action"]
	if !actionPresent {
		return "", errActionNotPresent
	}
	actionString, isActionString := action.(string)
	if !isActionString {
		return "", errActionMustBeAString
	}
	allowedActions := []string{"create", "resolve"}
	err := cmdutil.ValidateEnumField(actionString, allowedActions, errActionValueNotValid)
	if err != nil {
		return "", err
	}

	if pagerDutyEventAction, ok := sensuToPagerDutyEventType[actionString]; ok {
		return pagerDutyEventAction, nil
	}

	return sensuToPagerDutyEventType["create"], nil
}

func buildDedupKey(cmdInput sensuCommandInput) (string, error) {
	if cmdInput.incidentKey != "" {
		return cmdInput.incidentKey, nil
	}

	client, clientExists := cmdInput.checkResult["client"]
	check, checkExists := cmdInput.checkResult["check"]
	if clientExists && checkExists {
		clientMap, isClientMap := client.(map[string]interface{})
		checkMap, isCheckMap := check.(map[string]interface{})
		if isClientMap && isCheckMap {
			clientName, clientNamePresent := clientMap["name"]
			checkName, checkNamePresent := checkMap["name"]
			if clientNamePresent && checkNamePresent {
				clientNameString, isClientNameString := clientName.(string)
				checkNameString, isCheckNameString := checkName.(string)
				if isCheckNameString && isClientNameString {
					return fmt.Sprintf("%v/%v", clientNameString, checkNameString), nil
				}
			}
		}
	}

	id, idPresent := cmdInput.checkResult["id"]
	if idPresent {
		idString, isIdString := id.(string)
		if isIdString {
			return idString, nil
		}
	}

	return "", errCouldNotBuildDedupKey
}

func buildSummary(dedupKey string, cmdInput sensuCommandInput) (string, error) {
	// The check field was validated in `buildDedupKey`
	checkOutput, checkOutputPresent := cmdInput.checkResult["check"].(map[string]interface{})["output"]
	if checkOutputPresent {
		checkOutputString, isCheckOutputString := checkOutput.(string)
		if isCheckOutputString {
			return fmt.Sprintf("%v : %v", dedupKey, checkOutputString), nil
		}
	}

	return "", errCouldNotBuildSummary
}
