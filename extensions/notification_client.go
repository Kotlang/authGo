// Package extensions provides functionality related to the notification service.
// It manages the connection with the notification service and registers events.
package extensions

import (
	"context"
	"errors"
	"os"
	"sync"

	pb "github.com/Kotlang/authGo/generated"
	"github.com/SaiNageswarS/go-api-boot/logger"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Singleton instance of the NotificationClient.
var notification_client *NotificationClient = &NotificationClient{}

// NotificationClient represents a client that manages the connection with the notification service.
type NotificationClient struct {
	// cached_conn caches the active gRPC connection to the notification service.
	cached_conn *grpc.ClientConn

	// conn_creation_lock ensures thread-safe creation of the gRPC connection.
	conn_creation_lock sync.Mutex
}

// getNotificationConnection establishes and returns a gRPC connection with the notification service.
// If a cached connection exists and is ready, it returns the cached connection.
// Otherwise, it creates a new connection based on the NOTIFICATION_TARGET environment variable.
func (c *NotificationClient) getNotificationConnection() *grpc.ClientConn {
	c.conn_creation_lock.Lock()
	defer c.conn_creation_lock.Unlock()

	if c.cached_conn == nil || c.cached_conn.GetState().String() != "READY" {
		if val, ok := os.LookupEnv("NOTIFICATION_TARGET"); ok {
			conn, err := grpc.Dial(val, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				logger.Error("Failed getting connection with notification service", zap.Error(err))
				return nil
			}
			c.cached_conn = conn
		} else {
			logger.Error("Failed to get NOTIFICATION_TARGET env variable")
		}

	}

	return c.cached_conn
}

// RegisterEvent registers an event with the notification service.
// The function communicates asynchronously with the notification service and returns a channel to communicate errors.
// Returns an error channel which sends nil if the event registration is successful, otherwise sends an error.
func RegisterEvent(grpcContext context.Context, event *pb.RegisterEventRequest) chan error {
	errChan := make(chan error)

	go func() {
		// call notification service.
		conn := notification_client.getNotificationConnection()
		if conn == nil {
			errChan <- errors.New("Failed to get connection with notification service")
			return
		}

		client := pb.NewNotificationServiceClient(conn)

		ctx := prepareCallContext(grpcContext)
		if ctx == nil {
			errChan <- errors.New("Failed to get context")
			return
		}

		_, err := client.RegisterEvent(ctx, event)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	return errChan
}

// prepareCallContext prepares the context for making a gRPC call to the notification service.
// It extracts the JWT token from the incoming context and appends it to the outgoing context.
func prepareCallContext(grpcContext context.Context) context.Context {
	jwtToken, err := grpc_auth.AuthFromMD(grpcContext, "bearer")
	if err != nil {
		logger.Error("Failed getting jwt token", zap.Error(err))
		return nil
	}

	return metadata.AppendToOutgoingContext(context.Background(), "Authorization", "bearer "+jwtToken)
}
