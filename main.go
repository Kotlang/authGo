package main

import (
	"os"

	pb "github.com/Kotlang/authGo/generated"
	"github.com/SaiNageswarS/go-api-boot/server"
	"github.com/rs/cors"
)

var grpcPort = ":50051"
var webPort = ":8081"

// main is the entry point of the program. It loads secrets from Azure Key Vault, initializes the gRPC server, and starts the server on the specified ports.
// This program provides authentication and profile management services through gRPC.
// The gRPC server listens on port 50051 and the web server listens on port 8081.
// If you're not familiar with programming, this program provides a way for applications to securely authenticate users and manage user profiles.
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
