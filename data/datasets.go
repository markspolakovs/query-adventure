package data

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"query-adventure/cfg"
)

type Query struct {
	Name      string   `yaml:"name" json:"name"`
	Challenge string   `yaml:"challenge" json:"challenge"`
	Points    uint     `yaml:"points" json:"points"`
	Query     string   `yaml:"query" json:"query,omitempty"`
	Hints     []string `yaml:"hints" json:"hints"`
}

type Dataset struct {
	Name        string  `yaml:"name" json:"name"`
	Description string  `yaml:"description" json:"description"`
	Keyspace    string  `yaml:"keyspace" json:"keyspace"`
	Queries     []Query `yaml:"queries" json:"queries"`
}

type Datasets []Dataset

func (d Datasets) FilterQueries() Datasets {
	res := make(Datasets, len(d))
	for i, ds := range d {
		res[i] = Dataset{
			Name:        ds.Name,
			Description: ds.Description,
			Keyspace:    ds.Keyspace,
			Queries:     make([]Query, len(ds.Queries)),
		}
		for j, q := range ds.Queries {
			res[i].Queries[j] = Query{
				Name:      q.Name,
				Challenge: q.Challenge,
				Points:    q.Points,
				Hints:     q.Hints, // FIXME
			}
		}
	}
	return res
}

func LoadDatasets(g *cfg.Globals) (Datasets, error) {
	fd, err := os.Open(g.DatasetsPath)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", g.DatasetsPath, err)
	}
	defer func(fd *os.File) {
		_ = fd.Close()
	}(fd)
	var ds []Dataset
	err = yaml.NewDecoder(fd).Decode(&ds)
	if err != nil {
		return nil, fmt.Errorf("decode %q: %w", g.DatasetsPath, err)
	}
	return ds, nil
}
