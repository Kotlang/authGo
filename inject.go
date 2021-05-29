package main

import (
	"os"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/otp"
	"github.com/Kotlang/authGo/service"
	"github.com/joho/godotenv"
)

type Inject struct {
	AuthDb *db.AuthDb

	TenantDto *db.TenantDto

	EmailClient  *otp.EmailClient
	LoginService *service.LoginService
}

func NewInject() *Inject {
	godotenv.Load()
	mongo_uri := os.Getenv("MONGO_URI")

	authDb := db.NewAuthDb(mongo_uri)
	tenantDto := db.NewTenantDto(authDb)

	emailClient := otp.NewEmailClient()

	loginService := service.NewLoginService(tenantDto, emailClient)

	return &Inject{
		AuthDb:       authDb,
		TenantDto:    tenantDto,
		EmailClient:  emailClient,
		LoginService: loginService,
	}
}
