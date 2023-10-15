// Package otp provides functionality related to one-time passwords (OTP)
// and authentication mechanisms.
package otp

import (
	"time"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Channel represents an interface for various OTP delivery methods such as email or phone.
type Channel interface {
	// IsValid checks if the provided destination (email or phone) is valid for the current channel.
	IsValid(to string) bool

	// SendOtp sends a one-time password (OTP) to the specified destination.
	SendOtp(to string)

	// GetLoginInfo retrieves the login information for a specific tenant and destination.
	GetLoginInfo(tenant, to string) *models.LoginModel

	// SaveLoginInfo saves the login information for a specific tenant.
	SaveLoginInfo(tenant string, loginInfo *models.LoginModel) *models.LoginModel

	// Verify checks the validity of the OTP for the specified destination.
	Verify(to, otp string) bool
}

// OtpClient represents a client that can handle OTP related operations
// using multiple channels like email and phone.
type OtpClient struct {
	// db represents a connection to the authentication database.
	db *db.AuthDb

	// channels holds a list of different OTP delivery methods.
	channels []Channel
}

// NewOtpClient creates a new OtpClient with the specified database connection.
// It initializes channels for both Email and Phone-based OTP delivery.
func NewOtpClient(db *db.AuthDb) *OtpClient {
	return &OtpClient{
		db:       db,
		channels: []Channel{&EmailClient{Db: db}, &PhoneClient{Db: db}},
	}
}

// SendOtp sends a one-time password (OTP) to the specified destination using a valid channel.
// It checks for OTP send rate limits and ensures OTPs aren't sent more frequently than allowed.
func (c *OtpClient) SendOtp(tenant, to string) error {
	for _, channel := range c.channels {
		if channel.IsValid(to) {
			now := time.Now().Unix()
			// get login info from db or default info.
			loginInfo := channel.GetLoginInfo(tenant, to)
			if loginInfo.CreatedOn != 0 && (now-loginInfo.LastOtpSentTime) < 60 {
				return status.Error(codes.PermissionDenied, "Exceeded threshold of OTPs in a minute.")
			}

			loginInfo.LastOtpSentTime = now

			// send otp through the channel.
			channel.SendOtp(to)
			channel.SaveLoginInfo(tenant, loginInfo)
		}
	}

	return nil
}

// GetLoginInfo retrieves the login information for a specific tenant and destination using a valid channel.
func (c *OtpClient) GetLoginInfo(tenant, to string) *models.LoginModel {
	for _, channel := range c.channels {
		if channel.IsValid(to) {
			return channel.GetLoginInfo(tenant, to)
		}
	}

	return nil
}

// ValidateOtp verifies the provided OTP for the specified destination.
// If the OTP is valid, it updates the authenticated time for the user.
func (c *OtpClient) ValidateOtp(tenant, to, otp string) bool {
	for _, channel := range c.channels {
		if channel.IsValid(to) {
			isValid := channel.Verify(to, otp)
			if isValid {
				loginInfo := channel.GetLoginInfo(tenant, to)
				loginInfo.Otp = otp
				loginInfo.OtpAuthenticatedTime = time.Now().Unix()
				channel.SaveLoginInfo(tenant, loginInfo)
			}
			return isValid
		}
	}
	return false
}
