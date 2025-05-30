package integration

import (
	"net"

	"log-service/logs"

	"google.golang.org/grpc"
)

// newServer creates a new gRPC server and registers the log service
func newServer() *grpc.Server {
	server := grpc.NewServer()
	logs.RegisterLogServiceServer(server, &serverImpl{}) // Replace `serverImpl` with your actual service implementation
	return server
}

// getListener returns a listener that listens on the specified address
func getListener() (net.Listener, error) {
	return net.Listen("tcp", ":50051")
}

type serverImpl struct {
	logs.UnimplementedLogServiceServer
}
