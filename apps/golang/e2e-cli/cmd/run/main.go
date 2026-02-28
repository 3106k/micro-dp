package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"

	"github.com/user/micro-dp/e2e-cli/internal/config"
	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
	"github.com/user/micro-dp/e2e-cli/internal/runner"
	"github.com/user/micro-dp/e2e-cli/internal/runner/reporter"
	"github.com/user/micro-dp/e2e-cli/internal/suite"
)

func main() {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Fatalf("config error: %v", err)
	}

	scenarios, err := suite.Build(cfg)
	if err != nil {
		log.Fatalf("suite error: %v", err)
	}

	client := httpclient.New(cfg.BaseURL, cfg.Token, cfg.TenantID)
	run := runner.New(client)
	result := run.Run(context.Background(), scenarios)

	reporter.PrintConsole(os.Stdout, result)
	if err := reporter.WriteJSON(cfg.JSONOut, result); err != nil {
		log.Fatalf("failed to write json report: %v", err)
	}

	if result.Failed > 0 {
		os.Exit(1)
	}
}
