package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"go.uber.org/zap"
)

type Channel interface {
	IsValid(to string) bool
	SendOtp(to string, otp string)
	GetOrCreateLoginInfo(tenant, to string) *models.LoginModel
	GetLoginInfo(tenant, to string) *models.LoginModel
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

func generateSixDigitsOtp() string {
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		logger.Error("Failed generating otp", zap.Error(err))
		return "121451"
	}
	return fmt.Sprintf("%v", n.Int64())
}

func (c *OtpClient) SendOtp(tenant, to string) {
	otp := generateSixDigitsOtp()

	for _, channel := range c.channels {
		if channel.IsValid(to) {
			// create user if doesn't exist.
			loginInfo := channel.GetOrCreateLoginInfo(tenant, to)
			// send otp through the channel.
			channel.SendOtp(to, otp)

			// legacy code. For phone, twilio verifies otp at its end.
			// save otp for verification.
			loginInfo.Otp = otp
			<-c.db.Login(tenant).Save(loginInfo)
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
