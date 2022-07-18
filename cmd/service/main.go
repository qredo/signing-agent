package main

import (
	"fmt"
	"os"
	"os/signal"

	"gitlab.qredo.com/qredo-server/core-client/rest"

	"github.com/jessevdk/go-flags"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"go.uber.org/zap"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "not-set"
)

func startText() {
	fmt.Printf("Core Client service %v (%v) build date: %v\n\n", version, commit, buildDate)
}

type versionCmd struct{}

func (c *versionCmd) Execute([]string) error {
	return nil
}

type startCmd struct {
	ConfigFile string `short:"c" long:"config" description:"path to configuration file" default:"cc.yaml"`
}

func (c *startCmd) Execute([]string) error {
	var cfg config.Config
	cfg.Default()

	err := cfg.Load(c.ConfigFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	setCtrlC()
	log := logger(&cfg.Logging)

	router, err := rest.NewQRouter(log, &cfg)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if err = router.Start(); err != nil {
		log.Error("HTTP Listener error", "err", err)
	}
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	return nil
}

type initCmd struct {
	FileName string `short:"f" long:"file-name" description:"output file name" default:"cc.yaml"`
}

func (c *initCmd) Execute([]string) error {
	var cfg config.Config
	cfg.Default()
	if err := cfg.Save(c.FileName); err != nil {
		return err
	}

	fmt.Printf("written file %s\n\n", c.FileName)
	return nil
}

func main() {
	startText()

	var parser = flags.NewParser(nil, flags.Default)

	_, _ = parser.AddCommand("init", "init config", "write default config", &initCmd{})
	_, _ = parser.AddCommand("start", "start service", "", &startCmd{})
	_, _ = parser.AddCommand("version", "print version", "print service version and quit", &versionCmd{})

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}

func logger(cfg *config.Logging) *zap.SugaredLogger {
	logConfig := zap.NewProductionConfig()

	if cfg.Format == "text" {
		logConfig = zap.NewDevelopmentConfig()
	}

	logConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	if cfg.Level == "debug" {
		logConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logConfig.DisableStacktrace = true
	l, _ := logConfig.Build()

	return l.Sugar()
}

func setCtrlC() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for range sigChan {
			os.Exit(0)
		}
	}()
}
