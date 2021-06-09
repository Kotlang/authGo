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

	TenantRepo  *db.TenantRepository
	LoginRepo   *db.LoginRepository
	ProfileRepo *db.ProfileRepository

	EmailClient    *otp.EmailClient
	LoginService   *service.LoginService
	ProfileService *service.ProfileService
}

func NewInject() *Inject {
	godotenv.Load()
	inj := &Inject{}

	mongo_uri := os.Getenv("MONGO_URI")
	inj.AuthDb = db.NewAuthDb(mongo_uri)

	inj.TenantRepo = db.NewTenantRepository(inj.AuthDb)
	inj.LoginRepo = db.NewLoginRepository(inj.AuthDb)
	inj.ProfileRepo = db.NewProfileRepository(inj.AuthDb)

	inj.EmailClient = otp.NewEmailClient(inj.LoginRepo)

	inj.LoginService = service.NewLoginService(inj.TenantRepo, inj.ProfileRepo, inj.EmailClient)
	inj.ProfileService = service.NewProfileService(inj.ProfileRepo)

	return inj
}
