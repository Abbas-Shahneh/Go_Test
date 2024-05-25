package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"portScan/Config"
	"portScan/Controller"
	"portScan/Repository"
	"portScan/Routes"
	"portScan/Service"
)

func main() {
	clientOptions := options.Client().ApplyURI(Config.MongoURI)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err == nil {
		log.Println("Connect to mongodb://localhost:27017")
	} else {
		log.Fatal(err)
	}

	// Check the connection
	dberr := client.Ping(ctx, nil)
	if dberr != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	} else {
		log.Printf("Its OK")
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Printf("Contex TimeOut!!!")
		}
	}(client, ctx)

	repo := Repository.NewScanRepository(client)
	scanRequest := make(chan string, 100)
	service := Service.NewScanService(repo, scanRequest)
	controller := Controller.NewScanController(service)
	router := Routes.SetupRouter(controller)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("Server running at http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
