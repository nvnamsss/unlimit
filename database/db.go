package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// GormDBAdapter defines the interface for a database adapter using GORM ORM.
// It provides methods for database connection management, transaction control,
// and accessing the underlying GORM and SQL DB instances.
type GormDBAdapter interface {
	// Open establishes a database connection using the provided connection string and configuration.
	// Returns an error if the connection fails.
	Open(connectionString string, config gorm.Config) error

	// Begin starts a new database transaction and returns a new adapter instance
	// bound to the transaction context.
	Begin() GormDBAdapter

	// RollbackUselessCommitted rolls back the current transaction if it hasn't been committed yet.
	// This is typically used for cleanup in defer statements.
	RollbackUselessCommitted()

	// Commit permanently applies the changes made within the current transaction to the database.
	// If the transaction is already committed, this method has no effect.
	Commit()

	// Close terminates the database connection.
	// This should be called when the database connection is no longer needed.
	Close()

	// Gormer returns the underlying gorm.DB instance for advanced GORM operations.
	Gormer() *gorm.DB
}

// MongoDBAdapter defines the interface for a MongoDB adapter.
// It provides methods for database connection management, transaction control,
// and accessing the underlying MongoDB client and database instances.
type MongoDBAdapter interface {
	// Connect establishes a connection to MongoDB using the provided connection string and options.
	// Returns an error if the connection fails.
	Connect(ctx context.Context, connectionString string, clientOptions *options.ClientOptions) error

	// Database returns a handle to the specified database.
	Database(name string, opts ...*options.DatabaseOptions) MongoDatabase

	// StartSession starts a new session for transactions.
	StartSession(opts ...*options.SessionOptions) (MongoSession, error)

	// Client returns the underlying MongoDB client.
	Client() *mongo.Client

	// Close terminates the MongoDB connection.
	// This should be called when the connection is no longer needed.
	Close(ctx context.Context) error
}

// MongoDatabase represents a MongoDB database.
type MongoDatabase interface {
	// Collection returns a handle to a MongoDB collection.
	Collection(name string, opts ...*options.CollectionOptions) MongoCollection

	// Name returns the name of the database.
	Name() string
}

// MongoCollection represents a MongoDB collection.
type MongoCollection interface {
	// InsertOne inserts a single document into the collection.
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)

	// InsertMany inserts multiple documents into the collection.
	InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error)

	// FindOne finds a single document in the collection.
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult

	// Find finds multiple documents in the collection.
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)

	// UpdateOne updates a single document in the collection.
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	// UpdateMany updates multiple documents in the collection.
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	// DeleteOne deletes a single document from the collection.
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)

	// DeleteMany deletes multiple documents from the collection.
	DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

// MongoSession represents a MongoDB session used for transactions.
type MongoSession interface {
	// StartTransaction starts a transaction.
	StartTransaction(opts ...*options.TransactionOptions) error

	// AbortTransaction aborts the transaction.
	AbortTransaction(ctx context.Context) error

	// CommitTransaction commits the transaction.
	CommitTransaction(ctx context.Context) error

	// EndSession ends the session.
	EndSession(ctx context.Context)
}
