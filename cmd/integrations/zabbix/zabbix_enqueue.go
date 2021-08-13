package zabbix

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/spf13/cobra"
)

type zabbixCommandInput struct {
	integrationKey string
	messageType    string
	details        map[string]string
}

var errCouldNotBuildDedupKey = errors.New(`could not build dedupKey, ensure event contains "incident_key", or "id" and "hostname"`)
var errCouldNotBuildSummary = errors.New(`could not build summary, ensure event contains "name", "status", and "hostname"`)

func NewZabbixEnqueueCmd(config *cmdutil.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enqueue",
		Short: "Enqueue a Zabbix event to PagerDuty.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdInput := parseArgs(args)
			sendEvent, err := buildSendEvent(cmdInput)
			if err != nil {
				return err
			}

			return cmdutil.RunSendCommand(config, &sendEvent)
		},
	}

	return cmd
}

func buildSendEvent(cmdInput zabbixCommandInput) (eventsapi.EventV1, error) {
	dedupKey, err := buildDedupKey(cmdInput)
	if err != nil {
		return eventsapi.EventV1{}, errCouldNotBuildDedupKey
	}

	summary, err := buildSummary(cmdInput)
	if err != nil {
		return eventsapi.EventV1{}, errCouldNotBuildSummary
	}

	client, clientUrl := getClientAndClientUrl(cmdInput.details)

	sendEvent := eventsapi.EventV1{
		ServiceKey:  cmdInput.integrationKey,
		EventType:   cmdInput.messageType,
		IncidentKey: dedupKey,
		Client:      client,
		ClientURL:   clientUrl,
		Description: summary,
		Details:     cmdutil.StringMapToInterfaceMap(cmdInput.details),
	}

	return sendEvent, nil
}

func buildDedupKey(cmdInput zabbixCommandInput) (string, error) {
	details := cmdutil.StringMapToInterfaceMap(cmdInput.details)
	if providedDedupKey, isDedupKeyProvided := cmdutil.GetNestedStringField(details, "incident_key"); isDedupKeyProvided {
		return providedDedupKey, nil
	}
	id, isIdPresent := cmdutil.GetNestedStringField(details, "id")
	hostname, isHostnamePresent := cmdutil.GetNestedStringField(details, "hostname")
	if isIdPresent && isHostnamePresent {
		return fmt.Sprintf("%v-%v", id, hostname), nil
	}
	return "", errCouldNotBuildDedupKey
}

func buildSummary(cmdInput zabbixCommandInput) (string, error) {
	details := cmdutil.StringMapToInterfaceMap(cmdInput.details)
	name, isNamePresent := cmdutil.GetNestedStringField(details, "name")
	status, isStatusPresent := cmdutil.GetNestedStringField(details, "status")
	hostname, isHostnamePresent := cmdutil.GetNestedStringField(details, "hostname")
	if isNamePresent && isStatusPresent && isHostnamePresent {
		return fmt.Sprintf("%v : %v for %v", name, status, hostname), nil
	}
	return "", errCouldNotBuildSummary
}

func parseArgs(args []string) zabbixCommandInput {
	details := parseRawDetails(args[2])
	messageType := args[1]

	note, isNotePresent := cmdutil.GetNestedStringField(cmdutil.StringMapToInterfaceMap(details), "NOTE")
	if messageType == "trigger" && isNotePresent && strings.Contains(note, "Escalation cancelled") {
		messageType = "resolve"
		details["NOTE"] += " (converted from trigger to resolve by pdagent integration)"
	}

	return zabbixCommandInput{
		integrationKey: args[0],
		messageType:    messageType,
		details:        details,
	}
}

func parseRawDetails(rawDetails string) map[string]string {
	details := map[string]string{}
	var key, val string

	trimmedDetails := strings.TrimSpace(rawDetails)
	for _, line := range strings.Split(trimmedDetails, "\n") {
		detail := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(detail) == 2 {
			key = detail[0]
			val = detail[1]
			details[key] = val
		} else if key != "" {
			details[key] += detail[0]
		} else {
			details[detail[0]] = detail[0]
		}
	}

	return details
}

func getClientAndClientUrl(details map[string]string) (string, string) {
	clientUrl, isClientUrlPresent := cmdutil.GetNestedStringField(cmdutil.StringMapToInterfaceMap(details), "url")
	var client string
	if isClientUrlPresent && clientUrl != "" {
		client = "Zabbix"
	}

	return client, clientUrl
}
