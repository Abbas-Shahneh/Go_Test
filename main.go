package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"portScan/Config"
	"portScan/Controller"
	"portScan/Repository"
	"portScan/Routes"
	"portScan/Service"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	client, err := connectToMongoDB(Config.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	repo := Repository.NewScanRepository(client)
	scanRequest := make(chan string, 100)
	svc := Service.NewScanService(repo, scanRequest)
	ctrl := Controller.NewScanController(svc)
	router := Routes.SetupRouter(ctrl)

	srv := startServer(":8080", router)

	// Wait for interrupt signal to gracefully shutdown the server and service
	waitForShutdown(srv, svc)
}

func connectToMongoDB(uri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to MongoDB")
	return client, nil
}

func startServer(addr string, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Println("Server running at http://localhost" + addr)
	return srv
}

//Note: you must actually build the program for this to work. If you run the program via go run in a console and send a SIGTERM via ^C,
//the signal is written into the channel and the program responds, but appears to drop out of the loop unexpectedly.
//This is because the SIGRERM goes to go run as well! (This has cause me substantial confusion!) –
//William Pursell
// Nov 17, 2012 at 22:33

func waitForShutdown(srv *http.Server, svc Service.ScanService) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Shutting down server... Received signal: %s", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server gracefully stopped")

	if err := svc.Shutdown(); err != nil {
		log.Fatal("Service forced to shutdown:", err)
	}

	log.Println("Service gracefully stopped")
}
