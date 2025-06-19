package mocks

import (
	"strings"
	"unicode"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
)

type PhoneClient struct {
	Db           db.AuthDbInterface
	phone_number []string
}

func isAllDigit(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func (c *PhoneClient) IsValid(emailOrPhone string) bool {
	return len(emailOrPhone) == 10 && isAllDigit(emailOrPhone)
}

func (c *PhoneClient) SendOtp(phoneNumber string) error {
	c.phone_number = append(c.phone_number, phoneNumber)
	return nil
}

func (c *PhoneClient) SaveLoginInfo(tenant string, loginInfo *models.LoginModel) *models.LoginModel {
	userType := strings.TrimSpace(loginInfo.UserType)

	if len(userType) == 0 {
		loginInfo.UserType = "default"
	}
	<-c.Db.Login(tenant).Save(loginInfo)
	return loginInfo
}

func (c *PhoneClient) GetLoginInfo(tenant, phone string) *models.LoginModel {
	loginInfo := <-c.Db.Login(tenant).FindOneByPhone(phone)

	if loginInfo == nil {
		loginInfo = &models.LoginModel{
			Phone:    phone,
			UserType: "default",
		}
	}
	return loginInfo
}

// The mock verify considers only 123456 as valid otp
func (c *PhoneClient) Verify(to, otp string) bool {
	for _, phone := range c.phone_number {
		if to == phone && otp == "123456" {
			return true
		}
	}
	return false
}
