package db

import (
	"time"

	"github.com/couchbase/gocb/v2"

	"query-adventure/cfg"
)

type CBDatabase struct {
	cluster      *gocb.Cluster
	queryTimeout time.Duration
}

func Connect(g *cfg.Globals) (*CBDatabase, error) {
	cluster, err := gocb.Connect(g.DB.ConnectionString, gocb.ClusterOptions{
		Username: g.DB.Username,
		Password: g.DB.Password,
	})
	if err != nil {
		return nil, err
	}
	return &CBDatabase{
		cluster:      cluster,
		queryTimeout: g.QueryTimeout,
	}, nil
}

func (c *CBDatabase) Close() error {
	return c.cluster.Close(nil)
}
