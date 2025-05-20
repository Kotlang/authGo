package otp

import "github.com/Kotlang/authGo/db"

type DevOtpClient struct{}

func (s *DevOtpClient) SendOtp(tenant, to string) error {
	return nil
}

func (s *DevOtpClient) GetLoginInfo(tenant, to string) *db.LoginModel {
	return &db.LoginModel{
		Phone:    to,
		UserType: "member",
	}
}

func (s *DevOtpClient) ValidateOtp(tenant, to, otp string) bool {
	return true
}
