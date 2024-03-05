package interceptors

import (
	"context"
	"time"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServiceCheckUserExistenceInterceptor interface {
	CheckUserExistenceOverride(ctx context.Context) (context.Context, error)
}

// checks if the user exists and updates the last active time of the user
func UserExistsAndUpdateLastActiveUnaryInterceptor(db db.AuthDbInterface) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		// check if the service has overridden the interceptor
		if overrideService, ok := info.Server.(ServiceCheckUserExistenceInterceptor); ok {
			newCtx, err := overrideService.CheckUserExistenceOverride(ctx)
			if err != nil {
				return nil, err
			}
			ctx = newCtx
			return handler(ctx, req)
		}

		userId, tenant := auth.GetUserIdAndTenant(ctx)
		loginResChan, errChan := db.Login(tenant).FindOneById(userId)
		select {
		case login := <-loginResChan:

			if err := checkUserExistenceAndStatus(login); err != nil {
				return nil, err
			}

			login.LastActive = time.Now().Unix()
			err := <-db.Login(tenant).Save(login)
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

		// check if the service has overridden the interceptor
		if overrideService, ok := srv.(ServiceCheckUserExistenceInterceptor); ok {
			newCtx, err := overrideService.CheckUserExistenceOverride(stream.Context())
			if err != nil {
				return err
			}
			wrapped := grpc_middleware.WrapServerStream(stream)
			wrapped.WrappedContext = newCtx
			return handler(srv, wrapped)
		}

		userId, tenant := auth.GetUserIdAndTenant(stream.Context())
		loginResChan, errChan := db.Login(tenant).FindOneById(userId)
		select {
		case login := <-loginResChan:

			if err := checkUserExistenceAndStatus(login); err != nil {
				return err
			}

			login.LastActive = time.Now().Unix()
			err := <-db.Login(tenant).Save(login)
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

func checkUserExistenceAndStatus(loginInfo *models.LoginModel) error {
	if loginInfo.IsBlocked {
		logger.Error("User is blocked", zap.String("userId", loginInfo.UserId))
		return status.Error(codes.PermissionDenied, "User is blocked")
	}

	if loginInfo.DeletionInfo.MarkedForDeletion {
		logger.Error("User is marked for deletion", zap.String("userId", loginInfo.UserId))
		return status.Error(codes.PermissionDenied, "User is marked for deletion")
	}
	return nil
}
