package main

import (
	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/otp"
	"github.com/Kotlang/authGo/service"
)

type Inject struct {
	AuthDb *db.AuthDb

	Otp *otp.OtpClient

	LoginService   *service.LoginService
	ProfileService *service.ProfileService
}

func NewInject() *Inject {
	inj := &Inject{}
	inj.AuthDb = &db.AuthDb{}

	inj.Otp = otp.NewOtpClient(inj.AuthDb)

	inj.LoginService = service.NewLoginService(inj.AuthDb, inj.Otp)
	inj.ProfileService = service.NewProfileService(inj.AuthDb)

	return inj
}
