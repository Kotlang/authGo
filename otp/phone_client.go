package otp

import (
	"context"
	"os"
	"strings"
	"unicode"

	"github.com/Kotlang/authGo/db"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/verify/v2"
	"go.uber.org/zap"
)

type PhoneClientInterface interface {
	IsValid(emailOrPhone string) bool
	SendOtp(phoneNumber string) error
	SaveLoginInfo(tenant string, loginInfo *db.LoginModel) *db.LoginModel
	GetLoginInfo(tenant, phone string) *db.LoginModel
	Verify(to, otp string) bool
}

type PhoneClient struct {
	mongo odm.MongoClient
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

func (c *PhoneClient) SaveLoginInfo(tenant string, loginInfo *db.LoginModel) *db.LoginModel {
	userType := strings.TrimSpace(loginInfo.UserType)

	if len(userType) == 0 {
		loginInfo.UserType = "member"
	}

	if loginInfo != nil {
		<-odm.CollectionOf[db.LoginModel](c.mongo, tenant).Save(context.Background(), *loginInfo)
	}

	return loginInfo
}

// get login info using phone number
func (c *PhoneClient) GetLoginInfo(tenant, phone string) *db.LoginModel {
	loginInfo := <-db.FindOneByPhoneOrEmail(context.Background(), c.mongo, tenant, phone, "")
	if loginInfo == nil {
		loginInfo = &db.LoginModel{
			Phone:    phone,
			UserType: "member",
		}
	}
	return loginInfo
}

// sends otp to phone number using twilio.
func (c *PhoneClient) SendOtp(phoneNumber string) error {
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
		return err
	}
	logger.Info("Sending otp status", zap.String("status", *res.Status))
	return nil
}

// verifies otp and returns true if otp is valid.
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
