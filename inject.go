package main

import (
	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/otp"
	"github.com/Kotlang/authGo/service"
	"github.com/joho/godotenv"
)

type Inject struct {
	AuthDb *db.AuthDb

	EmailClient    *otp.EmailClient
	LoginService   *service.LoginService
	ProfileService *service.ProfileService
}

func NewInject() *Inject {
	godotenv.Load()
	inj := &Inject{}
	inj.AuthDb = &db.AuthDb{}

	inj.EmailClient = otp.NewEmailClient(inj.AuthDb)

	inj.LoginService = service.NewLoginService(inj.AuthDb, inj.EmailClient)
	inj.ProfileService = service.NewProfileService(inj.AuthDb)

	return inj
}
