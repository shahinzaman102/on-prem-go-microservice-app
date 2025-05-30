package api

import (
	"context"
	"fmt"
	"log-service/data"
	"net"
	"net/rpc"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	rpcPort = "5001"
)

// RPCServer is the type for our RPC Server. Methods that take this as a receiver are available
// over RPC, as long as they are exported.
type RPCServer struct {
	Config *Config
}

// RPCPayload is the type for data we receive from RPC
type RPCPayload struct {
	Name string
	Data string
}

// LogInfo writes our payload to mongo
func (r *RPCServer) LogInfo(payload RPCPayload, resp *string) error {
	// Use the global logger (api.Log)
	logger := Log.WithFields(logrus.Fields{
		"action": "LogInfo",
		"name":   payload.Name,
	})

	// Log the incoming payload
	logger.Info("Processing log entry via RPC")

	// Insert log entry into MongoDB
	collection := r.Config.Client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})
	if err != nil {
		// Log error if insertion fails
		logger.WithError(err).Error("Error writing to MongoDB")
		return err
	}

	// Log success
	logger.Info("Successfully processed payload via RPC")

	// Respond back to the caller
	*resp = "Processed payload via RPC"

	return nil
}

// RpcListen listens for incoming RPC requests and processes them
func (app *Config) RpcListen() error {
	// Use the global logger (api.Log)
	logger := Log.WithFields(logrus.Fields{
		"action":  "RpcListen",
		"rpcPort": rpcPort,
	})

	// Log that the RPC server is starting
	logger.Info("Starting RPC server")

	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		// Log error if the server fails to start
		logger.WithError(err).Fatal("Failed to start RPC server")
		return err
	}
	defer listen.Close()

	// Start accepting connections
	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			// Log if accepting a connection fails
			logger.WithError(err).Warn("Failed to accept RPC connection")
			continue
		}

		// Serve RPC connection in a separate goroutine
		go rpc.ServeConn(rpcConn)
	}

}
