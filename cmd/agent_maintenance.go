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

type agentMaintCommand struct {
	classes []string
	ids     []string
	checks  []string
	json    bool
}

func configureAgentMaintCommand(app *kingpin.CmdClause) {
	c := &agentMaintCommand{
		classes: []string{},
		ids:     []string{},
		checks:  []string{},
		json:    false,
	}

	maint := app.Command("maintenance", "Set checks to maintenance mode").Alias("maint").Action(c.maint)
	maint.Arg("checks", "The checks to put into maintenance, empty for all").StringsVar(&c.checks)
	maint.Flag("tags", "Limit to entities with these tags").StringsVar(&c.classes)
	maint.Flag("ids", "Limit to entities with these identities").StringsVar(&c.ids)
	maint.Flag("json", "Produce JSON output").Short('j').BoolVar(&c.json)

}

func (c *agentMaintCommand) showJSON(r map[string]scoutagent.MaintenanceReply) error {
	j, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	return nil
}

func (c *agentMaintCommand) showTable(r map[string]scoutagent.MaintenanceReply) error {
	table := newTable()
	table.SetHeader([]string{"Entity", "Transitioned", "Failed", "Skipped"})

	for n, result := range r {
		table.Append([]string{n, strings.Join(result.TransitionedChecks, ", "), strings.Join(result.FailedChecks, ", "), strings.Join(result.SkippedChecks, ", ")})
	}

	table.Render()

	return nil
}

func (c *agentMaintCommand) maint(_ *kingpin.ParseContext) error {
	defer wg.Done()

	err := commonConfigure()
	if err != nil {
		return err
	}

	api, err := scoutapi.NewAPIClient(cfile, log)
	if err != nil {
		return err
	}

	responses := make(map[string]scoutagent.MaintenanceReply, 0)

	stat, err := api.PauseChecks(ctx, c.checks, c.ids, []string{}, c.classes, func(r *scoutclient.MaintenanceOutput) {
		if !r.ResultDetails().OK() {
			log.Errorf("Failed response received from %s: %s", r.ResultDetails().Sender(), r.ResultDetails().StatusMessage())
			return
		}

		result := scoutagent.MaintenanceReply{}
		err = r.ParseMaintenanceOutput(&result)
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
