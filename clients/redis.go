package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type RedisClient interface {
	Get(context context.Context, key string) *redis.StringCmd
	MGet(context context.Context, keys ...string) *redis.SliceCmd
	Set(context context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	LPush(context context.Context, key string, values ...interface{}) *redis.IntCmd
	BLPop(context context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd
	Del(context context.Context, keys ...string) *redis.IntCmd
}

func RedisID(prefix string, id string) string {
	return fmt.Sprintf("%s:%s", prefix, id)
}

var _ RedisClient = (*redis.Client)(nil)

type redisClient struct {
	client *redis.Client
}

func NewRedisCacheClient(client *redis.Client) *redisClient {
	return &redisClient{client: client}
}

func (rc *redisClient) Get(ctx context.Context, key string, value interface{}) error {
	result := rc.client.Get(ctx, key)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key %s not found", key)
		}
		return fmt.Errorf("redis get error: %w", err)
	}

	jsonString, err := result.Result()
	if err != nil {
		return fmt.Errorf("failed to get result: %w", err)
	}

	if err := json.Unmarshal([]byte(jsonString), value); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

func (rc *redisClient) GetMany(ctx context.Context, keys []string, values interface{}) error {
  result := rc.client.MGet(ctx, keys...)
  if err := result.Err(); err != nil {
    return fmt.Errorf("redis mget error: %w", err)
  }

  jsonStrings, err := result.Result()
  if err != nil {
    return fmt.Errorf("failed to get result: %w", err)
  }

  data, err := json.Marshal(jsonStrings)
  if err != nil {
    return fmt.Errorf("failed to marshal json: %w", err)
  }

  if err := json.Unmarshal(data, values); err != nil {
    return fmt.Errorf("failed to unmarshal values: %w", err)
  }

  return nil
}

func (rc *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	jsonString := string(data)
	if err := rc.client.Set(ctx, key, jsonString, expiration).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (rc *redisClient) Delete(ctx context.Context, key string) error {
  if err := rc.client.Del(ctx, key).Err(); err != nil {
    return fmt.Errorf("redis del error: %w", err)
  }
  return nil
}

func (rc *redisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
  return rc.client.Del(ctx, keys...)
}

func (rc *redisClient) BLPop(context context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd {
  return rc.client.BLPop(context, timeout, keys...)
}

// ProtoClient wraps RedisClient to handle protobuf operations
type ProtoClient struct {
	client RedisClient
}

// NewProtoClient creates a new ProtoClient instance
func NewProtoClient(client RedisClient) *ProtoClient {
	return &ProtoClient{client: client}
}

// GetProto retrieves and unmarshals a protobuf message
func (pc *ProtoClient) GetProto(ctx context.Context, key string, msg proto.Message) error {
	result := pc.client.Get(ctx, key)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key %s not found", key)
		}
		return fmt.Errorf("redis get error: %w", err)
	}

	data, err := result.Bytes()
	if err != nil {
		return fmt.Errorf("failed to get bytes: %w", err)
	}

	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("failed to unmarshal proto: %w", err)
	}

	return nil
}

// SetProto marshals and stores a protobuf message
func (pc *ProtoClient) SetProto(ctx context.Context, key string, msg proto.Message, expiration time.Duration) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal proto: %w", err)
	}

	if err := pc.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

// LPushProto marshals and pushes a protobuf message to the head of a list
func (pc *ProtoClient) LPushProto(ctx context.Context, key string, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal proto: %w", err)
	}

	if err := pc.client.LPush(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("redis lpush error: %w", err)
	}

	return nil
}

// BLPopProto blocks and waits to pop and unmarshal a protobuf message
func (pc *ProtoClient) BLPopProto(ctx context.Context, timeout time.Duration, msg proto.Message, keys ...string) (string, error) {
	result := pc.client.BLPop(ctx, timeout, keys...)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("no data available within timeout")
		}
		return "", fmt.Errorf("redis blpop error: %w", err)
	}

	// BLPop returns [key, value]
	values, err := result.Result()
	if err != nil {
		return "", fmt.Errorf("failed to get result: %w", err)
	}

	if len(values) != 2 {
		return "", fmt.Errorf("unexpected result length: got %d, want 2", len(values))
	}

	if err := proto.Unmarshal([]byte(values[1]), msg); err != nil {
		return "", fmt.Errorf("failed to unmarshal proto: %w", err)
	}

	return values[0], nil
}

// DeleteKeys deletes one or more keys
func (pc *ProtoClient) DeleteKeys(ctx context.Context, keys ...string) error {
	if err := pc.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis del error: %w", err)
	}
	return nil
}

// ProtoKey generates a typed key for proto messages
func ProtoKey(prefix string, messageType string, id string) string {
	return fmt.Sprintf("%s:%s:%s", prefix, messageType, id)
}
