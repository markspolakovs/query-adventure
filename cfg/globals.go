package cfg

import (
	"time"

	"github.com/alecthomas/kong"
)

type DBCfg struct {
	ConnectionString   string `default:"couchbase://localhost"`
	QueryUsername      string `default:"Administrator"`
	QueryPassword      string `default:"password"`
	ManagementUsername string `default:"Administrator"`
	ManagementPassword string `default:"password"`
	ManagementBucket   string `default:"mgmt"`
	ManagementScope    string `default:"_default"`
	TxnsNoDurable      bool   `default:"false"`
}

type Globals struct {
	ConfigFile   kong.ConfigFlag
	QueryTimeout time.Duration            `default:"15s"`
	DatasetsPath string                   `default:"datasets.yml"`
	RateLimits   map[string]time.Duration `default:"query=5s;check=30s"`
	SessionKey   string                   `default:"CHANGEME"`
	DB           DBCfg                    `embed:"" prefix:"db."`
	HTTPPort     int                      `default:"7091"`
}
