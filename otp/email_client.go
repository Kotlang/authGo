package otp

import (
	"regexp"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
)

type EmailClient struct {
	Db *db.AuthDb
}

func (c *EmailClient) IsValid(emailOrPhone string) bool {
	match, _ := regexp.MatchString("^(.+)@(.+)$", emailOrPhone)
	return match
}

func (c *EmailClient) SendOtp(emailId string) {
	// TODO: Use twilio with send-grid to send otp to email.
}

func (c *EmailClient) CreateLoginInfo(tenant, emailId string) {
	loginInfo := &models.LoginModel{
		Email:    emailId,
		UserType: "default",
	}
	<-c.Db.Login(tenant).Save(loginInfo)
}

func (c *EmailClient) GetLoginInfo(tenant, emailId string) *models.LoginModel {
	return <-c.Db.Login(tenant).FindOneByEmail(emailId)
}

func (c *EmailClient) Verify(to, otp string) bool {
	return true
}
