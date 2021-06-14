package otp

import (
	"regexp"
	"time"

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

func (c *EmailClient) SendOtp(tenant, emailId string) {
	// TODO: Use twilio with send-grid to send otp to email.
}

func (c *EmailClient) GetOrCreateLoginInfo(tenant, emailId string) *models.LoginModel {
	loginInfo := <-c.Db.Login(tenant).FindOneByEmail(emailId)
	if loginInfo == nil {
		loginInfo = &models.LoginModel{
			Email:     emailId,
			CreatedOn: time.Now().Unix(),
			UserType:  "default",
		}
		<-c.Db.Login(tenant).Save(loginInfo)
	}
	return loginInfo
}

func (c *EmailClient) GetLoginInfo(tenant, emailId string) *models.LoginModel {
	return <-c.Db.Login(tenant).FindOneByEmail(emailId)
}

func (c *EmailClient) Verify(to, otp string) bool {
	return true
}
