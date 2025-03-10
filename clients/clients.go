package clients

import (
	"github.com/micahke/mirage/clients/cache"
)

type Clients struct {
	Logger         Logger
	Stats          StatsClient
	MongoClient    MongoClient
	DatabaseClient DatabaseClient
	Cache          cache.Cache
}
