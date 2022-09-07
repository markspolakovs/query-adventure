package db

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"query-adventure/cfg"
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
	RawQuery    string    `json:"raw_query"`
	RawPoints   uint      `json:"raw_points"`
	HintsUsed   uint      `json:"hints_used"`
	First       bool      `json:"first"`
	FinalPoints float64   `json:"points"`
}

func (cc *CompleteChallenge) calculateFinalPoints(g *cfg.Globals) {
	base := float64(cc.RawPoints)
	base *= math.Pow(g.ScoreHintMultiplier, float64(cc.HintsUsed))
	if cc.First {
		base *= g.ScoreFirstTeamMultiplier
	}
	cc.FinalPoints = math.Round(base*10) / 10
}

func completeChallengeDocKey(teamID, datasetID, queryID string) string {
	return fmt.Sprintf("%s::%s::%s", teamID, datasetID, queryID)
}

func (m *ManagementConnection) CompleteChallenge(ctx context.Context, g *cfg.Globals, dataset data.Dataset, query data.Query, team Team, email string, rawQuery string, hintsUsed uint) (CompleteChallenge, error) {
	now := time.Now()
	var cc CompleteChallenge
	_, err := m.cluster.Transactions().Run(func(tx *gocb.TransactionAttemptContext) error {
		cc = CompleteChallenge{
			DatasetID:   dataset.ID,
			QueryID:     query.ID,
			TeamID:      team.ID,
			User:        email,
			CompletedAt: now,
			RawQuery:    rawQuery,
			RawPoints:   query.Points,
			HintsUsed:   hintsUsed,
			First:       false,
		}
		// Check if other teams have completed it
		qr, err := tx.Query(fmt.Sprintf("SELECT COUNT(*) AS count FROM `%s`.`%s`.`%s` WHERE dataset_id = $1 AND query_id = $2 AND team_id = $3", m.bucket.Name(), m.s.Name(), cCompletedChallenges), &gocb.TransactionQueryOptions{
			PositionalParameters: []any{dataset.ID, query.ID, team.ID},
		})
		if err != nil {
			return fmt.Errorf("cc other teams query failed: %w", err)
		}
		var result struct {
			Count uint `json:"count"`
		}
		err = qr.One(&result)
		if err != nil {
			return fmt.Errorf("failed to parse other team query result: %w", err)
		}
		cc.First = result.Count == 0
		cc.calculateFinalPoints(g)
		id := completeChallengeDocKey(team.ID, dataset.ID, query.ID)
		_, err = tx.Insert(m.s.Collection(cCompletedChallenges), id, cc)
		if errors.Is(err, gocb.ErrDocumentExists) {
			return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("team %q has already completed challenge %s.%s", team.Name, dataset.ID, query.ID))
		}
		if err != nil {
			return fmt.Errorf("failed to insert cc %q: %w", id, err)
		}
		return nil
	}, &gocb.TransactionOptions{})
	if err != nil {
		return CompleteChallenge{}, fmt.Errorf("failed to execute cc txn: %w", err)
	}
	return cc, nil
}
