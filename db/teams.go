package db

import (
	"context"
	"errors"
	"fmt"

	"query-adventure/data"

	"github.com/couchbase/gocb/v2"
)

type Team struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Color   string   `json:"color"` // TODO unused
	Members []string `json:"members"`
}

func (m *ManagementConnection) GetAllTeams(ctx context.Context) ([]Team, error) {
	qr, err := m.s.Query(fmt.Sprintf(`SELECT RAW t FROM %s t`, cTeams), &gocb.QueryOptions{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute all teams query: %w", err)
	}
	result := make([]Team, 0)
	for qr.Next() {
		var row Team
		err = qr.Row(&row)
		if err != nil {
			return nil, fmt.Errorf("failed to parse all teams query result: %w", err)
		}
		result = append(result, row)
	}
	err = qr.Close()
	if err != nil {
		return nil, fmt.Errorf("all teams query close failure: %w", err)
	}
	return result, nil
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

// GetAllTeamCompleteChallenges returns all the challenges, along with whether teams have completed them. The result is
// keyed by dataset ID -> query ID -> team ID.
func (m *ManagementConnection) GetAllTeamCompleteChallenges(ctx context.Context, allDatasets data.Datasets) (map[string]map[string]map[string]bool, error) {
	allTeams, err := m.GetAllTeams(ctx)
	if err != nil {
		return nil, err
	}
	qr, err := m.s.Query(fmt.Sprintf(`SELECT team_id, dataset_id, query_id FROM %s`, cCompletedChallenges), &gocb.QueryOptions{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute all-team CC query: %w", err)
	}
	result := make(map[string]map[string]map[string]bool)
	for _, ds := range allDatasets {
		result[ds.ID] = make(map[string]map[string]bool)
		for _, q := range ds.Queries {
			result[ds.ID][q.ID] = make(map[string]bool)
			for _, team := range allTeams {
				result[ds.ID][q.ID][team.ID] = false
			}
		}
	}
	for qr.Next() {
		var row struct {
			TeamID    string `json:"team_id"`
			DatasetID string `json:"dataset_id"`
			QueryID   string `json:"query_id"`
		}
		err = qr.Row(&row)
		if err != nil {
			return nil, fmt.Errorf("failed to parse all-team CC result: %w", err)
		}
		result[row.DatasetID][row.QueryID][row.TeamID] = true
	}
	err = qr.Close()
	if err != nil {
		return nil, fmt.Errorf("all-teams CC close failure: %w", err)
	}
	return result, nil
}

func (m *ManagementConnection) GetTeamScores(ctx context.Context) (map[string]float64, error) {
	qr, err := m.s.Query(fmt.Sprintf(`SELECT team_id, SUM(points) AS points FROM %s GROUP BY team_id`, cCompletedChallenges), &gocb.QueryOptions{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute points query: %w", err)
	}
	result := make(map[string]float64)
	for qr.Next() {
		var row struct {
			TeamID string  `json:"team_id"`
			Points float64 `json:"points"`
		}
		err = qr.Row(&row)
		if err != nil {
			return nil, fmt.Errorf("failed to parse points query row: %w", err)
		}
		result[row.TeamID] = row.Points
	}
	err = qr.Close()
	if err != nil {
		return nil, fmt.Errorf("points query close: %w", err)
	}
	return result, nil
}

func (m *ManagementConnection) GetUsedHints(ctx context.Context, datasetID, queryID, teamID string) (uint, error) {
	result, _, err := m.getUsedHints(ctx, datasetID, queryID, teamID)
	if errors.Is(err, gocb.ErrDocumentNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return result, nil
}

// UseHint marks one hint as used. Returns the current number of hints, whether one was actually used, and the error.
// Will return (curr, false, nil) if using one more hint would take the team over the max.
func (m *ManagementConnection) UseHint(ctx context.Context, datasetID, queryID, teamID string, max int) (uint, bool, error) {
	curr, cas, err := m.getUsedHints(ctx, datasetID, queryID, teamID)
	if errors.Is(err, gocb.ErrDocumentNotFound) {
		if max == 0 {
			return 0, false, nil
		}
		_, err = m.s.Collection(cUsedHints).Insert(usedHintsKey(datasetID, queryID, teamID), 1, &gocb.InsertOptions{
			Context: ctx,
		})
	} else if err != nil {
		return 0, false, err
	} else {
		if curr+1 > uint(max) {
			return 0, false, nil
		}
		_, err = m.s.Collection(cUsedHints).Replace(usedHintsKey(datasetID, queryID, teamID), curr+1, &gocb.ReplaceOptions{
			Context: ctx,
			Cas:     cas,
		})
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to execute insert/replace: %w", err)
	}
	return curr + 1, true, nil
}

func (m *ManagementConnection) getUsedHints(ctx context.Context, datasetID, queryID, teamID string) (uint, gocb.Cas, error) {
	res, err := m.s.Collection(cUsedHints).Get(usedHintsKey(datasetID, queryID, teamID), &gocb.GetOptions{
		Context: ctx,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get used hints: %w", err)
	}
	var result uint
	err = res.Content(&result)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse used hints: %w", err)
	}
	return result, res.Cas(), nil
}

func usedHintsKey(datasetID, queryID, teamID string) string {
	return fmt.Sprintf("%s::%s::%s", datasetID, queryID, teamID)
}
