package service

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"gitlab.qredo.com/qredo-server/core-client/qhttp"

	"gitlab.qredo.com/qredo-server/qredo-core/qlog"

	"go.uber.org/zap"

	"github.com/urfave/cli/v2"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/qredo-core/qcmd"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "not-set"
)

type CoreClientService interface {
	Start() error
}

type coreClientService struct {
	cli    *qcmd.QCli
	config config.Config
	log    *zap.SugaredLogger
}

func (s *coreClientService) Start() error {

	// run the cli logic
	if err := s.cli.Run(nil); err != nil {
		return err
	}

	return nil
}

func New() CoreClientService {

	s := &coreClientService{
		cli: qcmd.NewQCli("core-client"),
	}
	s.setCtrlC()
	s.commandsRegister()

	return s
}

func (s *coreClientService) clean() error {
	return nil
}

// initialize child service after config has been loaded
func (s *coreClientService) run() error {

	router, err := qhttp.NewQRouter(s.log, &s.config)
	if err != nil {
		return err
	}

	if err := router.Start(); err != nil {
		s.log.Error("HTTP Listener error", "err", err)
	}

	return nil
}

func (s *coreClientService) commandsRegister() {

	cm := []*cli.Command{
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "Start Core Client Service.",
			Action: func(c *cli.Context) error {
				s.startText()
				configFile := c.String("config")
				err := s.config.Load(configFile)
				if err != nil {
					return err
				}
				s.log = qlog.GetLogger(&s.config.Logging)
				return s.run()
			},
		},
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Init the config file with default values.",
			Action: func(c *cli.Context) error {
				s.startText()
				cfgFilePath, err := filepath.Abs(c.String("config"))
				if err != nil {
					return err
				}

				s.config.Default()
				if err := s.config.Save(cfgFilePath); err != nil {
					fmt.Println(err.Error())
					return err
				}

				fmt.Printf("Config file created: %v\n", cfgFilePath)
				return nil
			},
		},
	}

	s.cli.SetCommands(cm)
}

func (s *coreClientService) startText() {
	fmt.Printf("Core Client service v%v (%v) build date: %v\n\n", version, commit, buildDate)
}

func (s *coreClientService) cleanup() error {
	return nil
}

func (s *coreClientService) setCtrlC() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for range sigChan {
			err := s.cleanup()
			if err != nil {
				fmt.Printf("\n%s\n", err.Error())
				os.Exit(1)
			}
			os.Exit(0)
		}
	}()
}
