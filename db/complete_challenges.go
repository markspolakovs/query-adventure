package db

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"query-adventure/data"

	"github.com/couchbase/gocb/v2"
	"github.com/labstack/echo/v4"
)

type CompleteChallenge struct {
	DatasetID   string    `json:"dataset_id"`
	QueryID     string    `json:"query_id"`
	TeamID      string    `json:"team_id"`
	User        string    `json:"user"`
	CompletedAt time.Time `json:"completed_at"`
	Points      uint      `json:"points"`
}

func completeChallengeDocKey(teamID, datasetID, queryID string) string {
	return fmt.Sprintf("%s::%s::%s", teamID, datasetID, queryID)
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
	id := completeChallengeDocKey(team.ID, dataset.ID, query.ID)
	_, err := m.s.Collection(cCompletedChallenges).Insert(id, cc, &gocb.InsertOptions{
		Context: ctx,
	})
	if errors.Is(err, gocb.ErrDocumentExists) {
		return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("team %q has already completed challenge %s.%s", team.Name, dataset.ID, query.ID))
	}
	if err != nil {
		return fmt.Errorf("failed to insert cc %q: %w", id, err)
	}
	return nil
}
