package cfg

import (
	"time"

	"github.com/alecthomas/kong"
)

type DBCfg struct {
	ConnectionString string `default:"couchbase://localhost"`
	Username         string `default:"Administrator"`
	Password         string `default:"password"`
}

type Globals struct {
	ConfigFile   kong.ConfigFlag
	QueryTimeout time.Duration            `default:"5s"`
	DatasetsPath string                   `default:"datasets.yml"`
	RateLimits   map[string]time.Duration `default:"query=10s;check=60s"`
	SessionKey   string                   `default:"CHANGEME"`
	DB           DBCfg                    `embed:"" prefix:"db."`
	HTTPPort     int                      `default:"7091"`
}
