package main

import (
	pb "github.com/Kotlang/authGo/generated"
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
	pb.RegisterLoginServer(bootServer.GrpcServer, inject.LoginService)
	pb.RegisterLoginVerifiedServer(bootServer.GrpcServer, inject.LoginVerifiedService)
	pb.RegisterProfileServer(bootServer.GrpcServer, inject.ProfileService)
	pb.RegisterProfileMasterServer(bootServer.GrpcServer, inject.ProfileMasterService)
	pb.RegisterLeadServiceServer(bootServer.GrpcServer, inject.LeadService)

	bootServer.Start(grpcPort, webPort)
}
