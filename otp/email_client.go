// Package otp provides functionality related to one-time passwords (OTP)
// and authentication mechanisms using email.
package otp

import (
	"regexp"
	"strings"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
)

// EmailClient represents a client that can handle OTP related operations
// for email-based authentication.
type EmailClient struct {
	// Db represents a connection to the authentication database.
	Db *db.AuthDb
}

// IsValid checks if the provided input is a valid email address.
// Returns true if the input is a valid email, otherwise false.
func (c *EmailClient) IsValid(emailOrPhone string) bool {
	match, _ := regexp.MatchString("^(.+)@(.+)$", emailOrPhone)
	return match
}

// SendOtp sends a one-time password (OTP) to the specified email address.
// TODO: Implement sending of OTP using Twilio with SendGrid.
func (c *EmailClient) SendOtp(emailId string) {
	// TODO: Use twilio with send-grid to send otp to email.
}

// SaveLoginInfo saves the login information for a specific tenant.
// If the UserType is not specified in the provided LoginModel, it defaults to "default".
// Returns the saved LoginModel.
func (c *EmailClient) SaveLoginInfo(tenant string, loginInfo *models.LoginModel) *models.LoginModel {
	userType := strings.TrimSpace(loginInfo.UserType)

	if len(userType) == 0 {
		loginInfo.UserType = "default"
	}

	<-c.Db.Login(tenant).Save(loginInfo)

	return loginInfo
}

// GetLoginInfo retrieves the login information for a specific tenant and email.
// If no login information is found, it initializes a new LoginModel with UserType set to "default".
// Returns the found or initialized LoginModel.
func (c *EmailClient) GetLoginInfo(tenant, email string) *models.LoginModel {
	loginInfo := <-c.Db.Login(tenant).FindOneByEmail(email)
	if loginInfo == nil {
		loginInfo = &models.LoginModel{
			Email:    email,
			UserType: "default",
		}
	}
	return loginInfo
}

// Verify checks the OTP sent to a specific recipient (email address or phone number).
// Currently, this function always returns true.
// TODO: Implement actual OTP verification.
func (c *EmailClient) Verify(to, otp string) bool {
	return true
}
