package interceptors

import (
	"context"
	"time"

	"github.com/Kotlang/authGo/db"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServiceCheckUserExistenceInterceptor interface {
	CheckUserExistenceOverride(ctx context.Context) (context.Context, error)
}

// checks if the user exists and updates the last active time of the user
func UserExistsAndUpdateLastActiveUnaryInterceptor(mongo odm.MongoClient) grpc.UnaryServerInterceptor {
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
		login, err := odm.Await(odm.CollectionOf[db.LoginModel](mongo, tenant).FindOneByID(ctx, userId))
		if err != nil {
			logger.Error("User not found", zap.String("userId", userId), zap.Error(err))
		} else {
			if err := checkUserExistenceAndStatus(login); err != nil {
				return nil, err
			}

			login.LastActive = time.Now().Unix()
			_, err = odm.Await(odm.CollectionOf[db.LoginModel](mongo, tenant).Save(ctx, *login))
			if err != nil {
				logger.Error("Error updating last active time", zap.String("userId", userId), zap.Error(err))
			}
		}

		resp, err := handler(ctx, req)
		return resp, err
	}
}

func checkUserExistenceAndStatus(loginInfo *db.LoginModel) error {
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
