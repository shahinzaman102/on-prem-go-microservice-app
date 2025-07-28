package main

import (
	"context"
	"fmt"
	"log-service/api"
	"log-service/data"
	"log-service/internal/tracing"
	"net/http"
	"net/rpc"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
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
	api.InitLogger()
	logger := api.Log

	ctx := context.Background()

	// Initialize tracing
	shutdown, err := tracing.InitTracer(ctx, "logger-service")
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize tracer")
	}
	defer shutdown(ctx)

	tracer := otel.Tracer("logger-service")
	ctx, span := tracer.Start(ctx, "main")
	defer span.End()

	logger.WithField("service", "logger-service").Info("Starting logger service")

	// Trace MongoDB connection
	ctx, mongoSpan := tracer.Start(ctx, "connect_to_mongodb")
	mongoClient, err := api.ConnectToMongo(mongoURL)
	mongoSpan.End()
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MongoDB")
	}
	client = mongoClient

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

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

	// Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Info("Starting Prometheus metrics server on :8082")
		if err := http.ListenAndServe(":8082", nil); err != nil {
			logger.WithError(err).Fatal("Error starting Prometheus metrics server")
		}
	}()

	// Trace RPC server setup
	ctx, rpcSpan := tracer.Start(ctx, "register_rpc_server")
	err = rpc.Register(&api.RPCServer{Config: &app})
	rpcSpan.End()
	if err != nil {
		logger.WithError(err).Fatal("Failed to register RPC server")
	}

	go app.RpcListen()
	go app.GRPCListen()

	// Trace web server startup
	ctx, srvSpan := tracer.Start(ctx, "start_http_server")
	logger.WithField("port", webPort).Info("Starting web server")
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}
	srvSpan.End()

	if err := srv.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("Web server encountered a fatal error")
	}
}
