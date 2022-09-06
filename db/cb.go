package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"

	"query-adventure/cfg"
)

type QueryConnection struct {
	cluster      *gocb.Cluster
	queryTimeout time.Duration
}

type ManagementConnection struct {
	cluster *gocb.Cluster
	s       *gocb.Scope
	bucket  *gocb.Bucket
}

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

func (m *ManagementConnection) Close() error {
	return m.cluster.Close(nil)
}

var mgmtCollections = [...]string{
	cTeams,
	cCompletedChallenges,
}

var mgmtIndexes = [...]string{
	fmt.Sprintf("CREATE INDEX idx_team_members ON `%s` (ALL members)", cTeams),
}

func (m *ManagementConnection) init() error {
	for _, coll := range mgmtCollections {
		err := m.bucket.Collections().CreateCollection(gocb.CollectionSpec{
			Name:      coll,
			ScopeName: m.s.Name(),
		}, nil)
		if errors.Is(err, gocb.ErrCollectionExists) {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to create collection %q: %w", coll, err)
		}
	}
	indexesNeedBuilding := 0
	for _, idx := range mgmtIndexes {
		_, err := m.s.Query(idx+" WITH {\"defer_build\": true}", nil)
		if errors.Is(err, gocb.ErrIndexExists) {
			continue
		}
		// ^ doesn't catch for primary indexes
		var qe *gocb.QueryError
		if errors.As(err, &qe) && qe.Errors[0].Code == 4300 {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
		indexesNeedBuilding++
	}
	if indexesNeedBuilding > 0 {
		for _, coll := range mgmtCollections {
			_, err := m.cluster.QueryIndexes().BuildDeferredIndexes(m.bucket.Name(), &gocb.BuildDeferredQueryIndexOptions{
				ScopeName:      m.s.Name(),
				CollectionName: coll,
			})
			if err != nil {
				return fmt.Errorf("failed to build deferred indexes on %s: %w", coll, err)
			}
		}
	}
	return nil
}
