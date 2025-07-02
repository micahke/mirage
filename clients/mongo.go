package clients

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoIndexView interface {
	CreateOne(ctx context.Context, model mongo.IndexModel) (string, error)
}

type MongoCollection interface {
	InsertOne(ctx context.Context, document interface{}) error
	InsertMany(ctx context.Context, documents []interface{}) error
	FindOne(ctx context.Context, filter interface{}, result interface{}) error
	Find(ctx context.Context, filter interface{}, results interface{}, options ...*options.FindOptions) error
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error)

	Indexes() MongoIndexView
	Exists(ctx context.Context, filter interface{}) (bool, error)
	Aggregate(ctx context.Context, pipeline interface{}, results interface{}) error
}

type MongoClient interface {
	Collection(database, collection string) MongoCollection
	InsertOne(ctx context.Context, req *InsertOneRequest) error
	InsertMany(ctx context.Context, req *InsertManyRequest) error
	FindOne(ctx context.Context, req *FindOneRequest, result interface{}) error
	Find(ctx context.Context, req *FindRequest, results interface{}, options ...*options.FindOptions) error
	Exists(ctx context.Context, req *ExistsRequest) (bool, error)
	Aggregate(ctx context.Context, req *AggregateRequest, results interface{}) error
	Disconnect(ctx context.Context) error
}

// Concrete implementation
type mongoCollection struct {
	coll *mongo.Collection
}

func (c *mongoCollection) Indexes() MongoIndexView {
	return &mongoIndexView{
		indexes: c.coll.Indexes(),
	}
}

func IsNoDocumentsFound(err error) bool {
	return err == mongo.ErrNoDocuments
}

// Implementation for indexes
type mongoIndexView struct {
	indexes mongo.IndexView
}

func (iv *mongoIndexView) CreateOne(ctx context.Context, model mongo.IndexModel) (string, error) {
	return iv.indexes.CreateOne(ctx, model)
}

func (c *mongoCollection) InsertOne(ctx context.Context, document interface{}) error {
	_, err := c.coll.InsertOne(ctx, document)
	return err
}

func (c *mongoCollection) InsertMany(ctx context.Context, documents []interface{}) error {
	_, err := c.coll.InsertMany(ctx, documents)
	return err
}

func (c *mongoCollection) FindOne(ctx context.Context, filter interface{}, result interface{}) error {
	return c.coll.FindOne(ctx, filter).Decode(result)
}

func (c *mongoCollection) Find(ctx context.Context, filter interface{}, results interface{}, opts ...*options.FindOptions) error {
	cursor, err := c.coll.Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, results)
}

func (c *mongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.coll.UpdateOne(ctx, filter, update, opts...)
}

func (c *mongoCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return c.coll.UpdateMany(ctx, filter, update)
}

func (c *mongoCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return c.coll.DeleteOne(ctx, filter)
}

func (c *mongoCollection) DeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return c.coll.DeleteMany(ctx, filter)
}

func (c *mongoCollection) Exists(ctx context.Context, filter interface{}) (bool, error) {
	count, err := c.coll.CountDocuments(ctx, filter)
	return count > 0, err
}

func (c *mongoCollection) Aggregate(ctx context.Context, pipeline interface{}, results interface{}) error {
	cursor, err := c.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, results)
}

type mongoClient struct {
	client *mongo.Client
}

func (c *mongoClient) Collection(database, collection string) MongoCollection {
	return &mongoCollection{
		coll: c.client.Database(database).Collection(collection),
	}
}

func (c *mongoClient) InsertOne(ctx context.Context, req *InsertOneRequest) error {
	return c.Collection(req.Database, req.Collection).InsertOne(ctx, req.Document)
}

func (c *mongoClient) InsertMany(ctx context.Context, req *InsertManyRequest) error {
	return c.Collection(req.Database, req.Collection).InsertMany(ctx, req.Documents)
}

func (c *mongoClient) FindOne(ctx context.Context, req *FindOneRequest, result interface{}) error {
	return c.Collection(req.Database, req.Collection).FindOne(ctx, req.Filter, result)
}

func (c *mongoClient) Find(ctx context.Context, req *FindRequest, results interface{}, opts ...*options.FindOptions) error {
	var opt *options.FindOptions = nil
	if req.Limit > 0 {
		opt = options.Find().SetLimit(req.Limit)
	}
	if req.Skip > 0 {
		if opt == nil {
			opt = options.Find()
		}
		opt.SetSkip(req.Skip)
	}
	if req.Sort != nil {
		if opt == nil {
			opt = options.Find()
		}
		opt.SetSort(req.Sort)
	}

	if opt == nil {
		return c.Collection(req.Database, req.Collection).Find(ctx, req.Filter, results)
	}

	return c.Collection(req.Database, req.Collection).Find(ctx, req.Filter, results, opt)
}

func (c *mongoClient) Exists(ctx context.Context, req *ExistsRequest) (bool, error) {
	return c.Collection(req.Database, req.Collection).Exists(ctx, req.Filter)
}

func (c *mongoClient) Aggregate(ctx context.Context, req *AggregateRequest, results interface{}) error {
	return c.Collection(req.Database, req.Collection).Aggregate(ctx, req.Pipeline, results)
}

func (c *mongoClient) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func NewMongoClient(ctx context.Context, uri, username, password string) MongoClient {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	fmt.Println("Using MONGO_URI: ", uri)
	uriString := fmt.Sprintf(uri, username, password)
	opts := options.Client().ApplyURI(uriString).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic("Failed to connect to MongoDB: " + err.Error())
	}
	err = client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err()
	if err != nil {
		panic("Failed to ping MongoDB: " + err.Error())
	}
	fmt.Println("Connected to MongoDB")
	return &mongoClient{client}
}
