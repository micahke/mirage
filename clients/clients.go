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
	Redis          RedisClient
	S3             S3Client
	S3Presign      PresignClient
}
