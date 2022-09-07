package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/alecthomas/kong"

	"query-adventure/auth"
	"query-adventure/cfg"
	"query-adventure/data"
	"query-adventure/db"
	"query-adventure/rest"
)

type RunCmd struct {
	GoogleCfg cfg.GoogleCfg `embed:"" prefix:"google."`
}

func (r *RunCmd) Run(g *cfg.Globals) error {
	log.Println("Connecting to CB...")
	qCB, mCB, err := db.Connect(g)
	if err != nil {
		return err
	}
	defer qCB.Close()
	defer mCB.Close()

	log.Println("Loading datasets...")
	datasets, err := data.LoadDatasets(g)
	if err != nil {
		return err
	}

	log.Println("Constructing authenticator...")
	authn, err := auth.NewGoogleAuthenticator(r.GoogleCfg)
	if err != nil {
		return err
	}

	log.Println("Building API...")
	api := rest.NewAPI(g, qCB, mCB, datasets, authn)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	return api.Start(ctx)
}

func main() {
	var CLI struct {
		cfg.Globals
		Run RunCmd `cmd:""`
	}
	ctx := kong.Parse(&CLI, kong.DefaultEnvars("Q"), kong.Configuration(kong.JSON))
	err := ctx.Run(&CLI.Globals)
	ctx.FatalIfErrorf(err)
}
