package main

import (
	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/otp"
	"github.com/Kotlang/authGo/service"
)

// Inject is a struct that holds the dependencies required by the application.
// It contains the AuthDb, OtpClient, LoginService, ProfileService and ProfileMasterService.
// AuthDb is used to interact with the database, OtpClient is used to generate and verify OTPs,
// LoginService is used to handle user login, ProfileService is used to handle user profile related operations,
// and ProfileMasterService is used to handle operations related to the master profile.
type Inject struct {
	AuthDb *db.AuthDb

	Otp *otp.OtpClient

	LoginService         *service.LoginService
	ProfileService       *service.ProfileService
	ProfileMasterService *service.ProfileMasterService
}

// NewInject is a function that returns a new instance of Inject with all the dependencies initialized.
func NewInject() *Inject {
	inj := &Inject{}
	inj.AuthDb = &db.AuthDb{}

	inj.Otp = otp.NewOtpClient(inj.AuthDb)

	inj.LoginService = service.NewLoginService(inj.AuthDb, inj.Otp)
	inj.ProfileService = service.NewProfileService(inj.AuthDb)
	inj.ProfileMasterService = service.NewProfileMasterService(inj.AuthDb)

	return inj
}
