package otp

import (
	"time"

	"github.com/Kotlang/authGo/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Channel interface {
	IsValid(to string) bool
	SendOtp(to string) error
	GetLoginInfo(tenant, to string) *db.LoginModel
	SaveLoginInfo(tenant string, loginInfo *db.LoginModel) *db.LoginModel
	Verify(to, otp string) bool
}

type OtpClientInterface interface {
	SendOtp(tenant, to string) error
	GetLoginInfo(tenant, to string) *db.LoginModel
	ValidateOtp(tenant, to, otp string) bool
}

type OtpClient struct {
	db       db.AuthDbInterface
	channels []Channel
}

func ProvideOtpClient(db db.AuthDbInterface) OtpClientInterface {
	return &OtpClient{
		db:       db,
		channels: []Channel{&EmailClient{Db: db}, &PhoneClient{Db: db}},
	}
}

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
			err := channel.SendOtp(to)
			if err != nil {
				return status.Error(codes.Internal, "Failed sending otp")
			}

			// if the user is new populate the UserId field to avoid email and phone clients generating two different ids
			if loginInfo.UserId == "" {
				loginInfo.UserId = loginInfo.Id()
			}

			channel.SaveLoginInfo(tenant, loginInfo)
			//If the message is sent succesfully return nil
			return nil
		}
	}
	return status.Error(codes.InvalidArgument, "Incorrect email or phone")
}

func (c *OtpClient) GetLoginInfo(tenant, to string) *db.LoginModel {
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
