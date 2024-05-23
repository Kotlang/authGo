//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/interceptors"
	"github.com/Kotlang/authGo/otp"
	"github.com/Kotlang/authGo/service"
	"github.com/SaiNageswarS/go-api-boot/cloud"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

type App struct {
	AuthDb   db.AuthDbInterface
	CloudFns cloud.Cloud
	Otp      otp.OtpClientInterface

	LoginService         *service.LoginService
	LoginVerifiedService *service.LoginVerifiedService
	ProfileService       *service.ProfileService
	ProfileMasterService *service.ProfileMasterService
	LeadService          *service.LeadService
	UnaryInterceptors    []grpc.UnaryServerInterceptor
	StreamInterceptors   []grpc.StreamServerInterceptor
}

func ProvideApp(authDb db.AuthDbInterface, cloudFns cloud.Cloud, otp otp.OtpClientInterface, loginVerifiedService *service.LoginVerifiedService, profileMasterService *service.ProfileMasterService, leadService *service.LeadService, loginService *service.LoginService, profileService *service.ProfileService) App {
	return App{
		AuthDb:   authDb,
		CloudFns: cloudFns,
		Otp:      otp,

		LoginService:         loginService,
		LoginVerifiedService: loginVerifiedService,
		ProfileService:       profileService,
		ProfileMasterService: profileMasterService,
		LeadService:          leadService,
		UnaryInterceptors:    append([]grpc.UnaryServerInterceptor{}, interceptors.UserExistsAndUpdateLastActiveUnaryInterceptor(authDb)),
		StreamInterceptors:   append([]grpc.StreamServerInterceptor{}, interceptors.UserExistsAndUpdateLastActiveStreamInterceptor(authDb)),
	}
}

func ProvideCloudFns() cloud.Cloud {
	return &cloud.GCP{}
}

func InitializeApp() App {
	wire.Build(ProvideApp, ProvideCloudFns, db.ProvideAuthDb, service.ProvideLeadService, service.ProvideLoginVerifiedService, service.ProvideProfileMasterService, otp.ProvideOtpClient, service.ProvideLoginService, service.ProvideProfileService)
	return App{}
}
