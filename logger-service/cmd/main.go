package main

import (
	"context"
	"fmt"
	"log-service/api"
	"log-service/data"
	"net/http"
	"net/rpc"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// Initialize logger using the global Log instance from api package
	api.InitLogger()  // Initialize the logger once at the start
	logger := api.Log // Use the global logger

	// Log the application start
	logger.WithField("service", "logger-service").Info("Starting logger service")

	// Use the new ConnectToMongo function from the api package
	mongoClient, err := api.ConnectToMongo(mongoURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MongoDB")
	}
	client = mongoClient

	// Create a context in order to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Close MongoDB connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			logger.WithError(err).Error("Error disconnecting from MongoDB")
			panic(err)
		}
		logger.Info("Disconnected from MongoDB")
	}()

	app := api.Config{
		Models: data.New(client),
		Client: client,
		Logger: logger,
	}

	// Expose Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Info("Starting Prometheus metrics server on :8082")
		if err := http.ListenAndServe(":8082", nil); err != nil {
			logger.WithError(err).Fatal("Error starting Prometheus metrics server")
		}
	}()

	// Register the gRPC Server
	err = rpc.Register(&api.RPCServer{Config: &app})
	if err != nil {
		logger.WithError(err).Fatal("Failed to register RPC server")
	}
	go app.RpcListen()
	go app.GRPCListen()

	// Start web server
	logger.WithField("port", webPort).Info("Starting web server")
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		logger.WithError(err).Fatal("Web server encountered a fatal error")
	}
}
