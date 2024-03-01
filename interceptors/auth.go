package interceptors

import (
	"context"

	"github.com/Kotlang/authGo/db"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func checkUserExistenceAndStatus(userId, tenant string, db db.AuthDbInterface) error {
	profileChan, errChan := db.Profile(tenant).FindOneById(userId)

	select {
	case profile := <-profileChan:
		if profile.IsBlocked {
			logger.Error("User is blocked", zap.String("userId", userId))
			return status.Error(codes.PermissionDenied, "User is blocked")
		}

		if profile.DeletionInfo.MarkedForDeletion {
			logger.Error("User is marked for deletion", zap.String("userId", userId))
			return status.Error(codes.PermissionDenied, "User is marked for deletion")
		}

	case err := <-errChan:
		logger.Error("User not found", zap.String("userId", userId), zap.Error(err))
		return status.Error(codes.NotFound, "User not found")
	}

	return nil
}

func UserExistsUnaryInterceptor(db db.AuthDbInterface) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userId, tenant := auth.GetUserIdAndTenant(ctx)
		if err := checkUserExistenceAndStatus(userId, tenant, db); err != nil {
			return nil, err
		}

		resp, err := handler(ctx, req)
		return resp, err
	}
}

func UserExistsStreamInterceptor(db db.AuthDbInterface) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		userId, tenant := auth.GetUserIdAndTenant(stream.Context())
		if err := checkUserExistenceAndStatus(userId, tenant, db); err != nil {
			return err
		}

		err := handler(srv, stream)
		return err
	}
}
