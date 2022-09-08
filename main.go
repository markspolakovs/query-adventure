package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
	"go.uber.org/multierr"

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

type TestCmd struct {
	Dataset string `help:"which dataset's queries to test - omit to run all"`
	Query   string `help:"which query in the dataset to test - omit to run alll"`
}

func (t *TestCmd) Run(g *cfg.Globals) error {
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

	if t.Dataset != "" {
		ds, ok := datasets.DatasetByID(t.Dataset)
		if !ok {
			return fmt.Errorf("dataset %q not found", t.Dataset)
		}
		datasets = data.Datasets{ds}
	}

	var errs error
	for _, ds := range datasets {
		var queries []data.Query
		if t.Query == "" {
			queries = ds.Queries
		} else {
			q, ok := ds.QueryByID(t.Query)
			if !ok {
				return fmt.Errorf("failed to find query %s.%s", ds.ID, t.Query)
			}
			queries = []data.Query{q}
		}
		for _, q := range queries {
			start := time.Now()
			err = qCB.ExecuteAndVerifyQuery(context.TODO(), ds.Keyspace, q.Query, q.Query)
			end := time.Now()
			if err != nil {
				log.Printf("FAIL %s.%s: %v", ds.ID, q.ID, err)
				multierr.AppendInto(&errs, err)
			} else {
				log.Printf("OK %s.%s took %v", ds.ID, q.ID, end.Sub(start))
			}
		}
	}
	return errs
}

func main() {
	var CLI struct {
		cfg.Globals
		Run  RunCmd  `cmd:""`
		Test TestCmd `cmd:""`
	}
	ctx := kong.Parse(&CLI, kong.DefaultEnvars("Q"), kong.Configuration(kong.JSON))
	err := ctx.Run(&CLI.Globals)
	ctx.FatalIfErrorf(err)
}
