// Package otp provides functionality related to one-time passwords (OTP)
// for phone-based authentication using Twilio.
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

// PhoneClient represents a client that can handle OTP related operations
// for phone-based authentication using the Twilio API.
type PhoneClient struct {
	// Db represents a connection to the authentication database.
	Db *db.AuthDb
}

// isAllDigit checks if the provided string consists of only digits.
// Returns true if all characters in the string are digits, otherwise false.
func isAllDigit(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsValid checks if the provided input string is a valid phone number format.
// Returns true if the input string is a 10-digit phone number, otherwise false.
func (c *PhoneClient) IsValid(emailOrPhone string) bool {
	return len(emailOrPhone) == 10 && isAllDigit(emailOrPhone)
}

// SendOtp sends a one-time password (OTP) to the specified phone number using the Twilio API.
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

// SaveLoginInfo saves the login information for a specific tenant.
// If the UserType is not specified in the provided LoginModel, it defaults to "default".
// Returns the saved LoginModel.
func (c *PhoneClient) SaveLoginInfo(tenant string, loginInfo *models.LoginModel) *models.LoginModel {
	userType := strings.TrimSpace(loginInfo.UserType)

	if len(userType) == 0 {
		loginInfo.UserType = "default"
	}

	<-c.Db.Login(tenant).Save(loginInfo)

	return loginInfo
}

// GetLoginInfo retrieves the login information for a specific tenant and phone number.
// If no login information is found, it initializes a new LoginModel with UserType set to "default".
// Returns the found or initialized LoginModel.
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

// Verify verifies the provided OTP against the Twilio API for the specified phone number.
// Returns true if the OTP is valid, otherwise false.
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
