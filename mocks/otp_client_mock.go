package mocks

import (
	"time"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
	"github.com/Kotlang/authGo/otp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Channel interface {
	IsValid(to string) bool
	SendOtp(to string) error
	GetLoginInfo(tenant, to string) *models.LoginModel
	SaveLoginInfo(tenant string, loginInfo *models.LoginModel) *models.LoginModel
	Verify(to, otp string) bool
}

type OtpClientMock struct {
	db       db.AuthDbInterface
	channels []Channel
}

func NewMockOtpClient(db db.AuthDbInterface) otp.OtpClientInterface {
	return &OtpClientMock{
		db:       db,
		channels: []Channel{&EmailClient{Db: db}, &PhoneClient{Db: db}},
	}
}

func (c *OtpClientMock) SendOtp(tenant, to string) error {
	for _, channel := range c.channels {
		if channel.IsValid(to) {
			now := time.Now().Unix()
			//get login info from db or default info.
			loginInfo := channel.GetLoginInfo(tenant, to)
			// if loginInfo.CreatedOn != 0 && (now-loginInfo.LastOtpSentTime) < 60 {
			// 	return status.Error(codes.PermissionDenied, "Exceeded threshold of OTPs in a minute.")
			// }

			loginInfo.LastOtpSentTime = now

			// send otp through the channel.

			err := channel.SendOtp(to)
			if err != nil {
				return err
			}
			channel.SaveLoginInfo(tenant, loginInfo)
			return nil
		}
	}
	return status.Error(codes.InvalidArgument, "Incorrect email or phone")
}

func (c *OtpClientMock) GetLoginInfo(tenant, to string) *models.LoginModel {
	for _, channel := range c.channels {
		if channel.IsValid(to) {
			return channel.GetLoginInfo(tenant, to)
		}
	}

	return nil
}

func (c *OtpClientMock) ValidateOtp(tenant, to, otp string) bool {
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
