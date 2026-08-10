package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDBAdapter struct {
	client *mongo.Client
}

type mongoDatabase struct {
	database *mongo.Database
}

type mongoCollection struct {
	collection *mongo.Collection
}

type mongoSession struct {
	session mongo.Session
}

// NewMongoDatabase returns a new instance of MongoDBAdapter.
func NewMongoDatabase() MongoDBAdapter {
	return &mongoDBAdapter{}
}

// Connect establishes a MongoDB connection.
func (m *mongoDBAdapter) Connect(ctx context.Context, connectionString string, clientOptions *options.ClientOptions) error {
	if clientOptions == nil {
		clientOptions = options.Client().ApplyURI(connectionString)
	} else {
		clientOptions.ApplyURI(connectionString)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	m.client = client
	return nil
}

// Database returns a handle to the specified database.
func (m *mongoDBAdapter) Database(name string, opts ...*options.DatabaseOptions) MongoDatabase {
	return &mongoDatabase{
		database: m.client.Database(name, opts...),
	}
}

// StartSession starts a new MongoDB session.
func (m *mongoDBAdapter) StartSession(opts ...*options.SessionOptions) (MongoSession, error) {
	session, err := m.client.StartSession(opts...)
	if err != nil {
		return nil, err
	}

	return &mongoSession{session: session}, nil
}

// Client returns the underlying MongoDB client.
func (m *mongoDBAdapter) Client() *mongo.Client {
	return m.client
}

// Close terminates the MongoDB connection.
func (m *mongoDBAdapter) Close(ctx context.Context) error {
	if m.client != nil {
		return m.client.Disconnect(ctx)
	}
	return nil
}

// Collection returns a handle to a MongoDB collection.
func (md *mongoDatabase) Collection(name string, opts ...*options.CollectionOptions) MongoCollection {
	return &mongoCollection{
		collection: md.database.Collection(name, opts...),
	}
}

// Name returns the name of the database.
func (md *mongoDatabase) Name() string {
	return md.database.Name()
}

// InsertOne inserts a single document into the collection.
func (mc *mongoCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return mc.collection.InsertOne(ctx, document, opts...)
}

// InsertMany inserts multiple documents into the collection.
func (mc *mongoCollection) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	return mc.collection.InsertMany(ctx, documents, opts...)
}

// FindOne finds a single document in the collection.
func (mc *mongoCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	return mc.collection.FindOne(ctx, filter, opts...)
}

// Find finds multiple documents in the collection.
func (mc *mongoCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return mc.collection.Find(ctx, filter, opts...)
}

// UpdateOne updates a single document in the collection.
func (mc *mongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return mc.collection.UpdateOne(ctx, filter, update, opts...)
}

// UpdateMany updates multiple documents in the collection.
func (mc *mongoCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return mc.collection.UpdateMany(ctx, filter, update, opts...)
}

// DeleteOne deletes a single document from the collection.
func (mc *mongoCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return mc.collection.DeleteOne(ctx, filter, opts...)
}

// DeleteMany deletes multiple documents from the collection.
func (mc *mongoCollection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return mc.collection.DeleteMany(ctx, filter, opts...)
}

// StartTransaction starts a transaction.
func (ms *mongoSession) StartTransaction(opts ...*options.TransactionOptions) error {
	return ms.session.StartTransaction(opts...)
}

// AbortTransaction aborts the transaction.
func (ms *mongoSession) AbortTransaction(ctx context.Context) error {
	return ms.session.AbortTransaction(ctx)
}

// CommitTransaction commits the transaction.
func (ms *mongoSession) CommitTransaction(ctx context.Context) error {
	return ms.session.CommitTransaction(ctx)
}

// EndSession ends the session.
func (ms *mongoSession) EndSession(ctx context.Context) {
	ms.session.EndSession(ctx)
}
