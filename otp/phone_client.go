package otp

import (
	"os"
	"unicode"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/verify/v2"
	"go.uber.org/zap"
)

type PhoneClient struct {
	Db *db.AuthDb
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

func (c *PhoneClient) CreateLoginInfo(tenant, phoneNumber string) {
	loginInfo := &models.LoginModel{
		Phone:    phoneNumber,
		UserType: "default",
	}
	<-c.Db.Login(tenant).Save(loginInfo)
}

func (c *PhoneClient) GetLoginInfo(tenant, to string) *models.LoginModel {
	return <-c.Db.Login(tenant).FindOneByPhone(to)
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
