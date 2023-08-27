package main

import (
	"os"

	pb "github.com/Kotlang/authGo/generated"
	"github.com/SaiNageswarS/go-api-boot/server"
	"github.com/rs/cors"
)

var grpcPort = ":50051"
var webPort = ":8081"

func main() {
	// go-api-boot picks up keyvault name from environment variable.
	os.Setenv("AZURE-KEYVAULT-NAME", "kotlang-secrets")
	server.LoadSecretsIntoEnv(true)
	inject := NewInject()

	corsConfig := cors.New(
		cors.Options{
			AllowedHeaders: []string{"*"},
		})
	bootServer := server.NewGoApiBoot(corsConfig)
	pb.RegisterLoginServer(bootServer.GrpcServer, inject.LoginService)
	pb.RegisterProfileServer(bootServer.GrpcServer, inject.ProfileService)
	pb.RegisterProfileMasterServer(bootServer.GrpcServer, inject.ProfileMasterService)

	bootServer.Start(grpcPort, webPort)
}
