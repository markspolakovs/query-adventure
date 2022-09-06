package db

import (
	"fmt"

	"github.com/couchbase/gocb/v2"

	"query-adventure/cfg"
)

func Connect(g *cfg.Globals) (*QueryConnection, *ManagementConnection, error) {
	qCluster, err := gocb.Connect(g.DB.ConnectionString, gocb.ClusterOptions{
		Username: g.DB.QueryUsername,
		Password: g.DB.QueryPassword,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect using query creds: %w", err)
	}
	mCluster, err := gocb.Connect(g.DB.ConnectionString, gocb.ClusterOptions{
		Username: g.DB.ManagementUsername,
		Password: g.DB.ManagementPassword,
	})
	if err != nil {
		_ = qCluster.Close(nil)
		return nil, nil, fmt.Errorf("failed to connect using management creds: %w", err)
	}
	q := &QueryConnection{
		cluster:      qCluster,
		queryTimeout: g.QueryTimeout,
	}
	mgmt := &ManagementConnection{
		cluster: mCluster,
		bucket:  mCluster.Bucket(g.DB.ManagementBucket),
		s:       mCluster.Bucket(g.DB.ManagementBucket).Scope(g.DB.ManagementScope),
	}
	err = mgmt.init()
	if err != nil {
		_ = qCluster.Close(nil)
		_ = mCluster.Close(nil)
		return nil, nil, fmt.Errorf("failed to initialize mgmt: %w", err)
	}
	return q, mgmt, nil
}

func (c *QueryConnection) Close() error {
	return c.cluster.Close(nil)
}
