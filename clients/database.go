package clients

import "context"

type InsertOneRequest struct {
	Database   string
	Collection string
	Document   interface{}
}

type InsertManyRequest struct {
	Database   string
	Collection string
	Documents  []interface{}
}

type FindOneRequest struct {
	Database   string
	Collection string
	Filter     interface{}
}

type FindRequest struct {
	Database   string
	Collection string
	Filter     interface{}
	Limit      int64
	Skip       int64
	Sort       interface{}
}

type ExistsRequest struct {
	Database   string
	Collection string
	Filter     interface{}
}

type AggregateRequest struct {
	Database   string
	Collection string
	Pipeline   interface{}
}

type UpdateOneRequest struct {
	Database   string
	Collection string
	Filter     interface{}
	Update     interface{}
}

type ReplaceOneRequest struct {
	Database    string
	Collection  string
	Filter      interface{}
	Replacement interface{}
}

type DatabaseClient interface {
	InsertOne(context.Context, *InsertOneRequest) error
	InsertMany(context.Context, *InsertManyRequest) error
	FindOne(context.Context, *FindOneRequest, interface{}) error
	Find(context.Context, *FindRequest, interface{}) error
}
