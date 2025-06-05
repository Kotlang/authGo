package main

import (
	"context"

	"github.com/Kotlang/authGo/appconfig"
	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/Kotlang/authGo/interceptors"
	"github.com/Kotlang/authGo/otp"
	"github.com/Kotlang/authGo/service"
	"github.com/SaiNageswarS/go-api-boot/cloud"
	"github.com/SaiNageswarS/go-api-boot/config"
	"github.com/SaiNageswarS/go-api-boot/dotenv"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"github.com/SaiNageswarS/go-api-boot/server"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func main() {
	dotenv.LoadEnv()
	cloudFns := &cloud.Azure{}

	ccfgg := &appconfig.AppConfig{}
	config.LoadConfig("config.ini", ccfgg)

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(ccfgg.MongoURI))
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	logger.Info("MongoDB connected")

	otpClient := &otp.DevOtpClient{}

	boot, err := server.New().
		GRPCPort(":50051").
		HTTPPort(":8080").
		// Dependency injection
		Provide(ccfgg).
		ProvideAs(cloudFns, (*cloud.Cloud)(nil)).
		ProvideAs(mongoClient, (*odm.MongoClient)(nil)).
		ProvideAs(otpClient, (*otp.OtpClientInterface)(nil)).
		// Custom Interceptors
		Unary(interceptors.UserExistsAndUpdateLastActiveUnaryInterceptor(mongoClient)).
		// Register gRPC service impls
		RegisterService(server.Adapt(authPb.RegisterLoginServer), service.ProvideLoginService).
		RegisterService(server.Adapt(authPb.RegisterLoginVerifiedServer), service.ProvideLoginVerifiedService).
		RegisterService(server.Adapt(authPb.RegisterProfileServer), service.ProvideProfileService).
		RegisterService(server.Adapt(authPb.RegisterProfileMasterServer), service.ProvideProfileMasterService).
		RegisterService(server.Adapt(authPb.RegisterLeadServiceServer), service.ProvideLeadService).
		Build()

	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	ctx, _ := context.WithCancel(context.Background())
	boot.Serve(ctx)
	logger.Info("Server shutdown cleanly")
}
