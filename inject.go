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

	TenantDto *db.TenantRepository
	LoginDto *db.LoginRepository

	EmailClient  *otp.EmailClient
	LoginService *service.LoginService
}

func NewInject() *Inject {
	godotenv.Load()
	inj := &Inject{}

	mongo_uri := os.Getenv("MONGO_URI")
	inj.AuthDb = db.NewAuthDb(mongo_uri)

	inj.TenantDto = db.NewTenantRepository(inj.AuthDb)
	inj.LoginDto = db.NewLoginRepository(inj.AuthDb)

	inj.EmailClient = otp.NewEmailClient(inj.LoginDto)

	inj.LoginService = service.NewLoginService(inj.TenantDto, inj.EmailClient)

	return inj
}
