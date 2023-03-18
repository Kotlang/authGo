package otp

import (
	"time"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Channel interface {
	IsValid(to string) bool
	SendOtp(to string)
	GetLoginInfo(tenant, to string) *models.LoginModel
	SaveLoginInfo(tenant, emailOrPhone string, LastOtpSentTime int64) *models.LoginModel
	Verify(to, otp string) bool
}

type OtpClient struct {
	db       *db.AuthDb
	channels []Channel
}

func NewOtpClient(db *db.AuthDb) *OtpClient {
	return &OtpClient{
		db:       db,
		channels: []Channel{&EmailClient{Db: db}, &PhoneClient{Db: db}},
	}
}

func (c *OtpClient) SendOtp(tenant, to string) error {
	for _, channel := range c.channels {
		if channel.IsValid(to) {
			// create user if doesn't exist.
			loginInfo := channel.GetLoginInfo(tenant, to)
			if loginInfo == nil {
				loginInfo = channel.SaveLoginInfo(tenant, to, 0)
			}

			now := time.Now().Unix()
			if (now - loginInfo.LastOtpSentTime) < 60 {
				return status.Error(codes.PermissionDenied, "Exceeded threshold of OTPs in a minute.")
			}

			// send otp through the channel.
			channel.SendOtp(to)
			channel.SaveLoginInfo(tenant, to, now)
		}
	}

	return nil
}

func (c *OtpClient) GetLoginInfo(tenant, to string) *models.LoginModel {
	for _, channel := range c.channels {
		if channel.IsValid(to) {
			return channel.GetLoginInfo(tenant, to)
		}
	}

	return nil
}

func (c *OtpClient) ValidateOtp(tenant, to, otp string) bool {
	for _, channel := range c.channels {
		if channel.IsValid(to) {
			return channel.Verify(to, otp)
		}
	}
	return false
}
