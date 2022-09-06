package db

import (
	"context"
	"fmt"

	"github.com/couchbase/gocb/v2"
)

// Collections
const (
	cTeams              string = "teams"
	cCompleteChallenges string = "completeChallenges"
)

type Team struct {
	Name    string   `json:"name"`
	Color   string   `json:"color"`
	Members []string `json:"members"`
}

func (m *ManagementConnection) GetTeamForUser(ctx context.Context, email string) (Team, error) {
	qr, err := m.s.Query(fmt.Sprintf("SELECT RAW t FROM %s t WHERE $1 IN t.members LIMIT 1", cTeams), &gocb.QueryOptions{
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
