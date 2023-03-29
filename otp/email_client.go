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

func (c *EmailClient) SaveLoginInfo(tenant, emailId string, LastOtpSentTime int64, loginInfo *models.LoginModel) *models.LoginModel {
	if loginInfo == nil {
		loginInfo = &models.LoginModel{
			Email:           emailId,
			UserType:        "default",
			LastOtpSentTime: LastOtpSentTime,
		}
	} else {
		loginInfo.LastOtpSentTime = LastOtpSentTime
	}
	<-c.Db.Login(tenant).Save(loginInfo)

	return loginInfo
}

func (c *EmailClient) GetLoginInfo(tenant, emailId string) *models.LoginModel {
	return <-c.Db.Login(tenant).FindOneByEmail(emailId)
}

func (c *EmailClient) Verify(to, otp string) bool {
	return true
}
