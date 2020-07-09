package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	scoutagent "github.com/choria-io/go-choria/scout/agent/scout"
	scoutapi "github.com/choria-io/go-choria/scout/client"
	scoutclient "github.com/choria-io/go-choria/scout/client/scout"
	"gopkg.in/alecthomas/kingpin.v2"
)

type agentTriggerCommand struct {
	classes []string
	ids     []string
	checks  []string
	json    bool
}

func configureAgentTriggerCommand(app *kingpin.CmdClause) {
	c := &agentTriggerCommand{
		classes: []string{},
		ids:     []string{},
		checks:  []string{},
		json:    false,
	}

	trigger := app.Command("trigger", "Triggers immediate check validation").Action(c.trigger)
	trigger.Arg("checks", "The checks to trigger, empty for all").StringsVar(&c.checks)
	trigger.Flag("tags", "Limit to entities with these tags").StringsVar(&c.classes)
	trigger.Flag("ids", "Limit to entities with these identities").StringsVar(&c.ids)
	trigger.Flag("json", "Produce JSON output").Short('j').BoolVar(&c.json)

}

func (c *agentTriggerCommand) showJSON(r map[string]scoutagent.TriggerReply) error {
	j, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	return nil
}

func (c *agentTriggerCommand) showTable(r map[string]scoutagent.TriggerReply) error {
	table := newTable()
	table.SetHeader([]string{"Entity", "Transitioned", "Failed", "Skipped"})

	for n, result := range r {
		table.Append([]string{n, strings.Join(result.TransitionedChecks, ", "), strings.Join(result.FailedChecks, ", "), strings.Join(result.SkippedChecks, ", ")})
	}

	table.Render()

	return nil
}

func (c *agentTriggerCommand) trigger(_ *kingpin.ParseContext) error {
	defer wg.Done()

	err := commonConfigure()
	if err != nil {
		return err
	}

	api, err := scoutapi.NewAPIClient(cfile, log)
	if err != nil {
		return err
	}

	responses := make(map[string]scoutagent.TriggerReply, 0)

	stat, err := api.TriggerChecks(ctx, c.checks, c.ids, []string{}, c.classes, func(r *scoutclient.TriggerOutput) {
		if !r.ResultDetails().OK() {
			log.Errorf("Failed response received from %s: %s", r.ResultDetails().Sender(), r.ResultDetails().StatusMessage())
			return
		}

		result := scoutagent.TriggerReply{}
		err = r.ParseTriggerOutput(&result)
		if err != nil {
			log.Errorf("Could not parse reply from %s: %s", r.ResultDetails().Sender(), err)
			return
		}

		responses[r.ResultDetails().Sender()] = result
	})
	if err != nil {
		return err
	}

	if stat.ResponsesCount() == 0 {
		return fmt.Errorf("no responses received")
	}

	if c.json {
		return c.showJSON(responses)
	}

	return c.showTable(responses)
}
