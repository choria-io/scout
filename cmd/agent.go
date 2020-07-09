package cmd

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

func configureAgentCommand(app *kingpin.Application) {
	agent := app.Command("agent", "Monitoring Agent")

	configureAgentRunCommand(agent)
	configureAgentChecksCommand(agent)
	configureAgentMaintCommand(agent)
	configureAgentResumeCommand(agent)
	configureAgentTriggerCommand(agent)
}
