package api

import (
	"context"
	"fmt"
	"log-service/data"
	"log-service/logs"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	gRpcPort = "50001"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

// WriteLog handles the WriteLog gRPC call
func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	// Start a timer for gRPC request duration
	start := time.Now()

	input := req.GetLogEntry()

	// Log entry creation
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)

	// Record duration for gRPC request
	duration := time.Since(start).Seconds()
	GrpcRequestDuration.WithLabelValues("WriteLog").Observe(duration)

	if err != nil {
		// Log failure and increment error counter for log insertion
		Log.WithFields(logrus.Fields{
			"action": "WriteLog",
			"error":  err,
			"name":   input.Name,
			"data":   input.Data,
		}).Error("Failed to insert log entry")

		LogInsertionErrors.Inc()
		res := &logs.LogResponse{Result: "failed"}
		return res, err
	}

	// Log success and increment success counter for log insertion
	Log.WithFields(logrus.Fields{
		"action": "WriteLog",
		"name":   input.Name,
		"data":   input.Data,
	}).Info("Successfully gRPC logged entry")

	LogInsertionTotal.WithLabelValues("success").Inc()
	// return response
	res := &logs.LogResponse{Result: "logged!"}
	return res, nil
}

// GRPCListen starts the gRPC server and listens for incoming requests
func (app *Config) GRPCListen() {
	// Configure logrus for JSON output
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetOutput(os.Stdout)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		Log.WithError(err).Fatal("Failed to listen for gRPC")
	}

	s := grpc.NewServer()

	logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})

	Log.Infof("gRPC Server started on port %s", gRpcPort)

	if err := s.Serve(lis); err != nil {
		Log.WithError(err).Fatal("Failed to serve gRPC server")
	}
}
