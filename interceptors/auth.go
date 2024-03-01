package interceptors

import (
	"context"

	"github.com/Kotlang/authGo/db"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UserExistsInterceptor(db db.AuthDbInterface) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		// check if user exists and is not blocked
		// if not exists or blocked, return error
		userId, tenant := auth.GetUserIdAndTenant(ctx)

		filter := bson.M{"_id": userId, "isBlocked": false}

		// check if user exists
		profileChan, errChan := db.Profile(tenant).FindOne(filter)

		select {
		case profile := <-profileChan:
			if profile.IsBlocked {
				logger.Error("User is blocked", zap.String("userId", userId))
				return nil, status.Error(codes.PermissionDenied, "User is blocked")
			}

			if profile.DeletionInfo.MarkedForDeletion {
				logger.Error("User is marked for deletion", zap.String("userId", userId))
				return nil, status.Error(codes.PermissionDenied, "User is marked for deletion")
			}

		case err := <-errChan:
			logger.Error("User not found", zap.String("userId", userId), zap.Error(err))
			return nil, status.Error(codes.NotFound, "User not found")
		}

		// Call the next handler in the chain
		resp, err := handler(ctx, req)

		return resp, err
	}
}
