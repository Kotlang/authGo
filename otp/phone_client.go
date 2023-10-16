package otp

import (
	"os"
	"strings"
	"unicode"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/verify/v2"
	"go.uber.org/zap"
)

type PhoneClientInterface interface {
	db.AuthDbInterface
	IsValid(emailOrPhone string) bool
	SendOtp(phoneNumber string)
	SaveLoginInfo(tenant string, loginInfo *models.LoginModel) *models.LoginModel
	GetLoginInfo(tenant, phone string) *models.LoginModel
	Verify(to, otp string) bool
}

type PhoneClient struct {
	Db db.AuthDbInterface
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

func (c *PhoneClient) SendOtp(phoneNumber string) {
	accountSid := os.Getenv("TWILIO-ACCOUNT-SID")
	authToken := os.Getenv("TWILIO-AUTH-TOKEN")
	client := twilio.NewRestClient(accountSid, authToken)

	channel := "sms"
	phoneNumber = "+91" + phoneNumber

	res, err := client.VerifyV2.CreateVerification("VAfa78c49eba6901f198481a166a704019", &openapi.CreateVerificationParams{
		Channel: &channel,
		To:      &phoneNumber,
		// CustomCode: &otp,
	})

	if err != nil {
		logger.Error("Failed sending otp", zap.Error(err))
		return
	}
	logger.Info("Sending otp status", zap.String("status", *res.Status))
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

func (c *PhoneClient) Verify(to, otp string) bool {
	accountSid := os.Getenv("TWILIO-ACCOUNT-SID")
	authToken := os.Getenv("TWILIO-AUTH-TOKEN")
	client := twilio.NewRestClient(accountSid, authToken)

	to = "+91" + to
	verificationCheck, err := client.VerifyV2.CreateVerificationCheck("VAfa78c49eba6901f198481a166a704019", &openapi.CreateVerificationCheckParams{
		Code: &otp,
		To:   &to,
	})

	if err != nil {
		logger.Error("Sdk validation failed.", zap.Error(err))
		return false
	}
	return *verificationCheck.Valid
}
