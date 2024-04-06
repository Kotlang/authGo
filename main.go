package main

import (
	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/SaiNageswarS/go-api-boot/server"
	"github.com/rs/cors"
)

var grpcPort = ":50051"
var webPort = ":8081"

func main() {

	inject := NewInject()
	inject.CloudFns.LoadSecretsIntoEnv()

	corsConfig := cors.New(
		cors.Options{
			AllowedHeaders: []string{"*"},
		})
	bootServer := server.NewGoApiBoot(corsConfig, inject.UnaryInterceptors, inject.StreamInterceptors)
	authPb.RegisterLoginServer(bootServer.GrpcServer, inject.LoginService)
	authPb.RegisterLoginVerifiedServer(bootServer.GrpcServer, inject.LoginVerifiedService)
	authPb.RegisterProfileServer(bootServer.GrpcServer, inject.ProfileService)
	authPb.RegisterProfileMasterServer(bootServer.GrpcServer, inject.ProfileMasterService)
	authPb.RegisterLeadServiceServer(bootServer.GrpcServer, inject.LeadService)

	bootServer.Start(grpcPort, webPort)
}
