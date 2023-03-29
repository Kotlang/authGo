package otp

import (
	"regexp"
	"strings"

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

func (c *EmailClient) SaveLoginInfo(tenant, emailId string, LastOtpSentTime int64, userType string) *models.LoginModel {
	userType = strings.TrimSpace(userType)

	if len(userType) == 0 {
		userType = "default"
	}

	loginInfo := &models.LoginModel{
		Email:           emailId,
		UserType:        userType,
		LastOtpSentTime: LastOtpSentTime,
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
