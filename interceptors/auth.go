package interceptors

import (
	"context"
	"time"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func checkUserExistenceAndStatus(profile *models.ProfileModel) error {
	if profile.IsBlocked {
		logger.Error("User is blocked", zap.String("userId", profile.UserId))
		return status.Error(codes.PermissionDenied, "User is blocked")
	}

	if profile.DeletionInfo.MarkedForDeletion {
		logger.Error("User is marked for deletion", zap.String("userId", profile.UserId))
		return status.Error(codes.PermissionDenied, "User is marked for deletion")
	}
	return nil
}

// checks if the user exists and updates the last active time of the user
func UserExistsAndUpdateLastActiveUnaryInterceptor(db db.AuthDbInterface) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userId, tenant := auth.GetUserIdAndTenant(ctx)

		profileChan, errChan := db.Profile(tenant).FindOneById(userId)
		select {
		case profile := <-profileChan:

			if err := checkUserExistenceAndStatus(profile); err != nil {
				return nil, err
			}

			profile.LastActive = time.Now().Unix()
			err := <-db.Profile(tenant).Save(profile)
			if err != nil {
				logger.Error("Error updating last active time", zap.String("userId", userId), zap.Error(err))
			}
		case err := <-errChan:
			logger.Error("User not found", zap.String("userId", userId), zap.Error(err))
			return nil, status.Error(codes.NotFound, "User not found")
		}

		resp, err := handler(ctx, req)
		return resp, err
	}
}

func UserExistsAndUpdateLastActiveStreamInterceptor(db db.AuthDbInterface) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		userId, tenant := auth.GetUserIdAndTenant(stream.Context())
		profileChan, errChan := db.Profile(tenant).FindOneById(userId)
		select {
		case profile := <-profileChan:

			if err := checkUserExistenceAndStatus(profile); err != nil {
				return err
			}

			profile.LastActive = time.Now().Unix()
			err := <-db.Profile(tenant).Save(profile)
			if err != nil {
				logger.Error("Error updating last active time", zap.String("userId", userId), zap.Error(err))
			}
		case err := <-errChan:
			logger.Error("User not found", zap.String("userId", userId), zap.Error(err))
			return status.Error(codes.NotFound, "User not found")
		}

		err := handler(srv, stream)
		return err
	}
}
