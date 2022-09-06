package db

import (
	"context"
	"fmt"

	"github.com/couchbase/gocb/v2"
)

type Team struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Color   string   `json:"color"` // TODO unused
	Members []string `json:"members"`
}

func (m *ManagementConnection) GetTeamForUser(ctx context.Context, email string) (Team, error) {
	// TODO consider a cache?
	qr, err := m.s.Query(fmt.Sprintf("SELECT RAW t FROM %s t WHERE ANY m IN t.members SATISFIES m = $1 END LIMIT 1", cTeams), &gocb.QueryOptions{
		Context:              ctx,
		PositionalParameters: []any{email},
	})
	if err != nil {
		return Team{}, fmt.Errorf("failed to perform teams query: %w", err)
	}
	var team Team
	err = qr.One(&team)
	if err != nil {
		return team, fmt.Errorf("failed to parse team info: %w", err)
	}
	return team, nil
}

func (m *ManagementConnection) GetTeamCompleteChallenges(ctx context.Context, team Team) (map[string][]string, error) {
	qr, err := m.s.Query(fmt.Sprintf(`SELECT dataset_id, query_id FROM %s WHERE team_id = $1`, cCompletedChallenges), &gocb.QueryOptions{
		Context:              ctx,
		PositionalParameters: []any{team.ID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute complete challenges query: %w", err)
	}
	result := make(map[string][]string)
	for qr.Next() {
		var row struct {
			DatasetID string `json:"dataset_id"`
			QueryID   string `json:"query_id"`
		}
		err = qr.Row(&row)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CC query row: %w", err)
		}
		result[row.DatasetID] = append(result[row.DatasetID], row.QueryID)
	}
	err = qr.Close()
	if err != nil {
		return nil, fmt.Errorf("CC query close: %w", err)
	}
	return result, nil
}
