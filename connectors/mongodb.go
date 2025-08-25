package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDBConnector implements DBConnector for MongoDB
type MongoDBConnector struct {
	config *ConnectionConfig
	client *mongo.Client
	db     *mongo.Database
}

// NewMongoDBConnector creates a new MongoDB connector
func NewMongoDBConnector(config *ConnectionConfig) *MongoDBConnector {
	return &MongoDBConnector{
		config: config,
	}
}

// Connect establishes a connection to MongoDB
func (m *MongoDBConnector) Connect(ctx context.Context) error {
	var uri string
	
	// Handle authentication - username and password are optional for MongoDB
	if m.config.Username != "" && m.config.Password != "" {
		// Full authentication with username and password
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			m.config.Username,
			m.config.Password,
			m.config.Host,
			m.config.Port,
			m.config.Database,
		)
	} else if m.config.Username != "" && m.config.Password == "" {
		// Username without password (some MongoDB setups)
		uri = fmt.Sprintf("mongodb://%s@%s:%d/%s",
			m.config.Username,
			m.config.Host,
			m.config.Port,
			m.config.Database,
		)
	} else {
		// No authentication - username and password are optional
		uri = fmt.Sprintf("mongodb://%s:%d/%s",
			m.config.Host,
			m.config.Port,
			m.config.Database,
		)
	}

	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMaxPoolSize(25)
	clientOptions.SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.db = client.Database(m.config.Database)
	return nil
}

// Ping tests the connection to MongoDB
func (m *MongoDBConnector) Ping(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("MongoDB connection not established")
	}
	return m.client.Ping(ctx, readpref.Primary())
}

// Close closes the MongoDB connection
func (m *MongoDBConnector) Close() error {
	if m.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return m.client.Disconnect(ctx)
	}
	return nil
}

// GetType returns the database type
func (m *MongoDBConnector) GetType() string {
	return "mongodb"
}

// Query executes a query (not applicable for MongoDB, returns error)
func (m *MongoDBConnector) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, fmt.Errorf("Query method not applicable for MongoDB, use Execute instead")
}

// Execute runs a MongoDB operation
func (m *MongoDBConnector) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	if m.db == nil {
		return nil, fmt.Errorf("MongoDB connection not established")
	}

	switch operation {
	// Database-level operations (don't require collection parameter)
	case "listCollections":
		filter := params["filter"]
		if filter == nil {
			filter = map[string]interface{}{}
		}
		
		// Check if a specific database is requested
		var targetDB *mongo.Database
		if dbName, ok := params["database"].(string); ok && dbName != "" {
			// Use the specified database
			targetDB = m.client.Database(dbName)
		} else {
			// Use the default connected database
			targetDB = m.db
		}
		
		cursor, err := targetDB.ListCollections(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to list collections: %w", err)
		}
		
		var collections []map[string]interface{}
		if err := cursor.All(ctx, &collections); err != nil {
			return nil, fmt.Errorf("failed to decode collections: %w", err)
		}
		
		return collections, nil
		
	// Collection-level operations (require collection parameter)
	default:
		collection, ok := params["collection"].(string)
		if !ok {
			return nil, fmt.Errorf("collection parameter required for MongoDB collection operations")
		}

		// Check if a specific database is requested
		var targetDB *mongo.Database
		if dbName, ok := params["database"].(string); ok && dbName != "" {
			// Use the specified database
			targetDB = m.client.Database(dbName)
		} else {
			// Use the default connected database
			targetDB = m.db
		}

		coll := targetDB.Collection(collection)
		return m.executeCollectionOperation(ctx, operation, coll, params)
	}
}

// executeCollectionOperation handles collection-specific operations
func (m *MongoDBConnector) executeCollectionOperation(ctx context.Context, operation string, coll *mongo.Collection, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "find":
		filter := params["filter"]
		if filter == nil {
			filter = map[string]interface{}{}
		}
		
		// Build find options
		findOptions := make([]*options.FindOptions, 0)
		
		// Handle limit parameter
		if limit, ok := params["limit"].(int64); ok {
			findOptions = append(findOptions, options.Find().SetLimit(limit))
		} else if limit, ok := params["limit"].(int); ok {
			findOptions = append(findOptions, options.Find().SetLimit(int64(limit)))
		}
		
		// Handle skip parameter
		if skip, ok := params["skip"].(int64); ok {
			findOptions = append(findOptions, options.Find().SetSkip(skip))
		} else if skip, ok := params["skip"].(int); ok {
			findOptions = append(findOptions, options.Find().SetSkip(int64(skip)))
		}
		
		// Handle sort parameter
		if sort, ok := params["sort"].(map[string]interface{}); ok {
			findOptions = append(findOptions, options.Find().SetSort(sort))
		}
		
		cursor, err := coll.Find(ctx, filter, findOptions...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute find: %w", err)
		}
		
		var results []map[string]interface{}
		if err := cursor.All(ctx, &results); err != nil {
			return nil, fmt.Errorf("failed to decode results: %w", err)
		}
		
		return results, nil

	case "findOne":
		filter := params["filter"]
		if filter == nil {
			filter = map[string]interface{}{}
		}
		
		var result map[string]interface{}
		err := coll.FindOne(ctx, filter).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, nil
			}
			return nil, fmt.Errorf("failed to execute findOne: %w", err)
		}
		
		return result, nil

	case "insert":
		document := params["document"]
		if document == nil {
			return nil, fmt.Errorf("document parameter required for insert operation")
		}
		
		result, err := coll.InsertOne(ctx, document)
		if err != nil {
			return nil, fmt.Errorf("failed to insert document: %w", err)
		}
		
		return result, nil

	case "insertMany":
		documents, ok := params["documents"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("documents parameter required for insertMany operation")
		}
		
		result, err := coll.InsertMany(ctx, documents)
		if err != nil {
			return nil, fmt.Errorf("failed to insert documents: %w", err)
		}
		
		return result, nil

	case "update":
		filter := params["filter"]
		update := params["update"]
		if filter == nil || update == nil {
			return nil, fmt.Errorf("filter and update parameters required for update operation")
		}
		
		result, err := coll.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update document: %w", err)
		}
		
		return result, nil

	case "updateMany":
		filter := params["filter"]
		update := params["update"]
		if filter == nil || update == nil {
			return nil, fmt.Errorf("filter and update parameters required for updateMany operation")
		}
		
		result, err := coll.UpdateMany(ctx, filter, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update documents: %w", err)
		}
		
		return result, nil

	case "upsert":
		filter := params["filter"]
		update := params["update"]
		if filter == nil || update == nil {
			return nil, fmt.Errorf("filter and update parameters required for upsert operation")
		}
		
		opts := options.Update().SetUpsert(true)
		result, err := coll.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert document: %w", err)
		}
		
		return result, nil

	case "delete":
		filter := params["filter"]
		if filter == nil {
			return nil, fmt.Errorf("filter parameter required for delete operation")
		}
		
		result, err := coll.DeleteOne(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to delete document: %w", err)
		}
		
		return result, nil

	case "deleteMany":
		filter := params["filter"]
		if filter == nil {
			return nil, fmt.Errorf("filter parameter required for deleteMany operation")
		}
		
		result, err := coll.DeleteMany(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to delete documents: %w", err)
		}
		
		return result, nil

	case "count":
		filter := params["filter"]
		if filter == nil {
			filter = map[string]interface{}{}
		}
		
		count, err := coll.CountDocuments(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to count documents: %w", err)
		}
		
		return count, nil

	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// IsConnected returns whether the connection is active
func (m *MongoDBConnector) IsConnected() bool {
	if m.client == nil {
		return false
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	return m.Ping(ctx) == nil
}
