package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/choria-io/go-choria/choria"
	"github.com/choria-io/go-choria/config"
	"github.com/choria-io/go-choria/scout"
	"github.com/choria-io/go-choria/server"
	"gopkg.in/alecthomas/kingpin.v2"
)

type agentRunCommand struct {
	jwt     string
	pidfile string
	clean   bool
}

func configureAgentRunCommand(app *kingpin.CmdClause) {
	c := &agentRunCommand{}

	run := app.Command("run", "Runs the Scout Server").Action(c.run)
	run.Flag("provision", "Path to the provisioning JWT file").StringVar(&c.jwt)
	run.Flag("clean", "Removes checks and overrides at startup").BoolVar(&c.clean)
	run.Flag("pid", "Write running PID to a file").StringVar(&c.pidfile)
}

func (c *agentRunCommand) run(_ *kingpin.ParseContext) error {
	defer wg.Done()

	err := c.configure()
	if err != nil {
		return err
	}

	log.Infof("Choria Scout Agent version %s starting with configuration %s", bi.Version(), cfg.ConfigFile)

	if c.pidfile != "" {
		err := ioutil.WriteFile(c.pidfile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
		if err != nil {
			return fmt.Errorf("could not write PID: %s", err)
		}
	}

	instance, err := server.NewInstance(fw)
	if err != nil {
		return err
	}

	switch fw.ProvisionMode() {
	case true:
		// TODO: needs provisioner updates
		// instance.SetComponent("provision_mode_scout")
	case false:
		instance.DenyAgent("rpcutil")
		instance.SetComponent("scout")
	}

	// prevent machines from starting till we are ready
	configuredMachineDir := cfg.Choria.MachineSourceDir
	cfg.Choria.MachineSourceDir = ""

	wg.Add(1)
	err = instance.Run(ctx, &wg)
	if err != nil {
		return err
	}

	if !fw.ProvisionMode() {
		cfg.Choria.MachineSourceDir = configuredMachineDir

		scoutEntity, err := scout.New(fw)
		if err != nil {
			return err
		}

		tags, err := scoutEntity.Tags()
		if err != nil {
			return err
		}

		// write tags as classes file
		tf, err := ioutil.TempFile("", "")
		if err != nil {
			return err
		}
		defer os.Remove(tf.Name())

		cfg.ClassesFile = tf.Name()
		fmt.Fprint(tf, strings.Join(tags, "\n"))
		tf.Close()

		err = scoutEntity.Start(ctx, &wg, c.clean)
		if err != nil {
			return err
		}

		err = instance.StartMachine(ctx, &wg)
		if err != nil {
			return err
		}

	} else {
		log.Warnf("Scout monitoring not started during provision mode")
	}

	<-ctx.Done()

	return nil
}

func (c *agentRunCommand) configure() error {
	if c.jwt != "" {
		bi.SetProvisionJWTFile(c.jwt)
	}

	switch {
	case choria.FileExist(cfile):
		cfg, err = config.NewConfig(cfile)
		if err != nil {
			return err
		}

	case bi.ProvisionJWTFile() != "":
		cfg, err = config.NewDefaultConfig()
		if err != nil {
			return fmt.Errorf("could not create default configuration for provisioning: %s", err)
		}

		cfg.ConfigFile = cfile

	default:
		return fmt.Errorf("could not find configuration file %q and provisioning is not enabled", cfile)
	}

	cfg.MainCollective = "scout"
	cfg.Collectives = []string{"scout"}
	cfg.ApplyBuildSettings(bi)
	cfg.DisableSecurityProviderVerify = true
	cfg.InitiatedByServer = true

	if cfg.ConfigFile != "" {
		if cfg.Choria.ScoutTags == "" {
			cfg.Choria.ScoutTags = filepath.Join(filepath.Dir(cfg.ConfigFile), "tags.json")
		}
		if cfg.Choria.ScoutOverrides == "" {
			cfg.Choria.ScoutOverrides = filepath.Join(filepath.Dir(cfg.ConfigFile), "overrides.json")
		}
		if cfg.Choria.MachineSourceDir == "" {
			cfg.Choria.MachineSourceDir = filepath.Join(filepath.Dir(cfg.ConfigFile), "checks")
		}
	}

	err := os.MkdirAll(cfg.Choria.MachineSourceDir, 0755)
	if err != nil {
		log.Errorf("Could not create machine directory %s: %s", cfg.Choria.MachineSourceDir, err)
	}

	if debug {
		cfg.LogLevel = "debug"
	}

	fw, err = choria.NewWithConfig(cfg)
	if err != nil {
		return err
	}

	log = fw.Logger("scout")

	return nil
}
