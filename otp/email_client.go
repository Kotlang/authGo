package otp

import (
	"regexp"
	"strings"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
)

type EmailClientInterface interface {
	db.AuthDbInterface
	IsValid(emailOrPhone string) bool
	SendOtp(emailId string) error
	SaveLoginInfo(tenant string, loginInfo *models.LoginModel) *models.LoginModel
	GetLoginInfo(tenant, email string) *models.LoginModel
	Verify(to, otp string) bool
}

type EmailClient struct {
	Db db.AuthDbInterface
}

func (c *EmailClient) IsValid(emailOrPhone string) bool {
	match, _ := regexp.MatchString("^(.+)@(.+)$", emailOrPhone)
	return match
}

func (c *EmailClient) SaveLoginInfo(tenant string, loginInfo *models.LoginModel) *models.LoginModel {
	userType := strings.TrimSpace(loginInfo.UserType)

	if len(userType) == 0 {
		loginInfo.UserType = "member"
	}

	<-c.Db.Login(tenant).Save(loginInfo)

	return loginInfo
}

func (c *EmailClient) GetLoginInfo(tenant, email string) *models.LoginModel {
	loginInfo := <-c.Db.Login(tenant).FindOneByPhoneOrEmail("", email)
	if loginInfo == nil {
		loginInfo = &models.LoginModel{
			Email:    email,
			UserType: "member",
		}
	}
	return loginInfo
}

func (c *EmailClient) SendOtp(emailId string) error {
	// TODO: Use twilio with send-grid to send otp to email.
	return nil
}

func (c *EmailClient) Verify(to, otp string) bool {
	return true
}
