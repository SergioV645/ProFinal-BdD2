package mongodb

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var Client *mongo.Client
var Database *mongo.Database
var HistorialCollection *mongo.Collection

func Connect() error {
	uri := os.Getenv("MONGODB_URI")

	if uri == "" {
		return fmt.Errorf("no se ha definido la variable MONGODB_URI")
	}

	clientOptions := options.Client().
		ApplyURI(uri).
		SetAppName("AurumMotors")

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return fmt.Errorf("error creando cliente MongoDB: %w", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return fmt.Errorf("no se pudo conectar con MongoDB Atlas: %w", err)
	}

	Client = client
	Database = client.Database("aurum_nosql")
	HistorialCollection = Database.Collection("historial_vehicular")

	return nil
}

func Disconnect() error {
	if Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	return Client.Disconnect(ctx)
}
