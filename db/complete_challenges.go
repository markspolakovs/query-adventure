package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"query-adventure/data"

	"github.com/couchbase/gocb/v2"
)

type CompleteChallenge struct {
	DatasetID   string    `json:"dataset_id"`
	QueryID     string    `json:"query_id"`
	TeamID      string    `json:"team_id"`
	User        string    `json:"user"`
	CompletedAt time.Time `json:"completed_at"`
	Points      uint      `json:"points"`
}

func (m *ManagementConnection) CompleteChallenge(ctx context.Context, dataset data.Dataset, query data.Query, team Team, email string) error {
	now := time.Now()
	cc := CompleteChallenge{
		DatasetID:   dataset.ID,
		QueryID:     query.ID,
		TeamID:      team.ID,
		User:        email,
		CompletedAt: now,
		Points:      query.Points,
	}
	id := strconv.FormatInt(now.Unix(), 10)
	// FIXME: what if they've already completed it
	_, err := m.s.Collection(cCompletedChallenges).Insert(id, cc, &gocb.InsertOptions{
		Context: ctx,
	})
	if err != nil {
		return fmt.Errorf("failed to insert cc %q: %w", id, err)
	}
	return nil
}
