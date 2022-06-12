package otp

import (
	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
)

type Channel interface {
	IsValid(to string) bool
	SendOtp(to string)
	GetLoginInfo(tenant, to string) *models.LoginModel
	CreateLoginInfo(tenant, emailOrPhone string)
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

func (c *OtpClient) SendOtp(tenant, to string) {
	for _, channel := range c.channels {
		if channel.IsValid(to) {
			// create user if doesn't exist.
			loginInfo := channel.GetLoginInfo(tenant, to)
			if loginInfo == nil {
				channel.CreateLoginInfo(tenant, to)
			}

			// send otp through the channel.
			channel.SendOtp(to)
		}
	}
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
