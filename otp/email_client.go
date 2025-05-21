package otp

import (
	"context"
	"regexp"
	"strings"

	"github.com/Kotlang/authGo/db"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type EmailClientInterface interface {
	IsValid(emailOrPhone string) bool
	SendOtp(emailId string) error
	SaveLoginInfo(tenant string, loginInfo *db.LoginModel) *db.LoginModel
	GetLoginInfo(tenant, email string) *db.LoginModel
	Verify(to, otp string) bool
}

type EmailClient struct {
	mongo odm.MongoClient
}

func (c *EmailClient) IsValid(emailOrPhone string) bool {
	match, _ := regexp.MatchString("^(.+)@(.+)$", emailOrPhone)
	return match
}

func (c *EmailClient) SaveLoginInfo(tenant string, loginInfo *db.LoginModel) *db.LoginModel {
	userType := strings.TrimSpace(loginInfo.UserType)

	if len(userType) == 0 {
		loginInfo.UserType = "member"
	}

	<-odm.CollectionOf[db.LoginModel](c.mongo, tenant).Save(context.Background(), *loginInfo)

	return loginInfo
}

func (c *EmailClient) GetLoginInfo(tenant, email string) *db.LoginModel {
	loginInfo := <-db.FindOneByPhoneOrEmail(context.Background(), c.mongo, tenant, "", email)
	if loginInfo == nil {
		loginInfo = &db.LoginModel{
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
