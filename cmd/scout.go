package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"github.com/choria-io/go-choria/build"
	"github.com/choria-io/go-choria/choria"
	"github.com/choria-io/go-choria/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cfile  string
	debug  bool
	bi     *build.Info
	cfg    *config.Config
	fw     *choria.Framework
	err    error
	log    *logrus.Entry
	wg     sync.WaitGroup
	ctx    context.Context
	cancel func()
	mu     = sync.Mutex{}
)

func Run() error {
	bi = &build.Info{}
	ctx, cancel = context.WithCancel(context.Background())
	log = logrus.NewEntry(logrus.New())

	scli := kingpin.New("scout", "Choria Scout")
	scli.Author("R.I.Pienaar <rip@devco.net>")
	scli.Version(bi.Version())
	scli.HelpFlag.Short('h')

	scli.Flag("config", "Path to the configuration file").StringVar(&cfile)
	scli.Flag("debug", "Enables debug level logging").BoolVar(&debug)

	configureRunCommand(scli)

	go interruptWatcher()

	wg.Add(1)
	kingpin.MustParse(scli.Parse(os.Args[1:]))

	wg.Wait()

	return nil
}

func forcequit() {
	grace := 2 * time.Second

	if cfg != nil {
		if cfg.SoftShutdownTimeout > 0 {
			grace = time.Duration(cfg.SoftShutdownTimeout) * time.Second
		}
	}

	<-time.NewTimer(grace).C

	dumpGoRoutines()

	log.Errorf("Forced shutdown triggered after %v", grace)

	os.Exit(1)
}

func interruptWatcher() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case sig := <-sigs:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				go forcequit()

				log.Infof("Shutting down on %s", sig)
				cancel()

			case syscall.SIGQUIT:
				dumpGoRoutines()
			}
		case <-ctx.Done():
			return
		}
	}
}

func dumpGoRoutines() {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().UnixNano()
	pid := os.Getpid()

	tdoutname := filepath.Join(os.TempDir(), fmt.Sprintf("scout-threaddump-%d-%d.txt", pid, now))
	memoutname := filepath.Join(os.TempDir(), fmt.Sprintf("scout-memoryprofile-%d-%d.mprof", pid, now))

	buf := make([]byte, 1<<20)
	stacklen := runtime.Stack(buf, true)

	err := ioutil.WriteFile(tdoutname, buf[:stacklen], 0644)
	if err != nil {
		log.Errorf("Could not produce thread dump: %s", err)
		return
	}

	log.Warnf("Produced thread dump to %s", tdoutname)

	mf, err := os.Create(memoutname)
	if err != nil {
		log.Errorf("Could not produce memory profile: %s", err)
		return
	}
	defer mf.Close()

	err = pprof.WriteHeapProfile(mf)
	if err != nil {
		log.Errorf("Could not produce memory profile: %s", err)
		return
	}

	log.Warnf("Produced memory profile to %s", memoutname)
}
