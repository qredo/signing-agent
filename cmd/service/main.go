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
	fmt.Printf("Core Client service v%v (%v) build date: %v\n\n", version, commit, buildDate)
}

var opts struct {
	ConfigFile string `short:"f" long:"config-file" description:"location of configuration file" default:"lib-client.yaml"`
	Version    func() `short:"v" long:"version" description:"print service version"`
}

type InitConfig struct {
	FileName string `short:"f" long:"file-name" description:"output file name" default:"cc.yaml"`
}

func (x *InitConfig) Execute([]string) error {
	var cfg config.Config
	cfg.Default()
	if err := cfg.Save(x.FileName); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func main() {
	startText()
	var cfg config.Config
	cfg.Default()

	// version already printed, so just exit silently. more to add if necessary
	opts.Version = func() {
		os.Exit(0)
	}

	parser := flags.NewParser(&opts, flags.None)
	_, _ = parser.AddCommand("init", "init config", "write default config", &InitConfig{})

	_, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = cfg.Load(opts.ConfigFile)
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
