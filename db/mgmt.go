package db

import (
	"errors"
	"fmt"
	"log"

	"github.com/couchbase/gocb/v2"
)

type ManagementConnection struct {
	cluster *gocb.Cluster
	s       *gocb.Scope
	bucket  *gocb.Bucket
}

func (m *ManagementConnection) Close() error {
	return m.cluster.Close(nil)
}

// Collections
const (
	cTeams               string = "teams"
	cCompletedChallenges string = "completedChallenges"
	cUsedHints           string = "usedHints"
)

var mgmtCollections = [...]string{
	cTeams,
	cCompletedChallenges,
	cUsedHints,
}

var mgmtIndexes = [...]string{
	fmt.Sprintf("CREATE PRIMARY INDEX ON %s", cTeams),
	fmt.Sprintf("CREATE INDEX idx_team_members ON `%s` (ALL members)", cTeams),
	fmt.Sprintf(`CREATE INDEX idx_completedChallenges ON %s (team_id, dataset_id, query_id)`, cCompletedChallenges),
}

func (m *ManagementConnection) init() error {
	log.Println("Initialising management collections")
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
	log.Println("Creating indexes")
	indexesNeedBuilding := 0
	for _, idx := range mgmtIndexes {
		_, err := m.s.Query(idx+" WITH {\"defer_build\": true}", nil)
		log.Printf("%s %v", idx, err)
		if errors.Is(err, gocb.ErrIndexExists) {
			continue
		}
		// ^ doesn't catch for primary indexes
		var qe *gocb.QueryError
		if errors.As(err, &qe) && len(qe.Errors) > 0 && qe.Errors[0].Code == 4300 {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
		indexesNeedBuilding++
	}
	if indexesNeedBuilding > 0 {
		log.Printf("Building %d indexes", indexesNeedBuilding)
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
