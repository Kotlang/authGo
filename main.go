package main

import (
	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/SaiNageswarS/go-api-boot/server"
	"github.com/rs/cors"
)

var grpcPort = ":50051"
var webPort = ":8081"

func main() {

	app := InitializeApp()
	app.CloudFns.LoadSecretsIntoEnv()

	corsConfig := cors.New(
		cors.Options{
			AllowedHeaders: []string{"*"},
		})
	bootServer := server.NewGoApiBoot(
		server.WithCorsConfig(corsConfig),
		server.AppendUnaryInterceptors(app.UnaryInterceptors),
		server.AppendStreamInterceptors(app.StreamInterceptors))

	authPb.RegisterLoginServer(bootServer.GrpcServer, app.LoginService)
	authPb.RegisterLoginVerifiedServer(bootServer.GrpcServer, app.LoginVerifiedService)
	authPb.RegisterProfileServer(bootServer.GrpcServer, app.ProfileService)
	authPb.RegisterProfileMasterServer(bootServer.GrpcServer, app.ProfileMasterService)
	authPb.RegisterLeadServiceServer(bootServer.GrpcServer, app.LeadService)

	bootServer.Start(grpcPort, webPort)
}
