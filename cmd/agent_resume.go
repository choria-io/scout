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

type agentResumeCommand struct {
	classes []string
	ids     []string
	checks  []string
	json    bool
}

func configureAgentResumeCommand(app *kingpin.CmdClause) {
	c := &agentResumeCommand{
		classes: []string{},
		ids:     []string{},
		checks:  []string{},
		json:    false,
	}

	resume := app.Command("resume", "Set checks to resume regular checks").Action(c.resume)
	resume.Arg("checks", "The checks to put into resume, empty for all").StringsVar(&c.checks)
	resume.Flag("tags", "Limit to entities with these tags").StringsVar(&c.classes)
	resume.Flag("ids", "Limit to entities with these identities").StringsVar(&c.ids)
	resume.Flag("json", "Produce JSON output").Short('j').BoolVar(&c.json)

}

func (c *agentResumeCommand) showJSON(r map[string]scoutagent.ResumeReply) error {
	j, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	return nil
}

func (c *agentResumeCommand) showTable(r map[string]scoutagent.ResumeReply) error {
	table := newTable()
	table.SetHeader([]string{"Entity", "Transitioned", "Failed", "Skipped"})

	for n, result := range r {
		table.Append([]string{n, strings.Join(result.TransitionedChecks, ", "), strings.Join(result.FailedChecks, ", "), strings.Join(result.SkippedChecks, ", ")})
	}

	table.Render()

	return nil
}

func (c *agentResumeCommand) resume(_ *kingpin.ParseContext) error {
	defer wg.Done()

	err := commonConfigure()
	if err != nil {
		return err
	}

	api, err := scoutapi.NewAPIClient(cfile, log)
	if err != nil {
		return err
	}

	responses := make(map[string]scoutagent.ResumeReply, 0)

	stat, err := api.ResumeChecks(ctx, c.checks, c.ids, []string{}, c.classes, func(r *scoutclient.ResumeOutput) {
		if !r.ResultDetails().OK() {
			log.Errorf("Failed response received from %s: %s", r.ResultDetails().Sender(), r.ResultDetails().StatusMessage())
			return
		}

		result := scoutagent.ResumeReply{}
		err = r.ParseResumeOutput(&result)
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
