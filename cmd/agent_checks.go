package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	scoutagent "github.com/choria-io/go-choria/scout/agent/scout"
	scoutapi "github.com/choria-io/go-choria/scout/client"
	"gopkg.in/alecthomas/kingpin.v2"
)

type agentChecksCommand struct {
	id   string
	json bool
}

func configureAgentChecksCommand(app *kingpin.CmdClause) {
	c := &agentChecksCommand{}

	checks := app.Command("checks", "Retrieve check statuses from an agent").Action(c.checks)
	checks.Arg("identity", "The entity identity to query").Required().StringVar(&c.id)
	checks.Flag("json", "Produce JSON output").Short('j').BoolVar(&c.json)
}

func (c *agentChecksCommand) showJSON(checks []*scoutagent.CheckState) error {
	j, err := json.MarshalIndent(checks, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	return nil
}

func (c *agentChecksCommand) showTable(checks []*scoutagent.CheckState) error {
	table := newTable()
	table.SetHeader([]string{"Name", "Status", "Start Time"})

	for _, check := range checks {
		table.Append([]string{check.Name, check.State, time.Unix(check.Started, 0).Format(time.RFC1123)})
	}

	table.Render()

	return nil
}

func (c *agentChecksCommand) checks(_ *kingpin.ParseContext) error {
	defer wg.Done()

	err := commonConfigure()
	if err != nil {
		return err
	}

	api, err := scoutapi.NewAPIClient(cfile, log)
	if err != nil {
		return err
	}

	checks, err := api.EntityChecks(ctx, c.id)
	if err != nil {
		return err
	}

	if c.json {
		return c.showJSON(checks)
	}

	return c.showTable(checks)
}
