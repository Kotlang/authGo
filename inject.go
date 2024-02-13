package main

import (
	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/otp"
	"github.com/Kotlang/authGo/service"
	"github.com/SaiNageswarS/go-api-boot/cloud"
)

type Inject struct {
	AuthDb   db.AuthDbInterface
	CloudFns cloud.Cloud
	Otp      otp.OtpClientInterface

	LoginService         *service.LoginService
	ProfileService       *service.ProfileService
	ProfileMasterService *service.ProfileMasterService
}

func NewInject() *Inject {
	inj := &Inject{}
	inj.AuthDb = &db.AuthDb{}
	inj.CloudFns = &cloud.GCP{}

	inj.Otp = otp.NewOtpClient(inj.AuthDb)

	inj.LoginService = service.NewLoginService(inj.AuthDb, inj.Otp)
	inj.ProfileService = service.NewProfileService(inj.AuthDb, inj.CloudFns)
	inj.ProfileMasterService = service.NewProfileMasterService(inj.AuthDb)

	return inj
}
