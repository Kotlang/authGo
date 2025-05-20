package main

import (
	"context"

	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/Kotlang/authGo/interceptors"
	"github.com/Kotlang/authGo/otp"
	"github.com/Kotlang/authGo/service"
	"github.com/SaiNageswarS/go-api-boot/cloud"
	"github.com/SaiNageswarS/go-api-boot/config"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"github.com/SaiNageswarS/go-api-boot/server"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func main() {
	cloudFns := &cloud.Azure{}

	ccfgg := &config.BootConfig{}
	config.LoadConfig("config.ini", ccfgg)

	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(ccfgg.MongoUri))
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	logger.Info("MongoDB connected")

	otpClient := &otp.DevOtpClient{}

	boot, err := server.New(ccfgg).
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
		Register(server.Adapt(authPb.RegisterLoginServer), service.ProvideLoginService).
		Register(server.Adapt(authPb.RegisterLoginVerifiedServer), service.ProvideLoginVerifiedService).
		Register(server.Adapt(authPb.RegisterProfileServer), service.ProvideProfileService).
		Register(server.Adapt(authPb.RegisterProfileMasterServer), service.ProvideProfileMasterService).
		Register(server.Adapt(authPb.RegisterLeadServiceServer), service.ProvideLeadService).
		Build()

	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	ctx, _ := context.WithCancel(context.Background())
	boot.Serve(ctx)
	logger.Info("Server shutdown cleanly")
}
