package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func DBSet() *mongo.Client {
	client, err := mongo.NewClient(option.Client().ApplyURL("mongodb://localhost:27017"))

	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeOut(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("Failed to connect ot mongodb")
		return nil
	}
	fmt.Println("Successfully conected to mongodb")
	return client
}

var Client *mongo.Client = DBSet()

func UserData(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("Ecoapp").Collection(collectionName)
	return collection

}

func ProductData(client *mongo.Client, collentionName string) *mongo.Collection {
	var productCollection *mongo.Collection = client.Database("Ecoapp").Collection(collectionName)
	return productCollection
}
